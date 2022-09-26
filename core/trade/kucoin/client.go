package kucoin

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/crypto"
)

var (
	ErrEmptyTarget = fmt.Errorf("nil target after marshalling")
)

type Client struct {
	url                 string
	apiKey              string
	secretKey           string
	secretPasphrase     string
	encryptedPassphrase string
	httpClient          *http.Client
}

func New(apiKey, secretKey, secretPasphrase string) *Client {
	const (
		mainnetURL = "https://api.kucoin.com/api/v1/status"
		testnetURL = "https://openapi-sandbox.kucoin.com"
	)

	sig := crypto.HmacSha256(secretKey, secretPasphrase)
	client := &Client{
		url:                 mainnetURL,
		apiKey:              apiKey,
		secretKey:           secretKey,
		secretPasphrase:     secretPasphrase,
		encryptedPassphrase: base64.StdEncoding.EncodeToString(sig),
		httpClient:          &http.Client{},
	}

	return client
}

func (c *Client) doTimeoutRequest(req *http.Request) (*http.Response, error) {
	ctx, _ := context.WithTimeout(req.Context(), time.Second*5)

	return c.httpClient.Do(req.WithContext(ctx))
}

func (c *Client) do(method, path string, payload map[string]string, target interface{}, authNeeded bool) error {
	var query url.Values
	var body []byte
	var err error
	switch method {
	case "GET":
		for k, v := range payload {
			query.Set(k, v)
		}
	case "POST":
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown http method")
	}

	queryString := query.Encode()
	fullUrl := c.url + path
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	if authNeeded {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		var sigData bytes.Buffer
		sigData.WriteString(timestamp)
		sigData.WriteString(method)
		sigData.WriteString(fullUrl)
		sigData.Write(body)
		sig := crypto.HmacSha256(c.secretKey, sigData.String())

		headers.Set("KC-API-KEY", c.apiKey)
		headers.Set("KC-API-PASSPHRASE", c.encryptedPassphrase)
		headers.Set("KC-API-TIMESTAMP", timestamp)
		headers.Set("KC-API-SIGN", string(sig))
		headers.Set("KC-API-KEY-VERSION", "2")
	}

	req, err := http.NewRequest(method, fullUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header = headers

	res, err := c.doTimeoutRequest(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	var response *Response
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return err
	} else if response.Code != "200000" {
		return fmt.Errorf("failed executing request: %s: %s", response.Code, response.Message)
	} else if res.StatusCode != 200 && res.StatusCode != 201 {
		return fmt.Errorf("status: %v message:%s", res.Status, string(response.Data))
	}

	err = json.Unmarshal(response.Data, target)
	if err != nil {
		return err
	} else if target == nil {
		return ErrEmptyTarget
	}

	return nil
}
