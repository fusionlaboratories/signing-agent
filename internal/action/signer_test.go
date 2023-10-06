package action

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
	"github.com/test-go/testify/assert"
)

func TestSigner_NewSigner_invalid_key(t *testing.T) {
	//Arrange//Act
	sut, err := NewSigner("", nil, util.NewTestLogger(), "invalid")

	//Assert
	assert.Nil(t, sut)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid bls key", err.Error())
}

func TestSigner_NewSigner_setsKey(t *testing.T) {
	//Arrange
	blsKey := "AAAAAAAAAAAAAAAAAAAAABDDv4z4cTfnlPDDVe/BiMibwqyitjYevyAVXLOf6vOt"

	//Act
	sut, err := NewSigner("", nil, util.NewTestLogger(), blsKey)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, sut)
	assert.NotEmpty(t, sut.(*actionSigner).blsPrivateKey)
}

func TestSigner_SetKey_invalid_key(t *testing.T) {
	//Arrange
	sut, _ := NewSigner("", nil, util.NewTestLogger(), defs.EmptyString)

	//Act
	err := sut.SetKey("new key")

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid bls key", err.Error())
}

func TestSigner_SetKey_sets_key(t *testing.T) {
	//Arrange
	blsKey := "AAAAAAAAAAAAAAAAAAAAABDDv4z4cTfnlPDDVe/BiMibwqyitjYevyAVXLOf6vOt"
	sut, _ := NewSigner("", nil, util.NewTestLogger(), "")

	//Act
	err := sut.SetKey(blsKey)

	//Assert
	assert.Nil(t, err)
	assert.NotEmpty(t, sut.(*actionSigner).blsPrivateKey)
}

func TestSigner_ActionApprove_getActionMessage_req_call_error(t *testing.T) {
	//Arrange
	testHeader := http.Header{}
	testHeader.Set("x-token", "test token")

	authMock := &auth.MockHeaderProvider{
		NextHeader: testHeader,
	}

	var (
		lastMethod      string
		lastURL         string
		lastHeaderValue string
	)

	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		lastMethod = r.Method
		lastURL = r.URL.String()
		lastHeaderValue = r.Header.Get("x-token")

		return nil, errors.New("req error")
	}

	sut := actionSigner{
		htc:          htcMock,
		authProvider: authMock,
		baseURL:      "apiURL",
		log:          util.NewTestLogger(),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "req error")
	assert.Equal(t, http.MethodGet, lastMethod)
	assert.Equal(t, "apiURL/actions/test_id", lastURL)
	assert.Equal(t, "test token", lastHeaderValue)
	assert.True(t, authMock.GetAuthHeaderCalled)
}

func TestSigner_ActionApprove_getActionMessage_status_not_pending(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"messages":["msg1"], "status":3, "id":"someID"}`))),
		}, nil
	}

	sut := actionSigner{
		htc:          htcMock,
		authProvider: authMock,
		baseURL:      "apiURL",
		log:          util.NewTestLogger(),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "action can't be signed, status not pending", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSigner_ActionApprove_getActionMessage_cant_decode_message(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"messages":["msg1"], "status":1, "id":"someID"}`))),
		}, nil
	}

	sut := actionSigner{
		htc:          htcMock,
		authProvider: authMock,
		baseURL:      "apiURL",
		log:          util.NewTestLogger(),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to decode the action message", detail)
	assert.Equal(t, http.StatusInternalServerError, code)
}

func TestSigner_ActionApprove_signAction_invalidKey(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	messages := fmt.Sprintf(`{"messages":["%s"], "status":1, "id":"someID"}`, hex.EncodeToString([]byte("some data")))
	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(messages))),
		}, nil
	}

	sut := actionSigner{
		htc:          htcMock,
		authProvider: authMock,
		baseURL:      "apiURL",
		log:          util.NewTestLogger(),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to generate signature, invalid blsKey", detail)
	assert.Equal(t, http.StatusInternalServerError, code)
}

