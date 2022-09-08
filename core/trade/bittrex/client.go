package bittrex

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*5)
	defer cancel()

	return c.httpClient.Do(req.WithContext(ctx))
}

func (c *Client) do(method, path, payload string, authNeeded bool) ([]byte, error) {
	req, err := http.NewRequest(method, c.url+path, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}

	if authNeeded {
		payloadSum := sha512.Sum512([]byte(payload))
		payloadHash := hex.EncodeToString(payloadSum[:])

		timestamp := time.Now().Unix() * 1000

		preSignature := []string{strconv.Itoa(int(timestamp)), req.URL.String(), method, payloadHash}
		signaturePayload := strings.Join(preSignature, "")

		mac := hmac.New(sha512.New, []byte(c.secretKey))
		_, err = mac.Write([]byte(signaturePayload))
		sig := hex.EncodeToString(mac.Sum(nil))

		req.Header.Add("Api-Key", c.apiKey)
		req.Header.Add("Api-Timestamp", fmt.Sprintf("%d", timestamp))
		req.Header.Add("Api-Content-Hash", payloadHash)
		req.Header.Add("Api-Signature", sig)
	}

	res, err := c.doTimeoutRequest(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	} else if res.StatusCode != 200 && res.StatusCode != 201 {
		return nil, fmt.Errorf("status: %v message:%s", res.Status, string(data))
	}

	return data, nil
}
