// adapted from "crypto/ecdsa"

package secp256k1

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"errors"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
)

// PublicKey represents an ECDSA public key.
type PublicKey struct {
	A G1Affine
}

// PrivateKey represents an ECDSA private key.
type PrivateKey struct {
	PublicKey
	D *big.Int
}

const (
	aesIV = "IV for ECDSA CTR"
)

var one = new(big.Int).SetInt64(1)

// randFieldElement returns a random element of the order of the given
// curve using the procedure given in FIPS 186-4, Appendix B.5.1.
func randFieldElement(rand io.Reader) (k *big.Int, err error) {
	b := make([]byte, fr.Bytes+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return

	}

	k = new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(fr.Modulus(), one)
	k.Mod(k, n)
	k.Add(k, one)
	return

}

// GenerateKey generates a public and private key pair.
func GenerateKey(rand io.Reader) (*PrivateKey, error) {

	k, err := randFieldElement(rand)
	if err != nil {
		return nil, err

	}

	priv := new(PrivateKey)
	priv.D = k
	priv.PublicKey.A.ScalarMultiplication(&g1GenAff, k)
	return priv, nil

}

// hashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func hashToInt(hash []byte) *big.Int {
	orderBits := fr.Bits
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]

	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))

	}
	return ret

}

var errZeroParam = errors.New("zero parameter")

// Sign signs a hash (which should be the result of hashing a larger message)
// using the private key, priv. If the hash is longer than the bit-length of the
// private key's curve order, the hash will be truncated to that length. It
// returns the signature as a pair of integers. Most applications should use
// SignASN1 instead of dealing directly with r, s.
func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	// This implementation derives the nonce from an AES-CTR CSPRNG keyed by:
	//
	//    SHA2-512(priv.D || entropy || hash)[:32]
	//
	// The CSPRNG key is indifferentiable from a random oracle as shown in
	// [Coron], the AES-CTR stream is indifferentiable from a random oracle
	// under standard cryptographic assumptions (see [Larsson] for examples).
	//
	// [Coron]: https://cs.nyu.edu/~dodis/ps/merkle.pdf
	// [Larsson]: https://web.archive.org/web/20040719170906/https://www.nada.kth.se/kurser/kth/2D1441/semteo03/lecturenotes/assump.pdf

	// Get 256 bits of entropy from rand.
	entropy := make([]byte, 32)
	_, err = io.ReadFull(rand, entropy)
	if err != nil {
		return

	}

	// Initialize an SHA-512 hash context; digest...
	md := sha512.New()
	md.Write(priv.D.Bytes()) // the private key,
	md.Write(entropy)        // the entropy,
	md.Write(hash)           // and the input hash;
	key := md.Sum(nil)[:32]  // and compute ChopMD-256(SHA-512),
	// which is an indifferentiable MAC.

	// Create an AES-CTR instance to use as a CSPRNG.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err

	}

	// Create a CSPRNG that xors a stream of zeros with
	// the output of the AES-CTR instance.
	csprng := &cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	return sign(priv, csprng, hash)

}

func sign(priv *PrivateKey, csprng *cipher.StreamReader, hash []byte) (r, s *big.Int, err error) {
	// SEC 1, Version 2.0, Section 4.1.3
	N := fr.Modulus()
	if N.Sign() == 0 {
		return nil, nil, errZeroParam

	}
	var k, kInv *big.Int
	for {
		for {
			k, err = randFieldElement(*csprng)
			if err != nil {
				r = nil
				return

			}

			var _kInv fr.Element
			_kInv.SetBigInt(k).
				Inverse(&_kInv).
				BigInt(kInv)

			var R G1Affine
			R.ScalarMultiplication(&g1GenAff, k)
			R.X.BigInt(r)

			r.Mod(r, N)
			if r.Sign() != 0 {
				break

			}

		}

		e := hashToInt(hash)
		s = new(big.Int).Mul(priv.D, r)
		s.Add(s, e)
		s.Mul(s, kInv)
		s.Mod(s, N) // N != 0
		if s.Sign() != 0 {
			break

		}

	}

	return

}

// Verify verifies the signature in r, s of hash using the public key, pub. Its
// return value records whether the signature is valid.
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {

	N := fr.Modulus()

	if r.Sign() <= 0 || s.Sign() <= 0 {
		return false

	}
	if r.Cmp(N) >= 0 || s.Cmp(N) >= 0 {
		return false

	}
	return verify(pub, hash, r, s)

}

func verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	// SEC 1, Version 2.0, Section 4.1.4
	e := hashToInt(hash)
	var w *big.Int
	N := fr.Modulus()

	var _sInv fr.Element
	_sInv.SetBigInt(s).
		Inverse(&_sInv).
		BigInt(w)

	u1 := e.Mul(e, w)
	u1.Mod(u1, N)
	u2 := w.Mul(r, w)
	u2.Mod(u2, N)

	var P1, P2 G1Affine
	P1.ScalarMultiplication(&g1GenAff, u1)
	P2.ScalarMultiplication(&pub.A, u2).
		Add(&P2, &P1)

	return P2.X.BigInt(w).Cmp(r) == 0

}

type zr struct{}

// Read replaces the contents of dst with zeros. It is safe for concurrent use.
func (zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0

	}
	return len(dst), nil

}

var zeroReader = zr{}
