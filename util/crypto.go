package util

import (
	"crypto/rand"
	"encoding/json"
	"io"

	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/crypto"
)

const (
	AMCLRandomSeedSize = 48
)

func RandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	n, err := io.ReadAtLeast(rand.Reader, b, size)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, errors.Errorf("read %v random bytes, expected %v", n, size)
	}

	return b, err
}

// CreateAMCLRng creates a new AMCL RNG with a random seed
func CreateAMCLRng() (*crypto.Rand, error) {
	b, err := RandomBytes(AMCLRandomSeedSize)
	if err != nil {
		return nil, err
	}

	return crypto.NewRand(b), nil
}

func BLSSign(seed, payload []byte) ([]byte, error) {

	_, blsSecret, err := crypto.BLSKeys(crypto.NewRand(seed), nil)
	if err != nil {
		return nil, errors.Wrap(err, "generate BLS key")
	}

	signature, err := crypto.BLSSign(payload, blsSecret)
	if err != nil {
		return nil, errors.Wrap(err, "BLS sign payload")
	}

	return signature, nil
}

func BLSVerify(seed, msg, sig []byte) error {

	blsPublic, _, err := crypto.BLSKeys(crypto.NewRand(seed), nil)
	if err != nil {
		return errors.New("generate BLS key")
	}

	return crypto.BLSVerify(msg, blsPublic, sig)
}

func ZKPOnePass(zkpID, zkpToken []byte, pin int) ([]byte, error) {

	rng, err := CreateAMCLRng()
	if err != nil {
		return nil, err
	}
	zkpOnePass, err := crypto.ClientOnePass(zkpID, pin, rng, zkpToken, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Generate ZKP One Pass")
	}

	return json.Marshal(zkpOnePass)
}
