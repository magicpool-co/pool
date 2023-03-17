// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2022 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package crypto

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"math/big"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	schnorr "github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
)

var (
	// singleZero is used during RFC6979 nonce generation.  It is provided
	// here to avoid the need to create it multiple times.
	singleZero = []byte{0x00}

	// zeroInitializer is used during RFC6979 nonce generation.  It is provided
	// here to avoid the need to create it multiple times.
	zeroInitializer = bytes.Repeat([]byte{0x00}, sha256.BlockSize)

	// singleOne is used during RFC6979 nonce generation.  It is provided
	// here to avoid the need to create it multiple times.
	singleOne = []byte{0x01}

	// oneInitializer is used during RFC6979 nonce generation.  It is provided
	// here to avoid the need to create it multiple times.
	oneInitializer = bytes.Repeat([]byte{0x01}, sha256.Size)
)

// hmacsha256 implements a resettable version of HMAC-SHA256.
type hmacsha256 struct {
	inner, outer hash.Hash
	ipad, opad   [sha256.BlockSize]byte
}

// Write adds data to the running hash.
func (h *hmacsha256) Write(p []byte) {
	h.inner.Write(p)
}

// initKey initializes the HMAC-SHA256 instance to the provided key.
func (h *hmacsha256) initKey(key []byte) {
	// Hash the key if it is too large.
	if len(key) > sha256.BlockSize {
		h.outer.Write(key)
		key = h.outer.Sum(nil)
	}
	copy(h.ipad[:], key)
	copy(h.opad[:], key)
	for i := range h.ipad {
		h.ipad[i] ^= 0x36
	}
	for i := range h.opad {
		h.opad[i] ^= 0x5c
	}
	h.inner.Write(h.ipad[:])
}

// ResetKey resets the HMAC-SHA256 to its initial state and then initializes it
// with the provided key.  It is equivalent to creating a new instance with the
// provided key without allocating more memory.
func (h *hmacsha256) ResetKey(key []byte) {
	h.inner.Reset()
	h.outer.Reset()
	copy(h.ipad[:], zeroInitializer)
	copy(h.opad[:], zeroInitializer)
	h.initKey(key)
}

// Resets the HMAC-SHA256 to its initial state using the current key.
func (h *hmacsha256) Reset() {
	h.inner.Reset()
	h.inner.Write(h.ipad[:])
}

// Sum returns the hash of the written data.
func (h *hmacsha256) Sum() []byte {
	h.outer.Reset()
	h.outer.Write(h.opad[:])
	h.outer.Write(h.inner.Sum(nil))
	return h.outer.Sum(nil)
}

// newHMACSHA256 returns a new HMAC-SHA256 hasher using the provided key.
func newHMACSHA256(key []byte) *hmacsha256 {
	h := new(hmacsha256)
	h.inner = sha256.New()
	h.outer = sha256.New()
	h.initKey(key)
	return h
}

