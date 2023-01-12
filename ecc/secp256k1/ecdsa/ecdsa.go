package ecdsa

import (
	"bytes"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
)

// PublicKey represents an ECDSA public key.
type PublicKey struct {
	A secp256k1.G1Affine
}

// PrivateKey represents an ECDSA private key.
type PrivateKey struct {
	PublicKey
	D big.Int
}

// params is the ECDSA data structure
type params struct {
	G secp256k1.G1Affine
	N *big.Int
}

// NewParams defines a new params data structure
func NewParams(g secp256k1.G1Affine) params {
	var pp params
	pp.G = g
	pp.N = fr.Modulus()
	return pp
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
	priv.D = k
	priv.PublicKey.A.ScalarMultiplication(&pp.G, &k)
	return priv, nil
}

// Sign performs the ECDSA signature
func (pp params) Sign(hashed *big.Int, privateKey *big.Int, rand io.Reader) ([2]*big.Int, error) {
	k, err := randFieldElement(rand)
	if err != nil {
		return [2]*big.Int{}, err
	}

	kCopy := new(big.Int).SetBytes(k.Bytes())
	var p secp256k1.G1Affine
	p.ScalarMultiplication(&pp.G, kCopy)
	inv := new(big.Int).ModInverse(&k, pp.N)

	privateKeyCopy := new(big.Int).SetBytes(privateKey.Bytes())
	var _x big.Int
	xPrivateKey := new(big.Int).Mul(p.X.BigInt(&_x), privateKeyCopy)

	sum := new(big.Int).Add(hashed, xPrivateKey)

	a := new(big.Int).Mul(inv, sum)
	r2 := new(big.Int).Mod(a, pp.N)
	return [2]*big.Int{&_x, r2}, nil
}

// Verify validates the ECDSA signature
func (pp params) Verify(hashed *big.Int, sig [2]*big.Int, publicKey secp256k1.G1Affine) bool {
	w := new(big.Int).ModInverse(sig[1], pp.N)
	wCopy := new(big.Int).SetBytes(w.Bytes())
	u1raw := new(big.Int).Mul(hashed, wCopy)
	u1 := new(big.Int).Mod(u1raw, pp.N)
	wCopy = new(big.Int).SetBytes(w.Bytes())
	u2raw := new(big.Int).Mul(sig[0], wCopy)
	u2 := new(big.Int).Mod(u2raw, pp.N)

	var gU1, publicKeyU2, p secp256k1.G1Affine
	gU1.ScalarMultiplication(&pp.G, u1)
	publicKeyU2.ScalarMultiplication(&publicKey, u2)

	p.Add(&gU1, &publicKeyU2)

	var _x big.Int
	pXmodN := new(big.Int).Mod(p.X.BigInt(&_x), pp.N)
	return bytes.Equal(pXmodN.Bytes(), sig[0].Bytes())
}
