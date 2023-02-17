// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Smoke_Test(t *testing.T) {
	SEEDHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f30"
	SEED, _ := hex.DecodeString(SEEDHex)

	// Messsage to encrypt and sign. Note it is zero padded.
	PHex := "48656c6c6f20426f622120546869732069732061206d6573736167652066726f6d20416c696365000000000000000000"

	// Generate BLS keys
	BLSpk, BLSsk, err := BLSKeys(NewRand(SEED), nil)
	if err != nil {
		t.Fatal("Failed to create BLS keys")
	}

	// Encrypt message
	P1, _ := hex.DecodeString(PHex)

	// BLS Sign a message
	S, err := BLSSign(P1, BLSsk)
	if err != nil {
		t.Fatal("Failed to sign message")
		return
	}

	// BLS verify signature
	if err := BLSVerify(P1, BLSpk, S); err != nil {
		t.Fatal("BLS verify fail")
	}
}

func TestBLS(t *testing.T) {
	seedHex := "3370f613c4fe81130b846483c99c032c17dcc1904806cc719ed824351c87b0485c05089aa34ba1e1c6bfb6d72269b150"
	seed, _ := hex.DecodeString(seedHex)

	messageStr := "test message"
	message := []byte(messageStr)

	pk1, sk1, err := BLSKeys(NewRand(seed), nil)
	assert.Equal(t, err, nil, "Should be equal")

	sig1, err := BLSSign(message, sk1)
	assert.Equal(t, nil, err, "Should be equal")

	if err := BLSVerify(message, pk1, sig1); err != nil {
		t.Fatal("BLS verify fail")
	}
}

func TestBLSADD(t *testing.T) {
	seed1Hex := "3370f613c4fe81130b846483c99c032c17dcc1904806cc719ed824351c87b0485c05089aa34ba1e1c6bfb6d72269b150"
	seed2Hex := "46389f32b7cdebbbc46b7165d8fae888c9de444898390a939977e1a066256a6f465e7d76307178aef81ae0c6841f9b7c"
	seed1, _ := hex.DecodeString(seed1Hex)
	seed2, _ := hex.DecodeString(seed2Hex)

	messageStr := "test message"
	message := []byte(messageStr)

	pk12GoldenHex := "0fff41dc3b28fee38f564158f9e391a5c6ac42179fcccdf5ee4513030b6d59900a832f9a886b2407dc8b0a3b51921326123d3974bd1864fb22f5a84e83f1f9f611ee082ed5bd6ca896d464f12907ba8acdf15c44f9cff2a2dbb3b32259a1fe4f11d470158066087363df20a11144d6521cf72dca1a7514154a95c7fe73b219989cc40d7fc7e0b97854fc3123c0cf50ae0452730996a5cb24641aff7102fcbb2af705d0f32d5787ca1c3654e4ae6aa59106e1e22e29018ba7c341f1e6472f800f"
	sig12GoldenHex := "0203799dc2941b810985d9eb694a5be4a1ad5817f9e5d7c31870bb9fb471f7353eafacdc548544f9e7b78a0a9372c63ab0"
	pk12Golden, _ := hex.DecodeString(pk12GoldenHex)
	sig12Golden, _ := hex.DecodeString(sig12GoldenHex)

	pktmp, sktmp, err := BLSKeys(NewRand(seed1), nil)
	assert.Equal(t, nil, err, "Should be equal")

	pk1, sk1, err := BLSKeys(nil, sktmp)
	assert.Equal(t, nil, err, "Should be equal")

	assert.Equal(t, pktmp, pk1, "Should be equal")
	assert.Equal(t, sktmp, sk1, "Should be equal")

	pk2, sk2, err := BLSKeys(NewRand(seed2), nil)
	assert.Equal(t, nil, err, "Should be equal")

	sig1, err := BLSSign(message, sk1)
	assert.Equal(t, nil, err, "Should be equal")

	sig2, err := BLSSign(message, sk2)
	assert.Equal(t, nil, err, "Should be equal")

	if err := BLSVerify(message, pk1, sig1); err != nil {
		t.Fatal("BLS verify fail")
	}

	if err := BLSVerify(message, pk2, sig2); err != nil {
		t.Fatal("BLS verify fail")
	}

	sig12, err := BLSAddG1(sig1, sig2)
	assert.Equal(t, nil, err, "Should be equal")

	pk12, err := BLSAddG2(pk1, pk2)
	assert.Equal(t, nil, err, "Should be equal")

	if err := BLSVerify(message, pk12, sig12); err != nil {
		t.Fatal("BLS verify fail")
	}

	assert.Equal(t, pk12, pk12Golden, "Should be equal")
	assert.Equal(t, sig12, sig12Golden, "Should be equal")
}

// Test_BLSV2 generates BLS key pairs using methods V1 and V2, verifies key and signature generation
func Test_BLSV2(t *testing.T) {
	var seed []byte
	var signingMessage = []byte("raw-message")
	sigHash := sha256.Sum256(signingMessage)
	for i := 0; i < 96; i++ {
		if i > 0 {
			seed = make([]byte, i)
		}
		pk1, sk1, err := GenerateBLSKeys(seed)
		if err == nil && i == 0 {
			t.Errorf("nil input seed should result in error")
		} else if err != nil && i > 0 {
			t.Errorf("v1 bls key generation error: %v", err)
		}

		pk2, sk2, err := GenerateBLSKeysV2(seed)
		if err == nil && i < MinSeedBytes {
			t.Errorf("nil input seed should result in error")
		} else if err != nil && i >= MinSeedBytes {
			t.Errorf("v1 bls key generation error: %v", err)
		}

		if i < MinSeedBytes {
			// Key generation should have returned error so do not attempt to generate signatures
			continue
		} else {
			// V1 and V2 methods should yeild different sk values from the same input
			if bytes.Equal(sk1, sk2) || bytes.Equal(pk1, pk2) {
				t.Errorf("(%v) v1 and v2 bls key generation should yield different results for non-nil seed input \nsk1: %x\n sk2: %x\n pk1: %x\n pk2:%x\n", i, sk1, sk2, pk1, pk2)
				continue
			}
		}
		// Sign message with v1 key
		sig1, err := BLSSign(sigHash[:], sk1)
		if err != nil {
			t.Errorf("(%v) fail to generate v1 signature: %v", i, err)
		}
		// Verify message with v1 key
		if err = BLSVerify(sigHash[:], pk1, sig1); err != nil {
			t.Errorf("(%v) invalid v1 signature: %v", i, err)
		}
		// Sign message with v2 key
		sig2, err := BLSSign(sigHash[:], sk2)
		if err != nil {
			t.Errorf("(%v) fail to generate v2 signature: %v", i, err)
		}
		// Verify message with v2 key
		if err = BLSVerify(sigHash[:], pk2, sig2); err != nil {
			t.Errorf("(%v) invalid v2 signature: %v", i, err)
		}
	}

}
