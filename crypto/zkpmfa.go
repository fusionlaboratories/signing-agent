package crypto

// #cgo LDFLAGS: -lamcl_core -lamcl_curve_BLS381 -lamcl_mpin_BLS381 -lamcl_pairing_BLS381
// #cgo CFLAGS: -I${SRCDIR}/include
// #include "amcl.h"
// #include "mpin_BLS381.h"
// #include "randapi.h"
// #include "utils.h"
import "C"
import (
	"math"
	"time"
)

// BLS381 constants
const (
	PGSBLS381 = int(C.PGS_BLS381)
	PFSBLS381 = int(C.PFS_BLS381)
	G1SBLS381 = 2*PFSBLS381 + 1
	G2SBLS381 = 4 * PFSBLS381

	zkpHashFunc = C.HASH_TYPE_BLS381
)

// ClientPass1Result holds the result of the Client Pass1
type ClientPass1Result struct {
	X   []byte
	SEC []byte
	U   []byte
	UT  []byte
}

// ServerPass1Result holds the result of the Server Pass1
type ServerPass1Result struct {
	Y    []byte
	HID  []byte
	HTID []byte
	U    []byte
	UT   []byte
}

// Client1PassResult is holds the result of One-Pass Client
type Client1PassResult struct {
	ID []byte
	ET int64
	U  []byte
	V  []byte
}

// NewMasterSecret generates a new random master secret
func NewMasterSecret(rand *Rand) (secret []byte, err error) {
	return randomGenerate(rand)
}

// GetServerSecret generates the server secret of the master secret
func GetServerSecret(ms []byte) (secret []byte, err error) {
	msOct := CreateOctet(ms)
	defer msOct.ClearAndFree()
	ssOct := NewOctet(G2SBLS381)
	defer ssOct.ClearAndFree()

	code := C.MPIN_BLS381_GET_SERVER_SECRET(msOct, ssOct)

	return ssOct.ToBytes(), codeToError(code)
}

// GetClientSecret generates the client secret of the identity and the master secret
func GetClientSecret(ms []byte, id []byte) (secret []byte, err error) {
	msOct := CreateOctet(ms)
	defer msOct.ClearAndFree()
	idOct := CreateOctet(id)
	defer idOct.Free()
	csOct := NewOctet(G1SBLS381)
	defer csOct.ClearAndFree()

	code := C.MPIN_BLS381_GET_CLIENT_SECRET(msOct, idOct, csOct)

	return csOct.ToBytes(), codeToError(code)
}

func recombineG1(p1 []byte, p2 []byte) (pCombine []byte, err error) {
	p1Oct := CreateOctet(p1)
	defer p1Oct.ClearAndFree()
	p2Oct := CreateOctet(p2)
	defer p2Oct.ClearAndFree()
	pOct := NewOctet(G1SBLS381)
	defer pOct.ClearAndFree()

	code := C.MPIN_BLS381_RECOMBINE_G1(p1Oct, p2Oct, pOct)

	return pOct.ToBytes(), codeToError(code)
}

func recombineG2(p1 []byte, p2 []byte) (pCombine []byte, err error) {
	p1Oct := CreateOctet(p1)
	defer p1Oct.ClearAndFree()
	p2Oct := CreateOctet(p2)
	defer p2Oct.ClearAndFree()
	pOct := NewOctet(G2SBLS381)
	defer pOct.ClearAndFree()

	code := C.MPIN_BLS381_RECOMBINE_G2(p1Oct, p2Oct, pOct)

	return pOct.ToBytes(), codeToError(code)
}

// RecombineServerSecret combines the full server secret out of server secret shares
func RecombineServerSecret(shares ...[]byte) (secret []byte, err error) {
	for i, share := range shares {
		if i == 0 {
			secret = share
			continue
		}

		secret, err = recombineG2(secret, share)
		if err != nil {
			return
		}
	}

	return
}

// RecombineClientSecret combines the full client secret out of client secret shares
func RecombineClientSecret(shares ...[]byte) (secret []byte, err error) {
	for i, share := range shares {
		if i == 0 {
			secret = share
			continue
		}

		secret, err = recombineG1(secret, share)
		if err != nil {
			return
		}
	}

	return
}

