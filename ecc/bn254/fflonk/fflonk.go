package fflonk

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

var (
	ErrRootsOne = errors.New("Fr does not contain all the t-th roots of 1")
	ErrRootsX   = errors.New("Fr does not contain all the t-th roots of the input")
)

// utils

// getGenFrStar returns a generator of Fr^{*}
func getGenFrStar() fr.Element {
	var res fr.Element
	res.SetUint64(5)
	return res
}

// returns the t t-th roots of x, return an error if they do not exist in Fr
func extractRoots(x fr.Element, t int) ([]fr.Element, error) {

	// for the t-th roots of x to exist we need
	// * t | r-1
	// * t² | p - (t-1)
	r := fr.Modulus()
	tBigInt := big.NewInt(int64(t))
	oneBigInt := big.NewInt(1)
	var a, b big.Int
	a.Sub(r, oneBigInt)
	a.Mod(&a, tBigInt)
	zeroBigInt := big.NewInt(0)
	if a.Cmp(zeroBigInt) != 0 {
		return nil, ErrRootsOne
	}
	a.SetUint64(uint64(t)).Mul(tBigInt, tBigInt)
	b.Sub(r, tBigInt).Add(&b, oneBigInt)
	a.Mod(&b, &a)
	if b.Cmp(zeroBigInt) != 0 {
		return nil, ErrRootsX
	}

	// ᵗ√(x) = x^{(p-1)/t + 1}
	var expo big.Int
	var tthRoot fr.Element
	r = fr.Modulus()
	tBigInt = big.NewInt(int64(t))
	expo.Sub(r, oneBigInt).
		Div(&expo, tBigInt).
		Add(&expo, oneBigInt)
	tthRoot.Exp(x, &expo)

	// compute the t-th roots of 1
	r.Sub(r, oneBigInt)
	tBigInt.Div(r, tBigInt)
	gen := getGenFrStar()
	gen.Exp(gen, tBigInt)

	res := make([]fr.Element, t)
	res[0].Set(&tthRoot)
	for i := 1; i < t; i++ {
		res[i].Mul(&res[i-1], &gen)
	}

	return res, nil

}
