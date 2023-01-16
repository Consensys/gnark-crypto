package ecdsa

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
)

// PublicKey represents an ECDSA public key
type PublicKey struct {
	A secp256k1.G1Affine
}

// PrivateKey represents an ECDSA private key
type PrivateKey struct {
	PublicKey
	Secret big.Int
}

// params are the ECDSA public parameters
type params struct {
	Base  secp256k1.G1Affine
	Order *big.Int
}

var one = new(big.Int).SetInt64(1)

// randFieldElement returns a random element of the order of the given
// curve using the procedure given in FIPS 186-4, Appendix B.5.1.
func randFieldElement(rand io.Reader) (k big.Int, err error) {
	b := make([]byte, fr.Bits/8+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return

	}

	k = *new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(fr.Modulus(), one)
	k.Mod(&k, n)
	k.Add(&k, one)
	return
}

// GenerateKey generates a public and private key pair.
func (pp params) GenerateKey(rand io.Reader) (*PrivateKey, error) {

	k, err := randFieldElement(rand)
	if err != nil {
		return nil, err

	}

	priv := new(PrivateKey)
	priv.Secret = k
	priv.PublicKey.A.ScalarMultiplication(&pp.Base, &k)
	return priv, nil
}

// hashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func hashToInt(hash []byte) big.Int {
	if len(hash) > fr.Bytes {
		hash = hash[:fr.Bytes]

	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - fr.Bits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))

	}
	return *ret
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

const (
	aesIV = "gnark-crypto IV." // must be 16 chars (equal block size)
)

func nonce(rand io.Reader, priv *PrivateKey, hash []byte) (csprng *cipher.StreamReader, err error) {
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
	md.Write(priv.Secret.Bytes()) // the private key,
	md.Write(entropy)             // the entropy,
	md.Write(hash)                // and the input hash;
	key := md.Sum(nil)[:32]       // and compute ChopMD-256(SHA-512),
	// which is an indifferentiable MAC.

	// Create an AES-CTR instance to use as a CSPRNG.
	block, _ := aes.NewCipher(key)

	// Create a CSPRNG that xors a stream of zeros with
	// the output of the AES-CTR instance.
	csprng = &cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	return csprng, err
}

// Sign performs the ECDSA signature
func (pp params) Sign(hash []byte, privateKey PrivateKey, rand io.Reader) ([2]*big.Int, error) {
	csprng, err := nonce(rand, &privateKey, hash)
	if err != nil {
		return [2]*big.Int{}, err
	}
	k, err := randFieldElement(csprng)
	if err != nil {
		return [2]*big.Int{}, err
	}

	kCopy := new(big.Int).SetBytes(k.Bytes())
	var p secp256k1.G1Affine
	p.ScalarMultiplication(&pp.Base, kCopy)
	inv := new(big.Int).ModInverse(&k, pp.Order)

	privateKeyCopy := new(big.Int).SetBytes(privateKey.Secret.Bytes())
	var _x big.Int
	xPrivateKey := new(big.Int).Mul(p.X.BigInt(&_x), privateKeyCopy)

	e := hashToInt(hash)
	sum := new(big.Int).Add(&e, xPrivateKey)

	a := new(big.Int).Mul(inv, sum)
	r2 := new(big.Int).Mod(a, pp.Order)
	return [2]*big.Int{&_x, r2}, nil
}

// Verify validates the ECDSA signature
func (pp params) Verify(hash []byte, sig [2]*big.Int, publicKey secp256k1.G1Affine) bool {
	w := new(big.Int).ModInverse(sig[1], pp.Order)
	wCopy := new(big.Int).SetBytes(w.Bytes())
	e := hashToInt(hash)
	u1raw := new(big.Int).Mul(&e, wCopy)
	u1 := new(big.Int).Mod(u1raw, pp.Order)
	wCopy = new(big.Int).SetBytes(w.Bytes())
	u2raw := new(big.Int).Mul(sig[0], wCopy)
	u2 := new(big.Int).Mod(u2raw, pp.Order)

	var gU1, publicKeyU2, p secp256k1.G1Affine
	gU1.ScalarMultiplication(&pp.Base, u1)
	publicKeyU2.ScalarMultiplication(&publicKey, u2)

	p.Add(&gU1, &publicKeyU2)

	var _x big.Int
	pXmodN := new(big.Int).Mod(p.X.BigInt(&_x), pp.Order)
	return bytes.Equal(pXmodN.Bytes(), sig[0].Bytes())
}
