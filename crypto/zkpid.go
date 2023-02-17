package crypto

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrInvalidID is returned when the ID cannot be decoded
	ErrInvalidID = errors.New("invalid ID")
)

// ID is the ZKP ID struct
type ID struct {
	Identity  string `json:"id"`
	Curve     string `json:"curve"`
	CreatedAt int64  `json:"created"`
	rawID     []byte
}

// NewID creates a new identity
func NewID(id string) (*ID, error) {
	i := &ID{
		Identity:  id,
		Curve:     curveName,
		CreatedAt: time.Now().Unix(),
	}

	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	i.rawID = b
	return i, nil
}

// IDFromHex decodes a hex-encodded ID
func IDFromHex(rawIDString string) (*ID, error) {
	rawB, err := hex.DecodeString(rawIDString)
	if err != nil {
		return nil, ErrInvalidID
	}
	i, err := IDFromBytes(rawB)
	if err != nil {
		return nil, ErrInvalidID
	}
	return i, nil
}

// IDFromBytes decodes a hex-encodded ID
func IDFromBytes(rawID []byte) (*ID, error) {
	i := &ID{}
	if err := json.Unmarshal(rawID, i); err != nil {
		return nil, ErrInvalidID
	}

	i.rawID = rawID
	return i, nil
}

// Bytes returns the raw ID byteslice
func (i *ID) Bytes() []byte {
	return i.rawID
}

// String returns hex-encoded raw ID
func (i *ID) String() string {
	return hex.EncodeToString(i.rawID)
}

// Hash returns the hash of the raw ID
func (i *ID) Hash() []byte {
	return hashID(i.rawID)
}
