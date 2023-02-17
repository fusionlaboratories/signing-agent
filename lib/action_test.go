package lib

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/util"
)

const (
	fixturePathAgent                   = "../testdata/lib/agent.json"
	fixturePathActionApproveGetMessage = "../testdata/lib/actionApproveGetMessage.json"
)

func popMockHttpResponse(alist []*http.Response) *http.Response {
	f := len(alist)
	rv := (alist)[f-1]
	return rv
}

func TestAction(t *testing.T) {
	var (
		cfg *config.Config
		err error
	)
	cfg = &config.Config{
		Base: config.Base{
			PIN:      1234,
			QredoAPI: "https://play-api.qredo.network/api/v1/p",
		},
		AutoApprove: config.AutoApprove{
			Enabled: true,
		},
	}

	kv := util.NewFileStore(TestDataDBStoreFilePath)
	err = kv.Init()
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	core, err := NewMock(cfg, kv)
	assert.NoError(t, err)
	var (
		accountCode = "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
		actionID    = "2D7YA7Ojo3zGRtHP9bw37wF5jq3"
		agent       = &Agent{}
	)
	data, err := os.ReadFile(fixturePathAgent)
	assert.NoError(t, err)
	err = json.Unmarshal(data, agent)
	assert.NoError(t, err)
	_ = core.store.AddAgent(accountCode, agent)
	_ = core.store.SetSystemAgentID(accountCode)

	t.Run(
		"ActionApprove",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			dataFromFixture, err := os.Open(fixturePathActionApproveGetMessage)
			assert.NoError(t, err)
			body := io.NopCloser(dataFromFixture)

			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)

			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				// Every time we call mocked server we get response from the stack via popMockHttpResponse func.
				// First we are calling ep to get action message then sign it and PUT to accept ep.
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(actionID)
			assert.NoError(t, err)
		})

	t.Run(
		"ActionApprove - without first message from Qredo",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			msg := []byte(`{"messages":[]}`)
			body := io.NopCloser(bytes.NewReader(msg))
			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(actionID)
			assert.Error(t, err)
		})

	t.Run(
		"ActionApprove - with first message from Qredo that is wrong",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			msg := []byte(`{"messages":["wrong message that is not a hex"]}`)
			body := io.NopCloser(bytes.NewReader(msg))
			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(actionID)
			assert.Error(t, err)
		})

	t.Run(
		"ActionReject",
		func(t *testing.T) {
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			}
			err = core.ActionReject(actionID)
			assert.NoError(t, err)
		})

}
