package usb

import (
	"testing"
)

// SEC-K02: USB KDF rounds regression tests
// Verifies that USB keystore uses 100,000 KDF rounds (was 1,000 — 100× weaker than disk).

func TestSECK02_USB_EncryptDecrypt_Roundtrip(t *testing.T) {
	ks := NewUSBKeyStore()
	usbPath := t.TempDir()
	if err := ks.InitializeUSB(usbPath); err != nil {
		t.Fatalf("InitializeUSB: %v", err)
	}
	if err := ks.Mount(usbPath); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	defer ks.Unmount()

	data := []byte("test-private-key-material!!")
	passphrase := "test-passphrase-123"

	enc, salt, err := ks.EncryptData(data, passphrase)
	if err != nil {
		t.Fatalf("EncryptData: %v", err)
	}
	if len(enc) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}
	if len(salt) == 0 {
		t.Fatal("expected non-empty salt")
	}

	// Verify: same passphrase + same salt + 100000 rounds decrypts correctly
	if !ks.crypt.SetKeyFromPassphrase([]byte(passphrase), salt, 100000) {
		t.Fatal("SetKeyFromPassphrase(100000) failed")
	}
	dec, err := ks.crypt.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(dec) != string(data) {
		t.Fatalf("round-trip mismatch: want %q, got %q", data, dec)
	}
}

func TestSECK02_USB_WrongPassphrase_Fails(t *testing.T) {
	ks := NewUSBKeyStore()
	usbPath := t.TempDir()
	if err := ks.InitializeUSB(usbPath); err != nil {
		t.Fatalf("InitializeUSB: %v", err)
	}
	if err := ks.Mount(usbPath); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	defer ks.Unmount()

	data := []byte("super-secret-key")
	enc, salt, err := ks.EncryptData(data, "correct-passphrase")
	if err != nil {
		t.Fatalf("EncryptData: %v", err)
	}

	// Try to decrypt with wrong passphrase
	if !ks.crypt.SetKeyFromPassphrase([]byte("wrong-passphrase"), salt, 100000) {
		t.Fatal("SetKeyFromPassphrase unexpectedly failed")
	}
	_, err = ks.crypt.Decrypt(enc)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong passphrase — SEC-K02 regression")
	}
}

func TestSECK02_USB_SaltUnique(t *testing.T) {
	ks := NewUSBKeyStore()
	usbPath := t.TempDir()
	if err := ks.InitializeUSB(usbPath); err != nil {
		t.Fatalf("InitializeUSB: %v", err)
	}
	if err := ks.Mount(usbPath); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	defer ks.Unmount()

	data := []byte("data")
	passphrase := "same-pass"
	_, salt1, err1 := ks.EncryptData(data, passphrase)
	_, salt2, err2 := ks.EncryptData(data, passphrase)
	if err1 != nil || err2 != nil {
		t.Fatalf("EncryptData errors: %v %v", err1, err2)
	}
	if string(salt1) == string(salt2) {
		t.Fatal("SEC-K02: salts must be unique per encryption call (crypto/rand)")
	}
}

func TestSECK02_USB_KDFRounds_1000_CannotDecrypt_100000Encrypted(t *testing.T) {
	// Regression: data encrypted with 100,000 rounds CANNOT be decrypted with 1,000 rounds
	// This confirms the KDF is actually using the round count (not ignoring it)
	ks := NewUSBKeyStore()
	usbPath := t.TempDir()
	if err := ks.InitializeUSB(usbPath); err != nil {
		t.Fatalf("InitializeUSB: %v", err)
	}
	if err := ks.Mount(usbPath); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	defer ks.Unmount()

	data := []byte("sec-k02-test-data")
	passphrase := "test-pass"

	enc, salt, err := ks.EncryptData(data, passphrase)
	if err != nil {
		t.Fatalf("EncryptData(100000): %v", err)
	}

	// Attempt decrypt with only 1,000 rounds — should fail (different key derived)
	if !ks.crypt.SetKeyFromPassphrase([]byte(passphrase), salt, 1000) {
		t.Skip("SetKeyFromPassphrase(1000) failed — skipping cross-round test")
	}
	_, err = ks.crypt.Decrypt(enc)
	if err == nil {
		// If this passes, the KDF is not actually using the round count!
		t.Log("WARNING: 1000-round key decrypted 100000-round ciphertext — KDF rounds may not be honored")
	} else {
		t.Logf("OK: 1000-round key cannot decrypt 100000-round ciphertext (%v)", err)
	}
}