// ExtractPIN extracts PIN from client secret and produces token
func ExtractPIN(id []byte, pin int, cs []byte) (token []byte, err error) {
	idOct := CreateOctet(id)
	defer idOct.Free()
	csOct := CreateOctet(cs)
	defer csOct.ClearAndFree()

	code := C.MPIN_BLS381_EXTRACT_PIN(zkpHashFunc, idOct, C.int(pin), csOct)

	return csOct.ToBytes(), codeToError(code)
}

type Client1Option = func(*ClientPass1Result) error

// WithPredefinedX is used to fix the X value for testing
func WithPredefinedX(x []byte) Client1Option {
	return func(cr *ClientPass1Result) error {
		cr.X = x
		return nil
	}
}

// ClientPass1 performs Pass1 on the client when using 2-pass protocol
func ClientPass1(id []byte, pin int, rng *Rand, token []byte, opts ...Client1Option) (*ClientPass1Result, error) {
	pass1R := &ClientPass1Result{}
	for _, opt := range opts {
		if err := opt(pass1R); err != nil {
			return nil, err
		}
	}

	var xOct *Octet
	if pass1R.X != nil {
		xOct = CreateOctet(pass1R.X)
	} else {
		xOct = NewOctet(PGSBLS381)
	}
	defer xOct.ClearAndFree()

	idOct := CreateOctet(id)
	defer idOct.Free()
	tOct := CreateOctet(token)
	defer tOct.ClearAndFree()
	sOct := NewOctet(G1SBLS381)
	defer sOct.ClearAndFree()
	uOct := NewOctet(G1SBLS381)
	defer uOct.Free()
	utOct := NewOctet(G1SBLS381)
	defer utOct.Free()

	code := C.MPIN_BLS381_CLIENT_1(zkpHashFunc, 0, idOct, rng.csprng(), xOct, C.int(pin), tOct, sOct, uOct, utOct, nil)
	if err := codeToError(code); err != nil {
		return nil, err
	}

	return &ClientPass1Result{
		X:   xOct.ToBytes(),
		SEC: sOct.ToBytes(),
		U:   uOct.ToBytes(),
		UT:  utOct.ToBytes(),
	}, nil
}

// ClientPass2 performs Pass2 on the client using ClientPass1Result and Y value from the server
func ClientPass2(p1r *ClientPass1Result, y []byte) (v []byte, err error) {
	xOct := CreateOctet(p1r.X)
	defer xOct.ClearAndFree()
	yOct := CreateOctet(y)
	defer yOct.Free()
	vOct := CreateOctet(p1r.SEC)
	defer vOct.Free()

	code := C.MPIN_BLS381_CLIENT_2(xOct, yOct, vOct)

	return vOct.ToBytes(), codeToError(code)
}

// ServerPass1 performs Pass1 on the server when using 2-pass protocol
func ServerPass1(id []byte, rand *Rand) (*ServerPass1Result, error) {
	idOct := CreateOctet(id)
	defer idOct.Free()
	hidOct := NewOctet(G1SBLS381)
	defer hidOct.Free()
	htidOct := NewOctet(G1SBLS381)
	defer htidOct.Free()

	C.MPIN_BLS381_SERVER_1(zkpHashFunc, C.int(0), idOct, hidOct, htidOct)

	y, err := randomGenerate(rand)
	if err != nil {
		return nil, err
	}

	return &ServerPass1Result{
		HID:  hidOct.ToBytes(),
		HTID: htidOct.ToBytes(),
		Y:    y,
	}, nil
}

