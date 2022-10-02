package rpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
)

type Request struct {
	HostID  string            `json:"-"`
	ID      json.RawMessage   `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Worker  string            `json:"worker,omitempty"`
	Params  []json.RawMessage `json:"params,omitempty"`
}

func NewRequest(method string, inputs ...interface{}) (*Request, error) {
	return NewRequestWithID(0, method, inputs...)
}

func MustNewRequest(method string, inputs ...interface{}) *Request {
	req, err := NewRequest(method, inputs...)
	if err != nil {
		panic(err)
	}

	return req
}

func NewRequestWithHostID(hostID, method string, inputs ...interface{}) (*Request, error) {
	req, err := NewRequestWithID(0, method, inputs...)
	if err != nil {
		return nil, err
	}
	req.HostID = hostID

	return req, nil
}

func NewRequestWithID(rawID interface{}, method string, inputs ...interface{}) (*Request, error) {
	if rawID == nil {
		rawID = 0
	}

	id, err := json.Marshal(rawID)
	if err != nil {
		return nil, err
	}

	params := make([]json.RawMessage, len(inputs))
	for i, input := range inputs {
		params[i], err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	}

	req := &Request{
		ID:      id,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return req, nil
}

type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (err *Error) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("rpc error %d", err.Code)
	}

	return err.Message
}

type Response struct {
	HostID  string          `json:"-"`
	ID      json.RawMessage `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error,omitempty"`
}

type ResponseForced struct {
	Response
	Error *Error `json:"error"`
}

func NewResponse(id json.RawMessage, result interface{}) (*Response, error) {
	if id == nil {
		id = common.JsonZero
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	res := &Response{
		ID:      id,
		JSONRPC: "2.0",
		Result:  data,
	}

	return res, nil
}

func NewResponseForced(id json.RawMessage, result interface{}) (*ResponseForced, error) {
	res, err := NewResponse(id, result)
	if err != nil {
		return nil, err
	}

	return &ResponseForced{*res, res.Error}, nil
}

func NewResponseFromJSON(id, result json.RawMessage) *Response {
	res := &Response{
		ID:      id,
		JSONRPC: "2.0",
		Result:  result,
	}

	return res
}

func NewResponseForcedFromJSON(id, result json.RawMessage) *ResponseForced {
	res := NewResponseFromJSON(id, result)

	return &ResponseForced{*res, res.Error}
}

func NewResponseWithError(id json.RawMessage, code int, message string) *Response {
	res := &Response{
		ID:      id,
		JSONRPC: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	return res
}

type Message struct {
	ID      json.RawMessage   `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Worker  string            `json:"worker,omitempty"`
	Params  []json.RawMessage `json:"params"`
	Result  json.RawMessage   `json:"result"`
	Error   *Error            `json:"error,omitempty"`
}

func (m *Message) ToRequest() *Request {
	req := &Request{
		ID:      m.ID,
		JSONRPC: m.JSONRPC,
		Method:  m.Method,
		Params:  m.Params,
	}

	return req
}

func (m *Message) ToResponse() *Response {
	res := &Response{
		ID:      m.ID,
		JSONRPC: m.JSONRPC,
		Result:  m.Result,
		Error:   m.Error,
	}

	return res
}

func ExecRPC(url string, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpRes, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer httpRes.Body.Close()
	res := new(Response)
	err = json.NewDecoder(httpRes.Body).Decode(&res)
	if err != nil {
		return nil, err
	} else if res.Error != nil {
		return nil, fmt.Errorf("RPC Error %d: %s: %s", res.Error.Code, res.Error.Message, res.Error.Data)
	}

	return res, nil
}

func ExecRPCBulk(url string, requests []*Request) ([]*Response, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpRes, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer httpRes.Body.Close()
	data, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		return nil, err
	}

	responses := make([]*Response, 0)
	err = json.Unmarshal(data, &responses)
	if err != nil {
		return nil, fmt.Errorf("failed: %v: %s: %s", err, body, data)
	}
	/*err = json.NewDecoder(httpRes.Body).Decode(&responses)
	if err != nil {
		return nil, err
	}*/

	for _, res := range responses {
		if res.Error != nil {
			return nil, fmt.Errorf("RPC Error %d: %s: %s", res.Error.Code, res.Error.Message, res.Error.Data)
		}
	}

	return responses, nil
}
