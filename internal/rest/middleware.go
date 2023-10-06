package rest

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/internal/defs"
)

func NewMiddleware(log *zap.SugaredLogger, logAllRequests bool) *Middleware {
	l := log.Desugar()
	ll := l.WithOptions(zap.AddCallerSkip(1)).Sugar()
	mw := &Middleware{
		log:            ll,
		logAllRequests: logAllRequests,
	}
	return mw
}

type Middleware struct {
	log            *zap.SugaredLogger
	logAllRequests bool
}

func (m *Middleware) sessionMiddleware(next appHandlerFunc) appHandlerFunc {

	return func(ctx *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {

		ctx.TraceID = uuid.New().String()
		context.Set(r, "ctx", *ctx)

		return next(ctx, w, r)
	}
}

func (m *Middleware) notProtectedMiddleware(next appHandlerFunc) appHandlerFunc {
	return func(ctx *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		return next(ctx, w, r)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	hijacked   bool
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	if lrw.hijacked {
		return
	}
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := lrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	lrw.hijacked = true
	return h.Hijack()
}

func (m *Middleware) loggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		lw := &loggingResponseWriter{
			w,
			false,
			http.StatusOK,
		}

		next.ServeHTTP(lw, r)

		var traceID string

		ctxI := context.Get(r, "ctx")
		if ctxI != nil {
			if ctx, ok := ctxI.(defs.RequestContext); ok {
				traceID = ctx.TraceID
			}
		}

		errI := context.Get(r, "error")
		if errI != nil {
			if err, ok := errI.(error); ok {
				m.log.Infow("REQ",
					"trace_id", traceID,
					"error", err.Error())
			} else {
				msg := fmt.Sprintf("REQ: %v: Unknown error type: %#v", traceID, errI)
				m.log.Error(msg)
			}
		} else {
			// do not log requestID if no error
			traceID = ""
		}

		// TODO: Make requests method logging configurable
		if m.logAllRequests || r.Method != http.MethodGet || errI != nil {
			m.log.Infof("REQ %s %v %v %v - [%v]", traceID, lw.statusCode, r.Method, r.RequestURI, time.Since(startTime))
		}
	})
}
