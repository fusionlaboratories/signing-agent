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

/*
#cgo LDFLAGS: -lamcl_curve_SECP256K1 -lamcl_core
#cgo CFLAGS: -O2 -I ${SRCDIR}/include
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <amcl.h>
#include <randapi.h>
#include <ecdh_SECP256K1.h>
*/
import "C"
import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pkg/errors"
)

// SECP256K1 constants
const (
	EGSSECP256K1 = int(C.EGS_SECP256K1)
	EFSSECP256K1 = int(C.EFS_SECP256K1)
	EPSSECP256K1 = 2*EFSSECP256K1 + 1

	secpHashFunc   = C.HASH_TYPE_SECP256K1
	secpAESKeySize = int(C.AESKEY_SECP256K1)
)

var (
	paramP1 = CreateOctet([]byte{0, 1, 2})
	paramP2 = CreateOctet([]byte{0, 1, 2, 3})
)

// Secp256k1Encrypt encrypts a message using ECP_SECP256K1_ECIES
func Secp256k1Encrypt(message, publicKey string) (C, V, T string, err error) {
	dec := &hexBatchDecoder{}
	wOctet := dec.decodeOctet(publicKey)
	defer wOctet.Free()

	if dec.err != nil {
		err = dec.err
		return
	}

	seed := make([]byte, 32)
	_, err = rand.Read(seed)
	if err != nil {
		return
	}
	rng := NewRand(seed)

	mOctet := CreateOctet([]byte(message))
	defer mOctet.ClearAndFree()

	//Results
	vPtr := NewOctet(EPSSECP256K1)
	defer vPtr.Free()
	hmacPtr := NewOctet(secpHashFunc)
	defer hmacPtr.Free()
	cypherPtr := NewOctet(len(message) + secpAESKeySize - (len(message) % secpAESKeySize))
	defer cypherPtr.Free()

	C.ECP_SECP256K1_ECIES_ENCRYPT(secpHashFunc, paramP1, paramP2, (*C.csprng)(rng), (*C.octet)(wOctet), (*C.octet)(mOctet), C.int(12), vPtr, cypherPtr, hmacPtr)

	return hex.EncodeToString(cypherPtr.ToBytes()), hex.EncodeToString(vPtr.ToBytes()), hex.EncodeToString(hmacPtr.ToBytes()), nil
}

// Secp256k1Decrypt decrypts an encrypoted message using ECP_SECP256K1_ECIES
func Secp256k1Decrypt(C, V, T, sK string) (message string, err error) {
	dec := &hexBatchDecoder{}
	cOct := dec.decodeOctet(C)
	defer cOct.Free()
	vOct := dec.decodeOctet(V)
	defer vOct.Free()
	tOct := dec.decodeOctet(T)
	defer tOct.Free()
	uOct := dec.decodeOctet(sK)
	defer uOct.Free()

	if dec.err != nil {
		err = dec.err
		return
	}

	//Cast the cipherText back to Octets
	mOct := NewOctet(len(C) + 16 - (len(C) % 16))
	defer mOct.ClearAndFree()

	if C.ECP_SECP256K1_ECIES_DECRYPT(secpHashFunc, paramP1, paramP2, vOct, cOct, tOct, uOct, mOct) != 1 {
		return "", errors.New("Cannot decrypt cipherText")
	}

	b := mOct.ToBytes()
	return string(b), nil
}

type hexBatchDecoder struct {
	err error
}

func (d *hexBatchDecoder) decodeOctet(s string) *Octet {
	if d.err != nil {
		return nil
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		d.err = err
		return nil
	}

	return CreateOctet(b)
}
