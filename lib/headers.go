package lib

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Request struct {
	Uri       string
	Body      []byte
	ApiKey    string
	Timestamp string
	Signature string
	RsaKey    *rsa.PrivateKey
}

func GenTimestamp(req *Request) {
	req.Timestamp = fmt.Sprintf("%v", time.Now().Unix())
}

func GetClientInitHttpHeaders(req *Request) http.Header {
	headers := http.Header{}
	headers.Add("x-api-key", req.ApiKey)
	headers.Add("x-sign", req.Signature)
	headers.Add("x-timestamp", req.Timestamp)
	return headers
}

func SignRequest(req *Request) error {
	h := sha256.New()
	h.Write([]byte(req.Timestamp))
	h.Write([]byte(req.Uri))
	h.Write(req.Body)
	dgst := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(nil, req.RsaKey, crypto.SHA256, dgst)
	if err != nil {
		return errors.Wrap(err, "cannot sign request")
	}

	req.Signature = base64.RawURLEncoding.EncodeToString(signature)
	return nil
}
