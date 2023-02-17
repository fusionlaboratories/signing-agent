package api

// swagger:ignore
type CoreClientServiceActionMessagesResponse struct {
	Messages []string `json:"messages"`
}

// swagger:ignore
type CoreClientServiceActionApproveRequest struct {
	Signatures []string `json:"signatures"`
	ClientID   string   `json:"client_id,omitempty"`
	ActionID   string   `json:"action_id,omitempty"`
}

// swagger:model ActionResponse
type ActionResponse struct {
	// The ID of the transaction
	// example: 2IXwq4klvWbnPf1YaAc1XD85jJX
	ActionID string `json:"actionID"`

	// The status of the transaction
	// enum: approved,rejected
	Status string `json:"status"`
}

func NewApprovedActionResponse(action_id string) ActionResponse {
	return ActionResponse{
		ActionID: action_id,
		Status:   "approved",
	}
}

func NewRejectedActionResponse(action_id string) ActionResponse {
	return ActionResponse{
		ActionID: action_id,
		Status:   "rejected",
	}
}
