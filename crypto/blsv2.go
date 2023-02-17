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
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"golang.org/x/crypto/hkdf"
)

// Qredo BLS KDF based on Filecoin's bls12-381 HD key generation implementation: https://github.com/filcloud/bls12-381/

// EIP-2333: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2333.md

const (
	RecommendedSeedLen = 32 // 256 bits
	MinSeedBytes       = 16 // 128 bits
	MaxSeedBytes       = 64 // 512 bits
	salt               = "BLS-SIG-KEYGEN-SALT-"

	//noinspection GoSnakeCaseUsage
	hkdf_mod_r_L = 48
)

var (
	ErrInvalidSeedLen = fmt.Errorf("seed length must be between %d and %d bits",
		MinSeedBytes*8, MaxSeedBytes*8)
	curveOrder, _ = big.NewInt(0).SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 0) // BLD12-381 group order
)

// DeriveMasterSK creates a master private key using the supplied
// seed as entropy
func DeriveMasterSK(seed []byte) (*big.Int, error) {
	if len(seed) < MinSeedBytes {
		return nil, ErrInvalidSeedLen
	}
	return hkdf_mod_r(seed)
}

// https://github.com/filcloud/bls12-381/blob/master/keyderivation.go#L74
func hkdf_mod_r(IKM []byte) (*big.Int, error) {
	okm := make([]byte, hkdf_mod_r_L)
	r := hkdf.New(sha256.New, IKM, []byte(salt), nil)
	n, err := r.Read(okm)
	if err != nil {
		return nil, err
	}
	if n != hkdf_mod_r_L {
		return nil, errors.New("hkdf: entropy limit reached")
	}
	bi := big.NewInt(0).SetBytes(okm)
	return big.NewInt(0).Mod(bi, curveOrder), nil
}

func toBytes(b *big.Int, length int) []byte {
	result := make([]byte, length)
	bytes := b.Bytes()
	copy(result[length-len(bytes):], bytes)
	return result
}

func reverseBytes(a []byte) []byte {
	b := make([]byte, len(a))
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		b[i], b[opp] = a[opp], a[i]
	}
	return b
}
