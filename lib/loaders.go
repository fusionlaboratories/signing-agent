package lib

import (
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"

	"github.com/pkg/errors"
)

func DecodeBase64RSAKey(req *Request, base64PrivateKey string) error {
	pemData, err := b64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return errors.Wrap(err, "decoding base64 RSA key")
	}

	block, _ := pem.Decode(pemData)

	req.RsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "parse RSA key")
	}
	return nil
}

func EncodeBase64RSAKey(privateKeyByte []byte) string {
	return b64.StdEncoding.EncodeToString(privateKeyByte)
}
