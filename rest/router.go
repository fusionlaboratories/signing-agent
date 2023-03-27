package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/hub/message"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/lib/actions"
	"github.com/qredo/signing-agent/lib/clients"
	rest_handlers "github.com/qredo/signing-agent/rest/handlers"
	"github.com/qredo/signing-agent/rest/version"
	"github.com/qredo/signing-agent/util"
)

const (
	PathHealthcheckVersion = "/healthcheck/version"
	PathHealthCheckConfig  = "/healthcheck/config"
	PathHealthCheckStatus  = "/healthcheck/status"
	PathClientFullRegister = "/register"
	PathClient             = "/client"
	PathAction             = "/client/action/{action_id}"
	PathClientFeed         = "/client/feed"
)

type Router struct {
	log                 *zap.SugaredLogger
	config              *config.Config
	router              http.Handler
	actionHandler       *rest_handlers.ActionHandler
	middleware          *Middleware
	version             *version.Version
	signingAgentHandler *rest_handlers.SigningAgentHandler
	healthCheckHandler  *rest_handlers.HealthCheckHandler
	clientsMng          clients.ServiceMng
}

func NewQRouter(log *zap.SugaredLogger, config *config.Config, version *version.Version) (*Router, error) {
	var err error

	log.Infof("Using %s store", config.Store.Type)
	store := util.CreateStore(config)
	if store == nil {
		log.Panicf("unsupported store type: %s", config.Store.Type)
	}

	err = store.Init()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise store")
	}

	core, err := lib.New(config, store)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise core")
	}

	rds := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.LoadBalancing.RedisConfig.Host, config.LoadBalancing.RedisConfig.Port),
		Password: config.LoadBalancing.RedisConfig.Password,
		DB:       config.LoadBalancing.RedisConfig.DB,
	})
	pool := goredis.NewPool(rds)
	rs := redsync.New(pool)

	serverConn := hub.NewWebsocketSource(hub.NewDefaultDialer(), genWSQredoCoreClientFeedURL(&config.Websocket), log, core, &config.Websocket)
	var messageMng message.Cacher
	if !config.AutoApprove.Enabled {
		messageMng = message.NewCacher(config.LoadBalancing.Enable, log, rds)
	}

	feedHub := hub.NewFeedHub(serverConn, log, messageMng)

	localFeed := fmt.Sprintf("ws://%s%s/client/feed", config.HTTP.Addr, defs.PathPrefix)
	syncronizer := actions.NewSyncronizer(&config.LoadBalancing, rds, rs)
	clientsManager := clients.NewManager(core, feedHub, log, config, hub.NewDefaultUpgrader(config.Websocket.ReadBufferSize, config.Websocket.WriteBufferSize), syncronizer)

	signingAgentHandler := rest_handlers.NewSigningAgentHandler(clientsManager, core, log, localFeed)
	healthCheckHandler := rest_handlers.NewHealthCheckHandler(serverConn, version, config, feedHub, localFeed)
	actionHandler := rest_handlers.NewActionHandler(actions.NewActionManager(core, syncronizer, log, config.LoadBalancing.Enable, messageMng))

	rt := &Router{
		log:                 log,
		config:              config,
		middleware:          NewMiddleware(log, config.HTTP.LogAllRequests),
		version:             version,
		signingAgentHandler: signingAgentHandler,
		healthCheckHandler:  healthCheckHandler,
		actionHandler:       actionHandler,
		clientsMng:          clientsManager,
	}

	rt.router = rt.SetHandlers()

	return rt, nil
}

// SetHandlers set all handlers
func (r *Router) SetHandlers() http.Handler {

	routes := []route{
		{PathHealthcheckVersion, http.MethodGet, r.healthCheckHandler.HealthCheckVersion},
		{PathHealthCheckConfig, http.MethodGet, r.healthCheckHandler.HealthCheckConfig},
		{PathHealthCheckStatus, http.MethodGet, r.healthCheckHandler.HealthCheckStatus},
		{PathClientFullRegister, http.MethodPost, r.signingAgentHandler.RegisterAgent},
		{PathClient, http.MethodGet, r.signingAgentHandler.GetClient},
		{PathAction, http.MethodPut, r.actionHandler.ActionApprove},
		{PathAction, http.MethodDelete, r.actionHandler.ActionReject},
		{PathClientFeed, defs.MethodWebsocket, r.signingAgentHandler.ClientFeed},
	}

	router := mux.NewRouter().PathPrefix(defs.PathPrefix).Subrouter()
	for _, route := range routes {

		middle := r.middleware.notProtectedMiddleware

		if route.method == defs.MethodWebsocket {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler)))
		} else {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler))).Methods(route.method)
		}
	}

	router.Use(r.middleware.loggingMiddleware)

	r.printRoutes(router)

	return r.setupCORS(router)
}

// Start starts the service
func (r *Router) Start() error {
	errChan := make(chan error)
	r.StartHTTPListener(errChan)

	return <-errChan
}

// StartHTTPListener starts the HTTP listener
func (r *Router) StartHTTPListener(errChan chan error) {
	r.log.Infof("CORS policy: %s", strings.Join(r.config.HTTP.CORSAllowOrigins, ","))
	r.log.Infof("Starting listener on %v", r.config.HTTP.Addr)

	r.clientsMng.Start()

	if r.config.HTTP.TLS.Enabled {
		r.log.Info("Start listening on HTTPS")
		errChan <- http.ListenAndServeTLS(r.config.HTTP.Addr, r.config.HTTP.TLS.CertFile, r.config.HTTP.TLS.KeyFile, context.ClearHandler(r.router))
	} else {
		r.log.Info("Start listening on HTTP")
		errChan <- http.ListenAndServe(r.config.HTTP.Addr, context.ClearHandler(r.router))
	}
}

// WriteHTTPError writes the error response as JSON
func WriteHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *defs.APIError

	if !errors.As(err, &apiErr) {
		apiErr = defs.ErrInternal().Wrap(err)
	}
	context.Set(r, "error", apiErr)

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(apiErr.Code())
	_, _ = w.Write(apiErr.JSON())
}

// FormatJSONResp encodes response as JSON and handle errors
func FormatJSONResp(w http.ResponseWriter, r *http.Request, v interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		WriteHTTPError(w, r, err)
		return
	}

	if v == nil {
		v = &struct {
			Code int
			Msg  string
		}{
			Code: http.StatusOK,
			Msg:  http.StatusText(http.StatusOK),
		}
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		WriteHTTPError(w, r, err)
		return
	}
}

// Stop shuts down the Signing Agent service
func (r *Router) Stop() {
	r.clientsMng.Stop()
}

func (r *Router) setupCORS(h http.Handler) http.Handler {
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With"}),
		handlers.AllowedOrigins(r.config.HTTP.CORSAllowOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD"}),
		handlers.AllowCredentials(),
	)
	return cors(h)
}

func (r *Router) printRoutes(router *mux.Router) {
	if err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if tpl, err := route.GetPathTemplate(); err == nil {
			if met, err := route.GetMethods(); err == nil {
				for _, m := range met {
					r.log.Debugf("Registered handler %v %v", m, tpl)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

// genWSQredoCoreClientFeedURL assembles and returns the Qredo WS client feed URL as a string.
func genWSQredoCoreClientFeedURL(config *config.WebSocketConfig) string {
	return config.QredoWebsocket
}
