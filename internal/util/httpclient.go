package util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
}

func NewHTTPClient() *Client {
	client := &http.Client{
		Timeout: time.Second * 100000,
	}
	return &Client{httpClient: client}
}

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoMockHTTPClientFunc fetches the mock client's `Do` func
	GetDoMockHTTPClientFunc func(req *http.Request) (*http.Response, error)
)

// Do is the mock client's `Do` func
func (mc *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoMockHTTPClientFunc(req)
}

func NewHTTPMockClient() *Client {
	return &Client{httpClient: &MockHTTPClient{}}
}

func (c *Client) Request(method string, url string, reqData interface{}, respData interface{}, headers http.Header) error {
	var body io.Reader
	if reqData != nil {
		jd, err := json.Marshal(reqData)
		if err != nil {
			return errors.Wrap(err, "marshal request as JSON")
		}
		body = bytes.NewBuffer(jd)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	if headers != nil {
		req.Header = headers
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		if b, err := io.ReadAll(resp.Body); err == nil && len(b) > 0 {
			return errors.Errorf("%v %v Status %v (%v) with body: %s", method, url, resp.StatusCode, resp.Status, b)
		}
		return errors.Errorf("%v %v Status %v (%v)", method, url, resp.StatusCode, resp.Status)
	}

	switch respData := respData.(type) {
	case nil:
		return nil
	case *[]byte:
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		if err := json.Unmarshal(b, respData); err != nil {
			return errors.Wrap(err, "decode response as JSON")
		}
	}

	return nil
}

func (c *Client) RequestNoLog(method string, url string, reqData interface{}, respData interface{}, headers http.Header) error {
	var body io.Reader
	if reqData != nil {
		jd, err := json.Marshal(reqData)
		if err != nil {
			return errors.Wrap(err, "marshal request as JSON")
		}
		body = bytes.NewBuffer(jd)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	if headers != nil {
		req.Header = headers
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		if b, err := io.ReadAll(resp.Body); err == nil && len(b) > 0 {
			return errors.Errorf("%v %v Status %v (%v) with body: %s", method, url, resp.StatusCode, resp.Status, b)
		}
		return errors.Errorf("%v %v Status %v (%v)", method, url, resp.StatusCode, resp.Status)
	}

	switch respData := respData.(type) {
	case nil:
		return nil
	case *[]byte:
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		if err := json.Unmarshal(b, respData); err != nil {
			return errors.Wrap(err, "decode response as JSON")
		}
	}

	return nil
}
