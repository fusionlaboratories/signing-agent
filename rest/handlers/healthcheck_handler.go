package handlers

import (
	"net/http"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/rest/version"
)

type HealthCheckHandler struct {
	version      *version.Version
	config       *config.Config
	source       hub.SourceStats
	connClients  hub.ConnectedClients
	localFeedUrl string
}

func NewHealthCheckHandler(source hub.SourceStats, version *version.Version, config *config.Config, connClients hub.ConnectedClients, localFeed string) *HealthCheckHandler {
	return &HealthCheckHandler{
		source:       source,
		version:      version,
		config:       config,
		connClients:  connClients,
		localFeedUrl: localFeed,
	}
}

// HealthCheckVersion
//
// swagger:route GET /healthcheck/version healthcheck HealthCheckVersion
//
// # Check application version
//
// This endpoint returns information about the application version.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: VersionResponse
func (h *HealthCheckHandler) HealthCheckVersion(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return h.version, nil
}

// HealthCheckConfig
//
// swagger:route GET /healthcheck/config healthcheck HealthcheckConfig
//
// # Check application configuration
//
// This endpoint returns the application configuration.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ConfigResponse
func (h *HealthCheckHandler) HealthCheckConfig(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return h.config, nil
}

// HealthCheckStatus
//
// swagger:route GET /healthcheck/status healthcheck HealthcheckStatus
//
// # Check application status
//
// This endpoint returns the application status.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: StatusResponse
func (h *HealthCheckHandler) HealthCheckStatus(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	readyState := h.source.GetReadyState()
	sourceFeedUrl := h.source.GetFeedUrl()
	connectedFeedClients := h.connClients.GetExternalFeedClients()

	response := api.HealthCheckStatusResponse{
		WebsocketStatus: api.NewWebsocketStatus(readyState, sourceFeedUrl, h.localFeedUrl, connectedFeedClients),
	}

	return response, nil
}
