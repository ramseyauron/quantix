// MIT License
//
// Copyright (c) 2024 quantix
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package fingerprint implements the USI-compatible canonical identity type
// for the Quantix protocol.
//
// # Universal Sovereign Identity (USI) — Fingerprint
//
// In the Quantix/USI model, identity is NOT issued by institutions. It is
// derived entirely from cryptographic key ownership:
//
//   - A Fingerprint is the SHA-256 hash of a SPHINCS+ public key.
//   - 32 raw bytes, represented as 64 lowercase hex characters.
//   - Any entity that controls the private key controls the identity.
//   - No registration, no CA, no trusted third party.
//
// This is mathematically enforced privacy: an address can only be linked to
// a public key if the owner chooses to reveal it. Before revelation the
// Fingerprint is pseudonymous — it commits the owner to a key without
// exposing the key itself.
//
// ## Relationship to GenerateAddress (encode.go)
//
// GenerateAddress produces a Base58-with-checksum human-legible address for
// wallets and QR codes. Fingerprint is the lower-level primitive: it lives
// on-chain in transaction Sender/Receiver fields and in StateDB account keys.
// GenerateAddress is a human-friendly encoding of the Fingerprint.
//
// ## Wire format
//
//	Fingerprint = hex.EncodeToString(SHA-256(publicKey))
//	Length: exactly 64 hex characters (32 bytes)
//
// This matches USI paper §3.1 "Sovereign Identity Derivation".
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const (
	// Len is the number of hex characters in a canonical Fingerprint.
	// 32 bytes × 2 hex chars/byte = 64 characters.
	Len = 64
)

// Fingerprint is a 32-byte SHA-256 digest of a SPHINCS+ public key.
// It is the canonical on-chain identity for all Quantix accounts.
// No institution issues it — it is derived solely from key ownership.
type Fingerprint [32]byte

// ErrInvalid is returned when a string cannot be parsed as a valid Fingerprint
// (wrong length or non-hex characters).
var ErrInvalid = errors.New("invalid fingerprint: must be 64 lowercase hex characters")

// New derives a Fingerprint from a raw SPHINCS+ public key.
// This is the canonical constructor: SHA-256(publicKey).
// The caller is responsible for providing the correct public key bytes.
func New(publicKey []byte) Fingerprint {
	return sha256.Sum256(publicKey)
}

// Parse parses a 64-character hex string into a Fingerprint.
// Returns ErrInvalid if the input is malformed.
func Parse(s string) (Fingerprint, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != Len {
		return Fingerprint{}, ErrInvalid
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return Fingerprint{}, ErrInvalid
	}
	var fp Fingerprint
	copy(fp[:], b)
	return fp, nil
}

// String returns the canonical 64-character lowercase hex representation.
// This is the wire format used in transaction Sender/Receiver fields.
func (fp Fingerprint) String() string {
	return hex.EncodeToString(fp[:])
}

// Bytes returns a copy of the raw 32-byte digest.
func (fp Fingerprint) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, fp[:])
	return b
}

// Equal returns true if two Fingerprints represent the same identity.
func (fp Fingerprint) Equal(other Fingerprint) bool {
	return fp == other
}

// IsZero returns true if the Fingerprint is the zero value (unset).
// A zero Fingerprint is never a valid on-chain address.
func (fp Fingerprint) IsZero() bool {
	return fp == Fingerprint{}
}

// Validate returns nil if the Fingerprint is non-zero, or ErrInvalid if it is
// the zero value (indicating an uninitialised identity).
func (fp Fingerprint) Validate() error {
	if fp.IsZero() {
		return ErrInvalid
	}
	return nil
}

// MarshalText implements encoding.TextMarshaler so Fingerprint serialises as
// its 64-character hex string in JSON and other text-based encodings.
func (fp Fingerprint) MarshalText() ([]byte, error) {
	return []byte(fp.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON/YAML deserialization.
func (fp *Fingerprint) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*fp = parsed
	return nil
}
