package crypto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type testCase struct {
	TestID     int    `json:"TEST"`
	Curve      string `json:"CURVE"`
	MS1        string `json:"MS1"`
	MS2        string `json:"MS2"`
	SS1        string `json:"SS1"`
	SS2        string `json:"SS2"`
	SS         string `json:"SS"`
	CS1        string `json:"CS1"`
	CS2        string `json:"CS2"`
	CS         string `json:"CS"`
	Pin1       int    `json:"PIN1"`
	Pin2       int    `json:"PIN2"`
	Token      string `json:"TOKEN"`
	MpinID     string `json:"MPINId"`
	U          string `json:"U"`
	Y          string `json:"Y"`
	V          string `json:"V"`
	X          string `json:"X"`
	AuthResult int    `json:"AuthResult"`
}

type testCases []testCase

func setupTestCases(t *testing.T) testCases {
	tcData, err := os.ReadFile("testVectors/mpin_bls381.json")
	if err != nil {
		t.Fatal(err)
	}
	tc := testCases{}
	if err := json.Unmarshal(tcData, &tc); err != nil {
		t.Fatal(err)
	}

	return tc
}

func unhex(t *testing.T, hexstr string) []byte {
	out, err := hex.DecodeString(hexstr)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestTwoPass(t *testing.T) {
	cases := setupTestCases(t)
	rng := NewRand([]byte{0})

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case %d, TestVector %d", i, tc.TestID), func(t *testing.T) {
			ms1 := unhex(t, tc.MS1)
			ms2 := unhex(t, tc.MS2)

			id := unhex(t, tc.MpinID)
			hashID := hashID(id)

			// Setup part

			// Server secret shares
			ss1, err := GetServerSecret(ms1)
			if err != nil {
				t.Fatal(err)
			}
			ss2, err := GetServerSecret(ms2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ss1, unhex(t, tc.SS1)) {
				t.Fatalf("Server secret 1 don't match. Expected: %s, found: %x", tc.SS1, ss1)
			}
			if !bytes.Equal(ss2, unhex(t, tc.SS2)) {
				t.Fatalf("Server secret 2 don't match. Expected: %s, found: %x", tc.SS2, ss2)
			}
			// Full server secret
			ss, err := RecombineServerSecret(ss1, ss2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ss, unhex(t, tc.SS)) {
				t.Fatalf("Full Server secret don't match. Expected: %s, found: %x", tc.SS, ss)
			}

			// Client secrets shares
			cs1, err := GetClientSecret(ms1, hashID)
			if err != nil {
				t.Fatal(err)
			}
			cs2, err := GetClientSecret(ms2, hashID)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cs1, unhex(t, tc.CS1)) {
				t.Fatalf("Client secret 1 don't match. Expected: %s, found: %x", tc.CS1, cs1)
			}
			if !bytes.Equal(cs2, unhex(t, tc.CS2)) {
				t.Fatalf("Client secret 2 don't match. Expected: %s, found: %x", tc.CS2, cs2)
			}
			// Full client secret
			cs, err := RecombineClientSecret(cs1, cs2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cs, unhex(t, tc.CS)) {
				t.Fatalf("Full Client secret don't match. Expected: %s, found: %x", tc.CS, cs)
			}

			// Extract PIN
			token, err := ExtractPIN(id, tc.Pin1, cs)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(token, unhex(t, tc.Token)) {
				t.Fatalf("Token after PIN don't match. Expected: %s, found: %x", tc.Token, token)
			}

			// Authenticate part

			// TWO PASS AUTHENTICATION ============================================================================

			// Compute Client Pass1
			// If a predefined X is in use, the RNG SHOULD BE nil
			cpass1, err := ClientPass1(id, tc.Pin2, nil, token, WithPredefinedX(unhex(t, tc.X)))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cpass1.X, unhex(t, tc.X)) {
				t.Fatalf("Client Pass1: X value doesn't match. Expected: %s, found: %x", tc.X, cpass1.X)
			}
			if !bytes.Equal(cpass1.U, unhex(t, tc.U)) {
				t.Fatalf("Client Pass1: U value doesn't match. Expected: %s, found: %x", tc.U, cpass1.U)
			}

			// Compute Server Pass1
			pass1, err := ServerPass1(id, rng)
			if err != nil {
				t.Fatal(err)
			}
			// Server sends Y value to the client
			// Use Y from test vectors
			y := unhex(t, tc.Y)
			v, err := ClientPass2(cpass1, y)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(v, unhex(t, tc.V)) {
				t.Fatalf("Client Pass2: V value doesn't match. Expected: %s, found: %x", tc.V, v)
			}

			expErr := codeToError(intToC(tc.AuthResult))
			if err := ServerPass2(pass1.HID, pass1.HTID, y, ss, cpass1.U, cpass1.UT, v, nil); expErr != err {
				t.Fatalf("Server Pass2: Expecting err %v, found: %v", expErr, err)
			}

		})
	}
}

