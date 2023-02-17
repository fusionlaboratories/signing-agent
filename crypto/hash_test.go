package crypto

// Forked from btc suite, will redesign

import (
	"bytes"
	"testing"
)

// TestHash integrationtests the Hash API.
func TestHash(t *testing.T) {
	// Hash of block 234439.
	blockHashStr := "14a0810ac680a3eb3f82edc878cea25ec41d6b790744e5daeef"
	blockHash, err := NewHashFromStr(blockHashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Hash of block 234440 as byte slice.
	buf := []byte{
		0x79, 0xa6, 0x1a, 0xdb, 0xc6, 0xe5, 0xa2, 0xe1,
		0x39, 0xd2, 0x71, 0x3a, 0x54, 0x6e, 0xc7, 0xc8,
		0x75, 0x63, 0x2e, 0x75, 0xf1, 0xdf, 0x9c, 0x3f,
		0xa6, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	hash, err := NewHash(buf)
	if err != nil {
		t.Errorf("NewHash: unexpected error %v", err)
	}

	// Ensure proper size.
	if len(hash) != HashSize {
		t.Errorf("NewHash: hash length mismatch - got: %v, want: %v",
			len(hash), HashSize)
	}

	// Ensure contents match.
	if !bytes.Equal(hash[:], buf) {
		t.Errorf("NewHash: hash contents mismatch - got: %v, want: %v",
			hash[:], buf)
	}

	// Ensure contents of hash of block 234440 don't match 234439.
	if hash.IsEqual(blockHash) {
		t.Errorf("IsEqual: hash contents should not match - got: %v, want: %v",
			hash, blockHash)
	}

	// Ensure nil hashes are handled properly.
	if !(*Hash)(nil).IsEqual(nil) {
		t.Error("IsEqual: nil hashes should match")
	}
	if hash.IsEqual(nil) {
		t.Error("IsEqual: non-nil hash matches nil hash")
	}

	// Invalid size for SetBytes.
	err = hash.SetBytes([]byte{0x00})
	if err == nil {
		t.Errorf("SetBytes: failed to received expected err - got: nil")
	}

	// Invalid size for NewHash.
	invalidHash := make([]byte, HashSize+1)
	_, err = NewHash(invalidHash)
	if err == nil {
		t.Errorf("NewHash: failed to received expected err - got: nil")
	}
}