func TestSigner_ActionApprove_signAction_request_error(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	messages := fmt.Sprintf(`{"messages":["%s"], "status":1, "id":"someID"}`, hex.EncodeToString([]byte("some data")))
	htcMock := util.NewHTTPMockClient()

	var (
		lastMethod      string
		lastURL         string
		lastSignRequest signRequest
	)

	count := 0
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		if count == 0 {
			count++
			return &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(messages))),
			}, nil
		}

		jd := json.NewDecoder(r.Body)

		defer r.Body.Close()
		jd.Decode(&lastSignRequest)

		lastMethod = r.Method
		lastURL = r.URL.String()
		return nil, errors.New("some req error")
	}

	sut := actionSigner{
		htc:           htcMock,
		authProvider:  authMock,
		baseURL:       "apiURL",
		log:           util.NewTestLogger(),
		blsPrivateKey: []byte("data"),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to sign the action", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.Equal(t, http.MethodPost, lastMethod)
	assert.Equal(t, "apiURL/actions/test_id", lastURL)

	assert.Equal(t, approve, lastSignRequest.Status)
	assert.NotEmpty(t, lastSignRequest.Signatures)
}

func TestSigner_ActionApprove_signAction_request_success(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	messages := fmt.Sprintf(`{"messages":["%s"], "status":1, "id":"someID"}`, hex.EncodeToString([]byte("some data")))
	htcMock := util.NewHTTPMockClient()

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(messages))),
		}, nil

	}

	sut := actionSigner{
		htc:           htcMock,
		authProvider:  authMock,
		baseURL:       "apiURL",
		log:           util.NewTestLogger(),
		blsPrivateKey: []byte("data"),
	}

	//Act
	err := sut.ActionApprove("test_id")

	//Assert
	assert.Nil(t, err)
}

func TestSigner_ActionReject_signAction_request_success(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	messages := fmt.Sprintf(`{"messages":["%s"], "status":1, "id":"someID"}`, hex.EncodeToString([]byte("some data")))
	htcMock := util.NewHTTPMockClient()

	var (
		lastMethod      string
		lastURL         string
		lastSignRequest signRequest
	)

	count := 0
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		if count == 0 {
			count++
			return &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(messages))),
			}, nil
		}

		jd := json.NewDecoder(r.Body)

		defer r.Body.Close()
		jd.Decode(&lastSignRequest)

		lastMethod = r.Method
		lastURL = r.URL.String()
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(messages))),
		}, nil
	}

	sut := actionSigner{
		htc:           htcMock,
		authProvider:  authMock,
		baseURL:       "apiURL",
		log:           util.NewTestLogger(),
		blsPrivateKey: []byte("data"),
	}

	//Act
	err := sut.ActionReject("test_id")

	//Assert
	assert.Nil(t, err)

	assert.Equal(t, http.MethodPost, lastMethod)
	assert.Equal(t, "apiURL/actions/test_id", lastURL)

	assert.Equal(t, reject, lastSignRequest.Status)
	assert.NotEmpty(t, lastSignRequest.Signatures)
}

func TestSigner_ApproveActionMessage_signAction_request_success(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextHeader: http.Header{},
	}

	htcMock := util.NewHTTPMockClient()
	var (
		lastMethod      string
		lastURL         string
		lastSignRequest signRequest
	)

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		jd := json.NewDecoder(r.Body)

		defer r.Body.Close()
		jd.Decode(&lastSignRequest)

		lastMethod = r.Method
		lastURL = r.URL.String()
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	}

	sut := actionSigner{
		htc:           htcMock,
		authProvider:  authMock,
		baseURL:       "apiURL",
		log:           util.NewTestLogger(),
		blsPrivateKey: []byte("data"),
	}

	//Act
	err := sut.ApproveActionMessage("test_id", []byte("some data"))

	//Assert
	assert.Nil(t, err)

	assert.Equal(t, http.MethodPost, lastMethod)
	assert.Equal(t, "apiURL/actions/test_id", lastURL)

	assert.Equal(t, approve, lastSignRequest.Status)
	assert.NotEmpty(t, lastSignRequest.Signatures)
}
