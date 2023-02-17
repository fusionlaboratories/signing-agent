package lib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/util"
)

// mock the serve function
func mocServe(fakeError error) ServeCB {
	return func(url string, handler Handler, errHandler ErrHandler) (doneCH, stopCH chan struct{}, err error) {
		data := []byte(`{
		"id":"9b1deb4d-3b7d",
		"coreClientID":"cc229a7e3dd3",
		"type":"Action",
		"status":"OK",
		"timestamp":1665492366,
		"expireTime":1675492366
	}`)
		doneCH = make(chan struct{})
		stopCH = make(chan struct{})
		go func() {
			<-stopCH
			close(doneCH)
		}()
		handler(data)
		if fakeError != nil {
			errHandler(fakeError)
		}
		return doneCH, stopCH, nil
	}
}

func TestFeedRead(t *testing.T) {
	cfg := config.Config{Base: config.Base{QredoAPI: "mock_url"}}
	agent, err := New(&cfg, util.NewFileStore("mock.db"))
	assert.NoError(t, err)
	fakeError := errors.New("fake error")
	doneCH, stopCH, err := NewFeed("mock_url", agent, mocServe(fakeError)).ActionEvent(
		func(e *WsActionInfoEvent) {
			assert.Equal(t, "9b1deb4d-3b7d", e.ID)
			assert.Equal(t, "cc229a7e3dd3", e.AgentID)
			assert.Equal(t, "Action", e.Type)
			assert.Equal(t, "OK", e.Status)
			assert.Equal(t, int64(1665492366), e.Timestamp)
			assert.Equal(t, int64(1675492366), e.ExpireTime)
		},
		func(err error) {
			assert.Equal(t, err, fakeError)
		},
	)
	assert.NoError(t, err)
	stopCH <- struct{}{}
	<-doneCH
}
