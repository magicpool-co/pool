package mexc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
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
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

func New(apiKey, secretKey string) *Client {
	const (
		mainnetURL = "https://www.mexc.com"
	)

	client := &Client{
		url:        mainnetURL,
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

func (c *Client) do(method, path string, payload map[string]string, target interface{}, authNeeded bool) error {
	var query url.Values
	var body []byte
	var err error
	switch method {
	case "GET":
		query = url.Values{}
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

	parsedURL, err := url.Parse(fullUrl)
	if err != nil {
		return err
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	if authNeeded {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		var sigData bytes.Buffer
		sigData.WriteString(parsedURL.RequestURI())
		sigData.WriteString(timestamp) // @TODO: actually is apart of the query string
		sig := crypto.HmacSha256(c.secretKey, sigData.String())

		headers.Set("api_key", c.apiKey)
		headers.Set("req_time", timestamp)
		headers.Set("sign", hex.EncodeToString(sig))
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
	} else if response.Code != 200 {
		switch response.Code {
		case 429:
			return ErrTooManyRequests
		default:
			return fmt.Errorf("failed executing request: %s: %d: %s", fullUrl, response.Code, response.Message)
		}
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
