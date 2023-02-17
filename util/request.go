package util

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/qredo/signing-agent/defs"
)

func DecodeRequest(req interface{}, hr *http.Request) error {
	switch hr.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if err := DecodeJSON(hr, req); err != nil {
			if err != io.EOF {
				return defs.ErrBadRequest().WithDetail("invalid json").Wrap(err)
			}
		}
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return defs.ErrBadRequest().WithDetail(err.Error())
	}

	return nil
}

// DecodeJSON decodes request JSON body
// returns ServerError on failure
func DecodeJSON(r *http.Request, v interface{}) error {
	jd := json.NewDecoder(r.Body)

	defer r.Body.Close()
	if err := jd.Decode(v); err != nil {
		return err
	}

	return nil
}