// nonceRFC6979 generates an ECDSA nonce (`k`) deterministically according to RFC 6979.
// It takes a 32-byte hash as an input and returns 32-byte nonce to be used in ECDSA algorithm.
func nonceRFC6979BCH(privKey, hash, extra []byte) *secp256k1.ModNScalar {
	// Input to HMAC is the 32-byte private key and the 32-byte hash.  In
	// addition, it may include the optional 32-byte extra data and 16-byte
	// version.  Create a fixed-size array to avoid extra allocs and slice it
	// properly.
	const (
		privKeyLen = 32
		hashLen    = 32
		extraLen   = 16
	)
	var keyBuf [privKeyLen + hashLen + extraLen]byte

	// Truncate rightmost bytes of private key and hash if they are too long and
	// leave left padding of zeros when they're too short.
	if len(privKey) > privKeyLen {
		privKey = privKey[:privKeyLen]
	}
	if len(hash) > hashLen {
		hash = hash[:hashLen]
	}
	offset := privKeyLen - len(privKey) // Zero left padding if needed.
	offset += copy(keyBuf[offset:], privKey)
	offset += hashLen - len(hash) // Zero left padding if needed.
	offset += copy(keyBuf[offset:], hash)
	key := keyBuf[:offset]

	// Step B.
	//
	// V = 0x01 0x01 0x01 ... 0x01 such that the length of V, in bits, is
	// equal to 8*ceil(hashLen/8).
	//
	// Note that since the hash length is a multiple of 8 for the chosen hash
	// function in this optimized implementation, the result is just the hash
	// length, so avoid the extra calculations.  Also, since it isn't modified,
	// start with a global value.
	v := oneInitializer

	// Step C (Go zeroes all allocated memory).
	//
	// K = 0x00 0x00 0x00 ... 0x00 such that the length of K, in bits, is
	// equal to 8*ceil(hashLen/8).
	//
	// As above, since the hash length is a multiple of 8 for the chosen hash
	// function in this optimized implementation, the result is just the hash
	// length, so avoid the extra calculations.
	k := zeroInitializer[:hashLen]

	// Step D.
	//
	// K = HMAC_K(V || 0x00 || int2octets(x) || bits2octets(h1))
	//
	// Note that key is the "int2octets(x) || bits2octets(h1)" portion along
	// with potential additional data as described by section 3.6 of the RFC.
	hasher := newHMACSHA256(k)
	hasher.Write(v)
	hasher.Write(singleZero)
	hasher.Write(key)
	hasher.Write(extra)
	k = hasher.Sum()

	// Step E.
	//
	// V = HMAC_K(V)
	hasher.ResetKey(k)
	hasher.Write(v)
	v = hasher.Sum()

	// Step F.
	//
	// K = HMAC_K(V || 0x01 || int2octets(x) || bits2octets(h1))
	//
	// Note that key is the "int2octets(x) || bits2octets(h1)" portion along
	// with potential additional data as described by section 3.6 of the RFC.
	hasher.Reset()
	hasher.Write(v)
	hasher.Write(singleOne[:])
	hasher.Write(key[:])
	hasher.Write(extra)
	k = hasher.Sum()

	// Step G.
	//
	// V = HMAC_K(V)
	hasher.ResetKey(k)
	hasher.Write(v)
	v = hasher.Sum()

	// Step H.
	//
	// Repeat until the value is nonzero and less than the curve order.
	for {
		// Step H1 and H2.
		//
		// Set T to the empty sequence.  The length of T (in bits) is denoted
		// tlen; thus, at that point, tlen = 0.
		//
		// While tlen < qlen, do the following:
		//   V = HMAC_K(V)
		//   T = T || V
		//
		// Note that because the hash function output is the same length as the
		// private key in this optimized implementation, there is no need to
		// loop or create an intermediate T.
		hasher.Reset()
		hasher.Write(v)
		v = hasher.Sum()

		// Step H3.
		//
		// k = bits2int(T)
		// If k is within the range [1,q-1], return it.
		//
		// Otherwise, compute:
		// K = HMAC_K(V || 0x00)
		// V = HMAC_K(V)
		var secret secp256k1.ModNScalar
		overflow := secret.SetByteSlice(v)
		if !overflow && !secret.IsZero() {
			return &secret
		}

		// K = HMAC_K(V || 0x00)
		hasher.Reset()
		hasher.Write(v)
		hasher.Write(singleZero[:])
		k = hasher.Sum()

		// V = HMAC_K(V)
		hasher.ResetKey(k)
		hasher.Write(v)
		v = hasher.Sum()
	}
}

// padIntBytes pads a big int bytes with leading zeros if they
// are missing to get the length up to 32 bytes.
func padIntBytes(val *big.Int) []byte {
	b := val.Bytes()
	pad := bytes.Repeat([]byte{0x00}, 32-len(b))
	return append(pad, b...)
}

