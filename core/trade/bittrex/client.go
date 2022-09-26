package bittrex

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/crypto"
)

var (
	ErrEmptyTarget = fmt.Errorf("nil target after marshalling")
)

type Client struct {
	url        string
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

func New(apiKey, secretKey string) *Client {
	client := &Client{
		url:        "https://api.bittrex.com/v3/",
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

func (c *Client) do(method, path string, payload map[string]interface{}, target interface{}, authNeeded bool) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.url+path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}

	if authNeeded {
		payloadSum := sha512.Sum512(body)
		payloadHash := hex.EncodeToString(payloadSum[:])

		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		var sigData bytes.Buffer
		sigData.WriteString(timestamp)
		sigData.WriteString(req.URL.String())
		sigData.WriteString(method)
		sigData.WriteString(payloadHash)
		sig := crypto.HmacSha512(c.secretKey, sigData.String())

		req.Header.Add("Api-Key", c.apiKey)
		req.Header.Add("Api-Timestamp", timestamp)
		req.Header.Add("Api-Content-Hash", payloadHash)
		req.Header.Add("Api-Signature", hex.EncodeToString(sig))
	}

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
		return ErrEmptyTarget
	}

	return nil
}
