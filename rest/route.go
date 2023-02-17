package rest

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/defs"
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

	FormatJSONResp(w, r, resp, err)
}

type route struct {
	path    string
	method  string
	handler appHandlerFunc
}
