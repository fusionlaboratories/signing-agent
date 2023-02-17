package lib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func generatePrivateKeyBase64() string {
	privatekey, _ := rsa.GenerateKey(rand.Reader, 2048)
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyByte := pem.EncodeToMemory(privateKeyBlock)
	return EncodeBase64RSAKey(privateKeyByte)
}
