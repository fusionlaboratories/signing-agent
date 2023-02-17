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
#include <amcl/randapi.h>
*/
import "C"
import (
	"encoding/hex"
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	curveName = "BLS381"
)

// Error codes
const (
	amclOk           = 0
	mpinInvalidPoint = -14
	mpinInvalidPin   = -19
	blsFail          = 41
	blsInvalidG1     = 42
	blsInvalidG2     = 43
)

var (
	// ErrInvalidPoint is binding for C.MPIN_INVALID_POINT
	ErrInvalidPoint = errors.New("Invalid point")
	// ErrInvalidPin is binding to C.MPIN_BAD_PIN
	ErrInvalidPin = errors.New("Invalid PIN")
	// ErrInvalidTime is returned when the timestamp in One pass is out of bounds
	ErrInvalidTime = errors.New("Invalid time")
	// ErrBlsFail is binding for C.BLS_FAIL
	ErrBlsFail = errors.New("Invalid BLS signature")
	// ErrInvalidG1 is binding for C.BLS_INVALID_G1
	ErrInvalidG1 = errors.New("Invalid G1 point")
	// ErrInvalidG2 is binding for C.BLS_INVALID_G2
	ErrInvalidG2 = errors.New("Invalid G2 point")
)

// Octet adds functionality around C octet
type Octet = C.octet

// NewOctet creates an empty Octet with a given size
func NewOctet(maxSize int) *Octet {
	return &Octet{
		len: C.int(0),
		max: C.int(maxSize),
		val: (*C.char)(C.calloc(1, C.size_t(maxSize))),
	}
}

// CreateOctet creates new Octet with a value
func CreateOctet(val []byte) *Octet {
	if val == nil {
		return nil
	}

	return &Octet{
		len: C.int(len(val)),
		max: C.int(len(val)),
		val: C.CString(string(val)),
	}
}

// Clear clears the octet memory
func (o *Octet) Clear() {
	if o == nil {
		return
	}

	C.OCT_clear(o)
}

// Free frees the allocated memory
func (o *Octet) Free() {
	if o == nil {
		return
	}

	C.free(unsafe.Pointer(o.val))
}

func (o *Octet) ClearAndFree() {
	o.Clear()
	o.Free()
}

// ToBytes returns the bytes representation of the Octet
func (o *Octet) ToBytes() []byte {
	return C.GoBytes(unsafe.Pointer(o.val), o.len)
}

// ToString returns the hex encoded representation of the Octet
func (o *Octet) ToString() string {
	return hex.EncodeToString(o.ToBytes())
}

// Rand is a cryptographically secure random number generator
type Rand C.csprng

// NewRand create new seeded Rand
func NewRand(seed []byte) *Rand {
	if seed == nil {
		return nil
	}
	sOct := CreateOctet(seed)
	defer sOct.ClearAndFree()

	var rand C.csprng
	C.CREATE_CSPRNG(&rand, sOct)
	return (*Rand)(&rand)
}

// GetByte returns one random byte
func (rand *Rand) GetByte() byte {
	r := C.RAND_byte((*C.csprng)(rand))
	return byte(r)
}

func (rand *Rand) csprng() *C.csprng {
	return (*C.csprng)(rand)
}

func intToC(i int) C.int {
	return C.int(i)
}

func codeToError(code C.int) error {
	switch code {
	case amclOk:
		return nil
	case mpinInvalidPoint:
		return ErrInvalidPoint
	case mpinInvalidPin:
		return ErrInvalidPin
	case blsFail:
		return ErrBlsFail
	case blsInvalidG1:
		return ErrInvalidG1
	case blsInvalidG2:
		return ErrInvalidG2
	}

	return fmt.Errorf("AMCL error code %d", int(code))
}