// ServerPass2 performs Pass2 on server when using 2-pass protocol
// On successful authentication the err result is nil
func ServerPass2(hid []byte, htid []byte, y []byte, ss []byte, u []byte, ut []byte, v []byte, pa []byte) (err error) {
	hidOct := CreateOctet(hid)
	defer hidOct.Free()
	htidOct := CreateOctet(htid)
	defer htidOct.Free()
	yOct := CreateOctet(y)
	defer yOct.Free()
	ssOct := CreateOctet(ss)
	defer ssOct.ClearAndFree()
	uOct := CreateOctet(u)
	defer uOct.Free()
	utOct := CreateOctet(ut)
	defer utOct.Free()
	vOct := CreateOctet(v)
	defer vOct.Free()
	paOct := CreateOctet(pa)
	defer paOct.Free()

	code := C.MPIN_BLS381_SERVER_2(C.int(0), hidOct, htidOct, yOct, ssOct, uOct, utOct, vOct, nil, nil, paOct)

	return codeToError(code)
}

// ClientOnePass performs ZKP MFA One Pass on the client
func ClientOnePass(id []byte, pin int, rng *Rand, token []byte, msg []byte, opts ...Client1Option) (*Client1PassResult, error) {
	idOct := CreateOctet(id)
	defer idOct.Free()
	tOct := CreateOctet(token)
	defer tOct.ClearAndFree()
	uOct := NewOctet(G1SBLS381)
	defer uOct.Free()
	utOct := NewOctet(G1SBLS381)
	defer utOct.Free()
	xOct := NewOctet(PGSBLS381)
	defer xOct.ClearAndFree()
	yOct := NewOctet(PGSBLS381)
	defer yOct.Free()
	vOct := NewOctet(G1SBLS381)
	defer vOct.Free()

	var msgOct *Octet
	if msg == nil {
		msgOct = nil
	} else {
		msgOct = CreateOctet(msg)
		defer msgOct.Free()
	}

	timestamp := time.Now().Unix()
	code := C.MPIN_BLS381_CLIENT(zkpHashFunc, C.int(0), idOct, rng.csprng(), xOct, C.int(pin), tOct, vOct, uOct, utOct, nil, msgOct, C.int(timestamp), yOct)

	if err := codeToError(code); err != nil {
		return nil, err
	}

	return &Client1PassResult{
		ID: id,
		ET: timestamp,
		U:  uOct.ToBytes(),
		V:  vOct.ToBytes(),
	}, nil
}

// ServerOnePass performs ZKP MFA One Pass on the server
func ServerOnePass(client *Client1PassResult, ss []byte, msg []byte, timeBounds int64) error {
	if math.Abs(float64(time.Now().Unix()-client.ET)) > float64(timeBounds) {
		return ErrInvalidTime
	}

	ssOct := CreateOctet(ss)
	defer ssOct.ClearAndFree()
	idOct := CreateOctet(client.ID)
	defer idOct.Free()
	uOct := CreateOctet(client.U)
	defer uOct.Free()
	vOct := CreateOctet(client.V)
	defer vOct.Free()
	yOct := NewOctet(PGSBLS381)
	defer yOct.Free()
	hidOct := NewOctet(G1SBLS381)
	defer hidOct.Free()
	htidOct := NewOctet(G1SBLS381)
	defer htidOct.Free()

	var msgOct *Octet
	if msg == nil {
		msgOct = nil
	} else {
		msgOct = CreateOctet(msg)
		defer msgOct.Free()
	}

	code := C.MPIN_BLS381_SERVER(zkpHashFunc, C.int(0), hidOct, htidOct, yOct, ssOct, uOct, nil, vOct, nil, nil, idOct, msgOct, C.int(client.ET), nil)

	return codeToError(code)
}

// randomGenerate generates a random byte slice of size PGSBLS381
func randomGenerate(rand *Rand) (secret []byte, err error) {
	oct := NewOctet(PGSBLS381)
	defer oct.ClearAndFree()

	code := C.MPIN_BLS381_RANDOM_GENERATE((*C.csprng)(rand), oct)
	return oct.ToBytes(), codeToError(code)
}

// hashID produces the hash of the identity
func hashID(id []byte) []byte {
	idOct := CreateOctet(id)
	defer idOct.Free()

	hidOct := NewOctet(PFSBLS381)
	defer hidOct.Free()

	C.HASH_ID(zkpHashFunc, idOct, hidOct)

	return hidOct.ToBytes()
}
