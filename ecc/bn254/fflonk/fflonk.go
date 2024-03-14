package fflonk

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/kzg"
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

// Commit commits to a list of polynomial by intertwinning them like in the FFT, that is
// returns ∑_{i<t}Pᵢ(Xᵗ)Xⁱ for t polynomials
func Commit(p [][]fr.Element, pk kzg.ProvingKey, nbTasks ...int) (kzg.Digest, error) {
	buf := Fold(p)
	com, err := kzg.Commit(buf, pk, nbTasks...)
	return com, err
}

// Fold returns p folded as in the fft, that is ∑_{i<t}Pᵢ(Xᵗ)Xⁱ
func Fold(p [][]fr.Element) []fr.Element {
	t := len(p)
	sizeResult := 0
	for i := range p {
		if sizeResult < len(p[i]) {
			sizeResult = len(p[i])
		}
	}
	sizeResult = sizeResult*len(p) + len(p) - 1
	buf := make([]fr.Element, sizeResult)
	for i := range p {
		for j := range p[i] {
			buf[j*t+i].Set(&p[i][j])
		}
	}
	return buf
}

func eval(f []fr.Element, x fr.Element) fr.Element {
	var y fr.Element
	for i := len(f) - 1; i >= 0; i-- {
		y.Mul(&y, &x).Add(&y, &f[i])
	}
	return y
}

// returns the t t-th roots of x, return an error if they do not exist in Fr
// func extractRoots(x fr.Element, t int) ([]fr.Element, error) {

// 	// for the t-th roots of x to exist we need
// 	// * t | r-1
// 	// * t² | p - (t-1)
// 	r := fr.Modulus()
// 	tBigInt := big.NewInt(int64(t))
// 	oneBigInt := big.NewInt(1)
// 	var a, b big.Int
// 	a.Sub(r, oneBigInt)
// 	a.Mod(&a, tBigInt)
// 	zeroBigInt := big.NewInt(0)
// 	if a.Cmp(zeroBigInt) != 0 {
// 		return nil, ErrRootsOne
// 	}
// 	a.SetUint64(uint64(t)).Mul(tBigInt, tBigInt)
// 	b.Sub(r, tBigInt).Add(&b, oneBigInt)
// 	a.Mod(&b, &a)
// 	if b.Cmp(zeroBigInt) != 0 {
// 		return nil, ErrRootsX
// 	}

// 	// ᵗ√(x) = x^{(p-1)/t + 1}
// 	var expo big.Int
// 	var tthRoot fr.Element
// 	r = fr.Modulus()
// 	tBigInt = big.NewInt(int64(t))
// 	expo.Sub(r, oneBigInt).
// 		Div(&expo, tBigInt).
// 		Add(&expo, oneBigInt)
// 	tthRoot.Exp(x, &expo)

// 	// compute the t-th roots of 1
// 	r.Sub(r, oneBigInt)
// 	tBigInt.Div(r, tBigInt)
// 	gen := getGenFrStar()
// 	gen.Exp(gen, tBigInt)

// 	res := make([]fr.Element, t)
// 	res[0].Set(&tthRoot)
// 	for i := 1; i < t; i++ {
// 		res[i].Mul(&res[i-1], &gen)
// 	}

// 	return res, nil

// }
