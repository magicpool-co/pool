package mexc

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/crypto"
)

var (
	ErrEmptyTarget     = fmt.Errorf("nil target after marshalling")
	ErrTooManyRequests = fmt.Errorf("too many requests")
)

type Client struct {
	url        string
	legacyURL  string
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

func New(apiKey, secretKey string) *Client {
	const (
		mainnetURL = "https://api.mexc.com"
		legacyURL  = "https://www.mexc.com"
	)

	client := &Client{
		url:        mainnetURL,
		legacyURL:  legacyURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{},
	}

	return client
}

func (c *Client) doTimeoutRequest(req *http.Request) (*http.Response, error) {
	ctx, _ := context.WithTimeout(req.Context(), time.Second*5)

	return c.httpClient.Do(req.WithContext(ctx))
}

func (c *Client) do(method, path string, payload map[string]string, target interface{}, legacy, authNeeded bool) error {
	switch method {
	case "GET", "POST", "DELETE":
	default:
		return fmt.Errorf("unknown http method")
	}

	query := url.Values{}
	for k, v := range payload {
		query.Set(k, v)
	}

	baseUrl := c.url
	if legacy {
		baseUrl = c.legacyURL
	}

	if authNeeded {
		query.Set("recvWindow", "5000")
		query.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
		query.Set("signature", hex.EncodeToString(crypto.HmacSha256(c.secretKey, query.Encode())))
	}

	queryString := query.Encode()
	fullUrl := baseUrl + path
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	req, err := http.NewRequest(method, fullUrl, nil)
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"X-MEXC-APIKEY": []string{c.apiKey},
		"Content-Type":  []string{"application/json"},
	}

	res, err := c.doTimeoutRequest(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response *Response
	err = json.Unmarshal(data, &response)
	if legacy && err != nil {
		return err
	} else if err == nil && response != nil && response.Code > 200 {
		switch response.Code {
		case 429:
			return ErrTooManyRequests
		default:
			return fmt.Errorf("failed executing request: %s: %d: %s", fullUrl, response.Code, response.Message)
		}
	} else if res.StatusCode != 200 && res.StatusCode != 201 {
		return fmt.Errorf("status: %v message:%s", res.Status, string(response.Data))
	}

	if legacy {
		err = json.Unmarshal(response.Data, target)
	} else {
		err = json.Unmarshal(data, target)
	}
	if err != nil {
		return err
	} else if target == nil {
		return ErrEmptyTarget
	}

	return nil
}
