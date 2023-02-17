package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/defs"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockActionManager struct {
	ApproveCalled bool
	RejectCalled  bool
	LastActionId  string
	NextError     error
}

func (m *mockActionManager) Approve(actionID string) error {
	m.ApproveCalled = true
	m.LastActionId = actionID
	return m.NextError
}
func (m *mockActionManager) Reject(actionID string) error {
	m.RejectCalled = true
	m.LastActionId = actionID
	return m.NextError
}

func TestActionHandler_ActionApprove_empty_actionId(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("PUT", "/client/action/ ", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.False(t, actionManagerMock.ApproveCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestActionHandler_ActionApprove_error_on_approve(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{
		NextError: errors.New("error while approving"),
	}
	req, _ := http.NewRequest("PUT", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, "error while approving", err.Error())
	assert.True(t, actionManagerMock.ApproveCalled)
	assert.Nil(t, response)
}

func TestActionHandler_ActionApprove(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("PUT", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.NotNil(t, response)
	assert.True(t, actionManagerMock.ApproveCalled)
	assert.Nil(t, err)
	action_response, ok := response.(api.ActionResponse)
	assert.True(t, ok)
	assert.Equal(t, "some_action_id", action_response.ActionID)
	assert.Equal(t, "approved", action_response.Status)
}

func TestActionHandler_ActionReject_empty_actionId(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("DELETE", "/client/action/ ", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.False(t, actionManagerMock.RejectCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	_, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
}

func TestActionHandler_ActionReject_error_on_reject(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{
		NextError: errors.New("error on reject"),
	}
	req, _ := http.NewRequest("DELETE", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.True(t, actionManagerMock.RejectCalled)
	assert.Equal(t, "error on reject", err.Error())
}

func TestActionHandler_ActionReject(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("DELETE", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = NewActionHandler(actionManagerMock).ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.NotNil(t, response)
	assert.True(t, actionManagerMock.RejectCalled)
	assert.Nil(t, err)
	action_response, ok := response.(api.ActionResponse)
	assert.True(t, ok)
	assert.Equal(t, "some_action_id", action_response.ActionID)
	assert.Equal(t, "rejected", action_response.Status)
}
