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

// go/src/spxhash/hash/params.go
package spxhash

// SIPS-0001 https://github.com/quantix/sips/wiki/SIPS-0001

// Define prime constants for hash calculations.
const (
	prime32  = 0x9e3779b9         // Example prime constant for 32-bit hash
	prime64  = 0x9e3779b97f4a7c15 // Example prime constant for 64-bit hash
	saltSize = 16                 // Size of salt in bytes (128 bits = 16 bytes)

	// Argon2 parameters (tuned for blockchain hashing performance, not password storage)
	// Note: argon2.IDKey memory parameter is in KiB.
	// These values are intentionally low for a hash function used in block headers;
	// this is not password storage — security here comes from chain structure, not KDF hardness.
	memory           = 64   // Memory cost: 64 KiB (fast for blockchain hashing)
	iterations       = 1    // Number of iterations for Argon2id
	parallelism      = 1         // Degree of parallelism set to 1
	tagSize          = 32        // Tag size set to 256 bits (32 bytes)
	DefaultCacheSize = 100       // Default cache size for SphinxHash
)

// Size returns the number of bytes in the hash based on the bit size.
func (s *SphinxHash) Size() int {
	switch s.bitSize {
	case 256:
		// SHA-512/256 produces a 256-bit output, equivalent to 32 bytes.
		return 32 // 256 bits = 32 bytes (SHA-512/256)
	case 384:
		// SHA-384 produces a 384-bit output, equivalent to 48 bytes.
		return 48 // 384 bits = 48 bytes (SHA-384)
	case 512:
		// SHA-512 produces a 512-bit output, equivalent to 64 bytes.
		return 64 // 512 bits = 64 bytes (SHA-512)
	default:
		// Default to 256 bits (SHA-512/256) if bitSize is unspecified
		return 32 // Default to 256 bits (SHA-512/256)
	}
}

// BlockSize returns the hash block size based on the current bit size configuration.
func (s *SphinxHash) BlockSize() int {
	switch s.bitSize {
	case 256:
		return 128 // SHA-512/256 block size is 128 bytes
	case 384:
		return 128 // SHA-384 block size is 128 bytes
	case 512:
		return 128 // SHA-512 block size is 128 bytes
	default:
		return 136 // SHAKE256 block size is 136 bytes (1088 bits)
	}
}