func TestOnePass(t *testing.T) {
	cases := setupTestCases(t)
	rng := NewRand([]byte{0})

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case %d, TestVector %d", i, tc.TestID), func(t *testing.T) {
			ms1 := unhex(t, tc.MS1)
			ms2 := unhex(t, tc.MS2)

			id := unhex(t, tc.MpinID)
			hashID := hashID(id)

			// Setup part

			// Server secret shares
			ss1, err := GetServerSecret(ms1)
			if err != nil {
				t.Fatal(err)
			}
			ss2, err := GetServerSecret(ms2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ss1, unhex(t, tc.SS1)) {
				t.Fatalf("Server secret 1 don't match. Expected: %s, found: %x", tc.SS1, ss1)
			}
			if !bytes.Equal(ss2, unhex(t, tc.SS2)) {
				t.Fatalf("Server secret 2 don't match. Expected: %s, found: %x", tc.SS2, ss2)
			}
			// Full server secret
			ss, err := RecombineServerSecret(ss1, ss2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ss, unhex(t, tc.SS)) {
				t.Fatalf("Full Server secret don't match. Expected: %s, found: %x", tc.SS, ss)
			}

			// Client secrets shares
			cs1, err := GetClientSecret(ms1, hashID)
			if err != nil {
				t.Fatal(err)
			}
			cs2, err := GetClientSecret(ms2, hashID)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cs1, unhex(t, tc.CS1)) {
				t.Fatalf("Client secret 1 don't match. Expected: %s, found: %x", tc.CS1, cs1)
			}
			if !bytes.Equal(cs2, unhex(t, tc.CS2)) {
				t.Fatalf("Client secret 2 don't match. Expected: %s, found: %x", tc.CS2, cs2)
			}
			// Full client secret
			cs, err := RecombineClientSecret(cs1, cs2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cs, unhex(t, tc.CS)) {
				t.Fatalf("Full Client secret don't match. Expected: %s, found: %x", tc.CS, cs)
			}

			// Extract PIN
			token, err := ExtractPIN(id, tc.Pin1, cs)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(token, unhex(t, tc.Token)) {
				t.Fatalf("Token after PIN don't match. Expected: %s, found: %x", tc.Token, token)
			}

			// Authenticate part

			// ONE PASS AUTHENTICATION ============================================================================
			expErr := codeToError(intToC(tc.AuthResult))
			client1, err := ClientOnePass(id, tc.Pin2, rng, token, nil)
			if err != nil {
				t.Fatalf("Client One Pass: %v", err)
			}
			if err := ServerOnePass(client1, ss, nil, 1); err != expErr {
				t.Fatalf("Server One Pass: %v", err)
			}

			msgclient1, err := ClientOnePass(id, tc.Pin2, rng, token, []byte{1, 2, 3})
			if err != nil {
				t.Fatalf("Client One Pass: %v", err)
			}

			if err := ServerOnePass(msgclient1, ss, []byte{1, 2, 3}, 1); err != expErr {
				t.Fatalf("Server One Pass: %v", err)
			}

		})
	}
}
