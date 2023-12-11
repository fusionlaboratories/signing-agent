package rest

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/service"
	"github.com/qredo/signing-agent/internal/util"
)

const (
	PathHealthcheckVersion = "/healthcheck/version"
	PathHealthCheckConfig  = "/healthcheck/config"
	PathHealthCheckStatus  = "/healthcheck/status"
	PathClientFullRegister = "/register"
	PathClient             = "/client"
	PathAction             = "/client/action/{action_id}"
	PathClientFeed         = "/client/feed"
	PathApprove            = "/approve"
	PathGetToken           = "/token"
	PathRefreshToken       = "/refresh"
)

type route struct {
	path    string
	method  string
	handler appHandlerFunc
}

type Router struct {
	log       *zap.SugaredLogger
	config    config.Config
	handler   http.Handler
	subRouter *mux.Router

	middleware *Middleware
	version    api.Version

	agentService  service.AgentService
	actionService service.ActionService

	decode func(interface{}, *http.Request) error
}

func NewRouter(log *zap.SugaredLogger, config config.Config, version api.Version, service service.AgentService, actionService service.ActionService) *Router {
	app := &Router{
		log:           log,
		middleware:    NewMiddleware(log, config.HTTP.LogAllRequests),
		subRouter:     mux.NewRouter().PathPrefix(defs.PathPrefix).Subrouter(),
		version:       version,
		config:        config,
		agentService:  service,
		actionService: actionService,
		decode:        util.DecodeRequest,
	}

	app.setRoutes()
	return app
}

// setRoutes set all handlers
func (a *Router) setRoutes() {
	routes := []route{
		{PathHealthcheckVersion, http.MethodGet, a.HealthCheckVersion},
		{PathHealthCheckConfig, http.MethodGet, a.HealthCheckConfig},
		{PathHealthCheckStatus, http.MethodGet, a.HealthCheckStatus},
		{PathClientFullRegister, http.MethodPost, a.RegisterAgent},
		{PathClient, http.MethodGet, a.GetClient},
		{PathAction, http.MethodPut, a.ActionApprove},
		{PathAction, http.MethodDelete, a.ActionReject},
		{PathClientFeed, defs.MethodWebsocket, a.ClientFeed},
	}

	for _, route := range routes {
		middle := a.middleware.notProtectedMiddleware
		if route.method == defs.MethodWebsocket {
			a.subRouter.Handle(route.path, a.middleware.sessionMiddleware(middle(route.handler)))
		} else {
			a.subRouter.Handle(route.path, a.middleware.sessionMiddleware(middle(route.handler))).Methods(route.method)
		}
	}

	a.subRouter.Use(a.middleware.loggingMiddleware)

	a.printRoutes()
	a.setupCORS()
}

// Start starts the service
func (a *Router) Start() error {
	errChan := make(chan error)
	a.StartHTTPListener(errChan)

	return <-errChan
}

// StartHTTPListener starts the HTTP listener
func (a *Router) StartHTTPListener(errChan chan error) {
	a.log.Infof("CORS policy: %s", strings.Join(a.config.HTTP.CORSAllowOrigins, ","))
	a.log.Infof("Starting listener on %v", a.config.HTTP.Addr)

	if err := a.agentService.Start(); err != nil {
		a.log.Errorf("Failed to start the agent service, stopping ...")
		os.Exit(1)
	}

	if a.config.HTTP.TLS.Enabled {
		a.log.Info("Start listening on HTTPS")
		errChan <- http.ListenAndServeTLS(a.config.HTTP.Addr, a.config.HTTP.TLS.CertFile, a.config.HTTP.TLS.KeyFile, context.ClearHandler(a.handler))
	} else {
		a.log.Info("Start listening on HTTP")
		errChan <- http.ListenAndServe(a.config.HTTP.Addr, context.ClearHandler(a.handler))
	}
}

// Stop shuts down the Signing Agent service
func (a *Router) Stop() {
	a.agentService.Stop()
}

func (a *Router) setupCORS() {
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"X-Requested-With"}),
		handlers.AllowedOrigins(a.config.HTTP.CORSAllowOrigins),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodHead}),
		handlers.AllowCredentials(),
	)

	a.handler = cors(a.subRouter)
}

func (a Router) printRoutes() {
	if err := a.subRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if tpl, err := route.GetPathTemplate(); err == nil {
			if met, err := route.GetMethods(); err == nil {
				for _, m := range met {
					a.log.Debugf("Registered handler %v %v", m, tpl)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
