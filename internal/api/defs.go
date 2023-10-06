package api

type AgentRegisterRequest struct {
	APIKeyID    string `json:"APIKeyID" validate:"required"`
	Secret      string `json:"secret" validate:"required"`
	WorkspaceID string `json:"workspaceID" validate:"required"`
}

type AgentRegisterResponse struct {
	GetAgentDetailsResponse
}

type GetAgentDetailsResponse struct {
	Name    string `json:"name"`
	AgentID string `json:"agentID"`
	FeedURL string `json:"feedURL"`
}

type ActionResponse struct {
	ActionID string `json:"actionID"`
	Status   string `json:"status"`
}

type WebsocketStatus struct {
	ReadyState       string `json:"readyState"`
	RemoteFeedUrl    string `json:"remoteFeedURL"`
	ConnectedClients uint32 `json:"connectedClients"`
}

type HealthCheckStatusResponse struct {
	WebsocketStatus WebsocketStatus `json:"websocket"`
	LocalFeedUrl    string          `json:"localFeedURL"`
}

type Version struct {
	BuildVersion string `json:"buildVersion"`
	BuildType    string `json:"buildType"`
	BuildDate    string `json:"buildDate"`
}
