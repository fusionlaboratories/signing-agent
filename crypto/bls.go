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

/*
Package crypto - wrapper for encryption libraries required by service
*/
package crypto

/*
#cgo CFLAGS:  -O2 -I${SRCDIR}/include
#cgo LDFLAGS: -L/usr/local/lib -lamcl_bls_BLS381 -lamcl_pairing_BLS381 -lamcl_curve_BLS381 -lamcl_core
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <amcl/utils.h>
#include <amcl/randapi.h>
#include <amcl/bls_BLS381.h>
*/
import "C"

const (
	// BFSBLS381 Field size
	BFSBLS381 = int(C.BFS_BLS381)
	// BGSBLS381 Group size
	BGSBLS381 = int(C.BGS_BLS381)
	// G2Len G2 point size
	G2Len = 4 * BFSBLS381
	// SIGLen Signature length
	SIGLen = BFSBLS381 + 1
)

/*BLSKeys Generate BLS keys

Generate public and private key pair. If the seed value is nil then
generate the public key using the input secret key.

@param rand             cspring PRNG.
@param ski              input secret key
@param pk               public key
@param sko              output secret key
@param err              Return code error
*/
func BLSKeys(rand *Rand, ski []byte) (pk []byte, sko []byte, err error) {
	// Allocate memory
	ppk := NewOctet(G2Len)
	defer ppk.Free()
	var sk *Octet
	if ski == nil {
		sk = NewOctet(BGSBLS381)
	} else {
		sk = CreateOctet(ski)
	}
	defer sk.ClearAndFree()

	if rand == nil {
		rc := C.BLS_BLS381_KEY_PAIR_GENERATE(nil, sk, ppk)

		return ppk.ToBytes(), sk.ToBytes(), codeToError(rc)
	}

	rc := C.BLS_BLS381_KEY_PAIR_GENERATE((*C.csprng)(rand), sk, ppk)

	return ppk.ToBytes(), sk.ToBytes(), codeToError(rc)
}

/*BLSSign Sign a message

  The message is signed using the BLS algorithm

  @param m            Message to be signed
  @param sk           secret key
  @param S            Signature
  @param err          Return code error
*/
func BLSSign(m []byte, sk []byte) (s []byte, err error) {
	mo := CreateOctet(m)
	defer mo.ClearAndFree()
	sko := CreateOctet(sk)
	defer sko.ClearAndFree()
	sigo := NewOctet(SIGLen)
	defer sigo.Free()

	rc := C.BLS_BLS381_SIGN(sigo, mo, sko)

	return sigo.ToBytes(), codeToError(rc)
}

/*BLSVerify Verify a signature

  Verify a signature using the BLS algorithm

  @param m            Message that was signed
  @param pk           public key
  @param S            Signature
  @param err          Return code error
*/
func BLSVerify(m []byte, pk []byte, s []byte) error {
	mo := CreateOctet(m)
	defer mo.ClearAndFree()
	pko := CreateOctet(pk)
	defer pko.Free()
	sigo := CreateOctet(s)
	defer sigo.Free()

	rc := C.BLS_BLS381_VERIFY(sigo, mo, pko)
	return codeToError(rc)
}

/*BLSAddG1 Add two members from the group G1

  Add two members from the group G1

  @param R1           member of G1
  @param R2           member of G1
  @param R            member of G1. r = r1+r2
  @param err          Return code error
*/
func BLSAddG1(R1 []byte, R2 []byte) (R []byte, err error) {
	r1 := CreateOctet(R1)
	defer r1.Free()
	r2 := CreateOctet(R2)
	defer r2.Free()
	r := NewOctet(SIGLen)
	defer r.Free()

	rc := C.BLS_BLS381_ADD_G1(r1, r2, r)

	return r.ToBytes(), codeToError(rc)
}

/*BLSAddG2 Add two members from the group G2

  Add two members from the group G2

  @param R1           member of G2
  @param R2           member of G2
  @param R            member of G2. r = r1+r2
  @param err          Return code error
*/
func BLSAddG2(R1 []byte, R2 []byte) (R []byte, err error) {
	r1 := CreateOctet(R1)
	defer r1.Free()
	r2 := CreateOctet(R2)
	defer r2.Free()
	r := NewOctet(G2Len)
	defer r.Free()

	rc := C.BLS_BLS381_ADD_G2(r1, r2, r)

	return r.ToBytes(), codeToError(rc)
}