func SchnorrSignBCH(privKey *secp256k1.PrivateKey, hash []byte) *schnorr.Signature {
	var extraData = []byte{'S', 'c', 'h', 'n', 'o', 'r', 'r', '+', 'S', 'H', 'A', '2', '5', '6', ' ', ' '}
	nonce := nonceRFC6979BCH(privKey.ToECDSA().D.Bytes(), hash, extraData)

	// compute point R = k * G
	var R secp256k1.JacobianPoint
	k := nonce
	secp256k1.ScalarBaseMultNonConst(k, &R)

	// negate nonce if R.y is not a quadratic residue.
	R.ToAffine()
	Ry := new(big.Int).SetBytes(R.Y.Bytes()[:])
	P := privKey.ToECDSA().Params().P
	if big.Jacobi(Ry, P) != 1 {
		k.Negate()
	}

	// compute scalar e = Hash(R.x || compressed(P) || m) mod N
	Rx := new(big.Int).SetBytes(R.X.Bytes()[:])
	commitment := sha256.Sum256(append(append(padIntBytes(Rx), privKey.PubKey().SerializeCompressed()...), hash...))
	var e secp256k1.ModNScalar
	e.SetBytes(&commitment)

	// compute scalar s = (k + e * x) mod N
	s := new(secp256k1.ModNScalar).Mul2(&e, &privKey.Key).Add(k)

	// create the signature
	sig := schnorr.NewSignature(&R.X, s)

	return sig
}

func SchnorrVerifyBCH(sig *schnorr.Signature, pubKey *secp256k1.PublicKey, hash []byte) bool {
	const (
		signatureSize = 64
	)

	serialized := sig.Serialize()
	// The signature must be the correct length.
	sigLen := len(serialized)
	if sigLen < signatureSize || sigLen > signatureSize {
		return false
	}

	// The signature is validly encoded at this point, however, enforce
	// additional restrictions to ensure r is in the range [0, p-1], and s is in
	// the range [0, n-1] since valid Schnorr signatures are required to be in
	// that range per spec.
	var r secp256k1.FieldVal
	if overflow := r.SetByteSlice(serialized[0:32]); overflow {
		return false
	}

	var s secp256k1.ModNScalar
	if overflow := s.SetByteSlice(serialized[32:64]); overflow {
		return false
	}

	// The algorithm for producing a EC-Schnorr-DCRv0 signature is described in
	// README.md and is reproduced here for reference:
	//
	//
	// 1. Fail if m is not 32 bytes
	// 2. Fail if Q is not a point on the curve
	// 3. Fail if r >= p
	// 4. Fail if s >= n
	// 5. e = BLAKE-256(r || m) (Ensure r is padded to 32 bytes)
	// 6. Fail if e >= n
	// 7. R = s*G + e*Q
	// 8. Fail if R is the point at infinity
	// 9. Fail if R.y is odd
	// 10. Verified if R.x == r

	// Step 1.
	//
	// Fail if m is not 32 bytes
	if len(hash) != 32 {
		return false
	}

	// Step 2.
	//
	// Fail if Q is not a point on the curve
	if !pubKey.IsOnCurve() {
		return false
	}

	// Step 3.
	//
	// Fail if r >= p
	//
	// Note this is already handled by the fact r is a field element.

	// Step 4.
	//
	// Fail if s >= n
	//
	// Note this is already handled by the fact s is a mod n scalar.

	// Step 5.
	//
	// compute scalar e = Hash(R.x || compressed(P) || m) mod N
	Rx := new(big.Int).SetBytes(r.Bytes()[:])
	commitment := sha256.Sum256(append(append(padIntBytes(Rx), pubKey.SerializeCompressed()...), hash...))

	// Step 6.
	//
	// Fail if e >= n
	var e secp256k1.ModNScalar
	if overflow := e.SetBytes(&commitment); overflow != 0 {
		return false
	}

	// Negate e
	e.Negate()

	// Step 7.
	//
	// R = s*G + e*Q
	var Q, R, sG, eQ secp256k1.JacobianPoint
	pubKey.AsJacobian(&Q)
	secp256k1.ScalarBaseMultNonConst(&s, &sG)
	secp256k1.ScalarMultNonConst(&e, &Q, &eQ)
	secp256k1.AddNonConst(&sG, &eQ, &R)

	// Step 8.
	//
	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		return false
	}

	// Step 9.
	//
	// Fail if R.y is odd
	//
	// Note that R must be in affine coordinates for this check.
	R.ToAffine()
	P := pubKey.ToECDSA().Params().P
	yz := R.Y.Mul(&R.Z).Normalize()
	b := yz.Bytes()
	if big.Jacobi(new(big.Int).SetBytes(b[:]), P) != 1 {
		return false
	}

	// Step 10.
	//
	// Verified if R.x == r
	//
	// Note that R must be in affine coordinates for this check.
	if !r.Equals(&R.X) {
		return false
	}

	return true
}
