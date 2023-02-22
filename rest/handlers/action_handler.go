package handlers

import (
	"net/http"
	"strings"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/autoapprover"
	"github.com/qredo/signing-agent/defs"

	"github.com/gorilla/mux"
)

type ActionHandler struct {
	actionManager autoapprover.ActionManager
}

func NewActionHandler(actionManager autoapprover.ActionManager) *ActionHandler {
	return &ActionHandler{
		actionManager: actionManager,
	}
}

// ActionApprove
//
// swagger:route PUT /client/action/{action_id} action ActionApprove
//
// # Approve a transaction
//
// This endpoint approves a transaction based on the transaction ID, `action_id`, passed.
//
//	Parameters:
//	  + name: action_id
//	    in: path
//	    description: The ID of the transaction that is received from the feed.
//	    required: true
//	    type: string
//
// Produces:
//   - application/json
//
// Responses:
//
// 200: ActionResponse
// 400: ErrorResponse description:Bad request
// 404: ErrorResponse description:Not found
// 500: ErrorResponse description:Internal error
func (h *ActionHandler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}

	if err := h.actionManager.Approve(actionID); err != nil {
		return nil, err
	}

	return api.NewApprovedActionResponse(actionID), nil
}

// ActionReject
//
// swagger:route DELETE /client/action/{action_id} action actionReject
//
// # Reject a transaction
//
// This endpoint rejects a transaction based on the transaction ID, `action_id`, passed.
//
//	Parameters:
//	  + name: action_id
//	    in: path
//	    description: The ID of the transaction that is received from the feed.
//	    required: true
//	    type: string
//
// Produces:
//   - application/json
//
// Responses:
//
// 200: ActionResponse
// 400: ErrorResponse description:Bad request
// 404: ErrorResponse description:Not found
// 500: ErrorResponse description:Internal error
func (h *ActionHandler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}

	if err := h.actionManager.Reject(actionID); err != nil {
		return nil, err
	}

	return api.NewRejectedActionResponse(actionID), nil
}
