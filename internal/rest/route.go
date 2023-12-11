package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/internal/defs"
)

func WrapPathPrefix(uri string) string {
	return strings.Join([]string{defs.PathPrefix, uri}, "")
}

type appHandlerFunc func(ctx *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error)

func (a appHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := &defs.RequestContext{}

	resp, err := a(ctx, w, r)

	if strings.ToLower(r.Header.Get("connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("upgrade")) == "websocket" {
		if err != nil {
			var apiErr *defs.APIError

			if !errors.As(err, &apiErr) {
				apiErr = defs.ErrInternal().Wrap(err)
			}

			context.Set(r, "error", apiErr)
		}
		return
	}

	formatJSONResp(w, r, resp, err)
}

// formatJSONResp encodes response as JSON and handle errors
func formatJSONResp(w http.ResponseWriter, r *http.Request, v interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		writeHTTPError(w, r, err)
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
		writeHTTPError(w, r, err)
		return
	}
}

// writeHTTPError writes the error response as JSON
func writeHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *defs.APIError

	if !errors.As(err, &apiErr) {
		apiErr = defs.ErrInternal().Wrap(err)
	}
	context.Set(r, "error", apiErr)

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(apiErr.Code())
	_, _ = w.Write(apiErr.JSON())
}
