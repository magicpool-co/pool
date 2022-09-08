package binance

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/pkg/crypto"
)

type securityType int

const (
	securityTypeNone securityType = iota
	securityTypeAPIKey
	securityTypeSigned
)

type Client struct {
	url        string
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

func New(apiKey, secretKey string) *Client {
	const (
		mainnetURL = "https://api.binance.com"
		testnetURL = "https://testnet.binance.vision"
	)

	client := &Client{
		url:        testnetURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{},
	}

	return client
}

func (c *Client) doTimeoutRequest(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*5)
	defer cancel()

	return c.httpClient.Do(req.WithContext(ctx))
}

func (c *Client) do(method, path string, payload map[string]string, target interface{}, security securityType) error {
	var query, body url.Values
	switch method {
	case "GET":
		for k, v := range payload {
			query.Set(k, v)
		}
	case "POST":
		for k, v := range payload {
			body.Set(k, v)
		}
	default:
		return fmt.Errorf("unknown http method")
	}

	if security == securityTypeSigned {
		query.Set("recvWindow", strconv.FormatInt(5000, 10))
		// @TODO: maybe use the /api/v3/time to get server time
		query.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	}

	headers := http.Header{}
	queryString := query.Encode()
	bodyString := body.Encode()

	bodyBytes := new(bytes.Buffer)
	if bodyString != "" {
		bodyBytes = bytes.NewBufferString(bodyString)
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if security == securityTypeAPIKey || security == securityTypeSigned {
		headers.Add("X-MBX-APIKEY", c.apiKey)
	}

	if security == securityTypeSigned {
		sig := crypto.HmacSha256(c.secretKey, queryString+bodyString)
		query.Set("signature", hex.EncodeToString(sig))
		queryString = query.Encode()
	}

	fullUrl := c.url + path
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	req, err := http.NewRequest(method, fullUrl, bodyBytes)
	if err != nil {
		return err
	}
	req.Header = headers

	res, err := c.doTimeoutRequest(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	} else if res.StatusCode != 200 && res.StatusCode != 201 {
		return fmt.Errorf("status: %v message:%s", res.Status, string(data))
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return err
	} else if target == nil {
		return fmt.Errorf("nil target after marshalling")
	}

	return nil
}
