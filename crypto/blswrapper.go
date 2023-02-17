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
	"fmt"
)

//GenerateBLSKeys - generate BLS12-381 Pub/Priv key from seed
func GenerateBLSKeys(seed []byte) (blsPublic, blsSecret []byte, err error) {
	if seed == nil {
		return nil, nil, fmt.Errorf("nil seed input")
	}
	blsPublic, blsSecret, err = BLSKeys(NewRand(seed), nil)
	if err != nil {
		err = fmt.Errorf("failed to generate BLS keys: %v", err)
	}
	return
}

// GenerateBLSKeysV2 - generate BLS12-381 Pub/Priv key from seed using the version 2 BLS KDF implementation
// The input bit seed length must be a minumum of
func GenerateBLSKeysV2(seed []byte) (blsPublic, blsSecret []byte, err error) {
	if seed == nil {
		return nil, nil, fmt.Errorf("nil seed input")
	}
	// Derive master secret key from input seed
	masterKey, err := DeriveMasterSK(seed)
	if err != nil {
		return nil, nil, fmt.Errorf("DeriveMasterSK: %v", err)
	}
	blsPublic, blsSecret, err = BLSKeys(nil, reverseBytes(toBytes(masterKey, RecommendedSeedLen)))
	if err != nil {
		err = fmt.Errorf("failed to generate BLS keys: %v", err)
	}
	return
}
