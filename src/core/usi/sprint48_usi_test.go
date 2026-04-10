package usi

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Sprint 48 — USI additional error path coverage (cheap paths only)

// helper to build valid file_hash + fingerprint meta with bad sig
func makeMetaValidHashFP(data []byte) *USIMeta {
	fh := sha256.Sum256(data)
	pk := []byte("fake-public-key-32-bytes-exactly!!")
	fp := sha256.Sum256(pk)
	return &USIMeta{
		Version:      USIMetaVersion,
		Fingerprint:  hex.EncodeToString(fp[:]),
		PublicKey:    hex.EncodeToString(pk),
		FileHash:     hex.EncodeToString(fh[:]),
		FinalDocHash: hex.EncodeToString(fh[:]),
		Signature:    hex.EncodeToString(make([]byte, 32)),
		Nonce:        hex.EncodeToString(make([]byte, 16)),
		Commitment:   hex.EncodeToString(make([]byte, 32)),
		MerkleRoot:   hex.EncodeToString(make([]byte, 32)),
	}
}

func TestVerifyUSIMeta_InvalidSigHex_Step3Early(t *testing.T) {
	// sig hex decode fails before reaching km
	meta := makeMetaValidHashFP([]byte("hello"))
	meta.Signature = "not-valid-hex!!" // invalid hex → error at step 3
	ok, err := VerifyUSIMeta([]byte("hello"), meta, nil, nil)
	if ok || err == nil {
		t.Error("invalid signature hex should return (false, error)")
	}
}

func TestVerifyUSIMeta_InvalidNonceHex_Step3b(t *testing.T) {
	// valid sig hex, invalid nonce hex
	meta := makeMetaValidHashFP([]byte("hello"))
	meta.Nonce = "!!"
	ok, err := VerifyUSIMeta([]byte("hello"), meta, nil, nil)
	if ok || err == nil {
		t.Error("invalid nonce hex should return (false, error)")
	}
}

// Vault: additional edge path coverage
func TestCreateVault_EmptyCreator_Error(t *testing.T) {
	_, err := CreateVault([]byte("data"), "", nil, "pass")
	if err == nil {
		t.Error("CreateVault with empty creator should return error")
	}
}

func TestOpenVault_NilVault_Error(t *testing.T) {
	_, err := OpenVault(nil, "fp", "pass")
	if err == nil {
		t.Error("OpenVault with nil vault should return error")
	}
}

func TestOpenVault_WrongCallerFP_Error(t *testing.T) {
	v, err := CreateVault([]byte("payload"), "creator-fp", []string{"recipient-fp"}, "mypassphrase")
	if err != nil {
		t.Fatalf("CreateVault: %v", err)
	}
	_, err = OpenVault(v, "wrong-fp", "mypassphrase")
	if err == nil {
		t.Error("OpenVault with wrong caller FP should return error")
	}
}

func TestOpenVault_WrongPassphrase_Error(t *testing.T) {
	v, err := CreateVault([]byte("payload"), "creator-fp", []string{"creator-fp"}, "correctpass")
	if err != nil {
		t.Fatalf("CreateVault: %v", err)
	}
	_, err = OpenVault(v, "creator-fp", "wrongpass")
	if err == nil {
		t.Error("OpenVault with wrong passphrase should return error")
	}
}

func TestCreateVault_CreatorIsRecipient(t *testing.T) {
	v, err := CreateVault([]byte("hello world"), "creator-fp", []string{"creator-fp"}, "pass")
	if err != nil {
		t.Fatalf("CreateVault: %v", err)
	}
	data, err := OpenVault(v, "creator-fp", "pass")
	if err != nil {
		t.Fatalf("OpenVault: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("roundtrip: got %q, want %q", data, "hello world")
	}
}

// SignData error paths
func TestSignData_EmptyData_Returns_Error(t *testing.T) {
	_, err := SignData([]byte{}, "fp", []byte("sk"), []byte("pk"), nil, nil)
	if err == nil {
		t.Error("SignData with empty data should error")
	}
}

func TestSignData_EmptySK_Returns_Error(t *testing.T) {
	_, err := SignData([]byte("data"), "fp", nil, []byte("pk"), nil, nil)
	if err == nil {
		t.Error("SignData with nil SK should error")
	}
}

func TestSignData_EmptyPK_Returns_Error(t *testing.T) {
	_, err := SignData([]byte("data"), "fp", []byte("sk"), nil, nil, nil)
	if err == nil {
		t.Error("SignData with nil PK should error")
	}
}


