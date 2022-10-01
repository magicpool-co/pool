package blockchair

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

const (
	baseUrl = "https://api.blockchair.com"
)

type Client struct {
	url string
	key string
}

func New(key string) *Client {
	client := &Client{
		url: baseUrl,
		key: key,
	}

	return client
}

func redactAPIKeyFromError(url string, err error) error {
	msg := err.Error()
	for _, slug := range []string{"?key=", "&key="} {
		partial := strings.Split(url, slug)
		if len(partial) > 1 {
			partial = strings.Split(partial[1], "&")
			msg = strings.ReplaceAll(msg, partial[0], "<REDACTED>")
			break
		}
	}

	return fmt.Errorf(msg)
}

func (c *Client) do(method, path string, body, target interface{}) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	url := c.url + path
	if len(c.key) > 0 {
		url += "?key=" + c.key
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	res, err := client.Do(req)
	if err != nil {
		return redactAPIKeyFromError(url, err)
	}

	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

func parseContext(ctx *RawContext) error {
	if ctx.Code != 200 {
		return fmt.Errorf("%d from blockchair: %s", ctx.Code, ctx.Error)
	}

	return nil
}
