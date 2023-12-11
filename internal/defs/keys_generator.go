package defs

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/qredo/signing-agent/crypto"
)

func HmacSum(timestamp, method, url string, secret, body []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(timestamp)); err != nil {
		return nil, err
	}

	if _, err := mac.Write([]byte(method)); err != nil {
		return nil, err
	}

	if _, err := mac.Write([]byte(url)); err != nil {
		return nil, err
	}

	if _, err := mac.Write(body); err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func GenerateKeys() (string, string, string, string, error) {
	seed, err := randomBytes(48)
	if err != nil {
		return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to generate seed, err: %v", err)
	}

	blsPublic, blsPriv, err := crypto.BLSKeys(crypto.NewRand(seed), nil)
	if err != nil {
		return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to generate bls key, err: %v", err)
	}

	hashedSeed := sha256.Sum256(seed)
	ecPriv, ecPub := btcec.PrivKeyFromBytes(btcec.S256(), hashedSeed[:])
	ecPubS := ecPub.SerializeUncompressed()
	ecPrivS := ecPriv.Serialize()

	return encode(blsPublic), encode(blsPriv), encode(ecPubS), encode(ecPrivS), nil
}

func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)

	return b, err
}

func encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
