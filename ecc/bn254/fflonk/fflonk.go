package fflonk

import (
	"errors"
	"hash"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	"github.com/consensys/gnark-crypto/ecc/bn254/shplonk"
)

var (
	ErrRootsOne              = errors.New("Fr does not contain all the t-th roots of 1")
	ErrNbPolynomialsNbPoints = errors.New("the number of packs of polynomials should be the same as the number of pack of points")
)

// Opening fflonk proof for opening a list of list of polynomials ((fʲᵢ)ᵢ)ⱼ where each
// pack of polynomials (fʲᵢ)ᵢ (the pack is indexed by j) is opened on a powers of elements in
// the set (Sʲᵢ)ᵢ (indexed by j), where the power is |(fʲᵢ)ᵢ|.
//
// implements io.ReaderFrom and io.WriterTo
type OpeningProof struct {

	// shplonk opening proof of the folded polynomials
	SOpeningProof shplonk.OpeningProof

	// ClaimedValues ClaimedValues[i][j] contains the values
	// of fʲᵢ on Sⱼ^{|(fʲᵢ)ᵢ|}
	ClaimedValues [][][]fr.Element
}

// CommitAndFold commits to a list of polynomial by intertwinning them like in the FFT, that is
// returns ∑_{i<t}Pᵢ(Xᵗ)Xⁱ for t polynomials
func CommitAndFold(p [][]fr.Element, pk kzg.ProvingKey, nbTasks ...int) (kzg.Digest, error) {
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

// BatchOpen computes a batch opening proof of p (the (fʲᵢ)ᵢ ) on powers of points (the ((Sʲᵢ)ᵢ)ⱼ).
// The j-th pack of polynomials is opened on the power |(fʲᵢ)ᵢ| of (Sʲᵢ)ᵢ.
// digests is the list (CommitAndFold(p[i]))ᵢ. It is assumed that the list has been computed beforehand
// and provided as an input to not duplicate computations.
func BatchOpen(p [][][]fr.Element, digests []kzg.Digest, points [][]fr.Element, hf hash.Hash, pk kzg.ProvingKey, dataTranscript ...[]byte) (OpeningProof, error) {

	var res OpeningProof

	if len(p) != len(points) {
		return res, ErrNbPolynomialsNbPoints
	}

	// step 0: compute the relevant powers of the ((Sʲᵢ)ᵢ)ⱼ)
	nbPolysPerPack := make([]int, len(p))
	for i := 0; i < len(p); i++ {
		nbPolysPerPack[i] = len(p[i])
	}
	pointsPowerM := make([][]fr.Element, len(points))
	var tmpBigInt big.Int
	for i := 0; i < len(p); i++ {
		tmpBigInt.SetUint64(uint64(nbPolysPerPack[i]))
		pointsPowerM[i] = make([]fr.Element, len(points[i]))
		for j := 0; j < len(points[i]); j++ {
			pointsPowerM[i][j].Exp(points[i][j], &tmpBigInt)
		}
	}

	// step 1: compute the claimed values, that is the evaluations of the polynomials
	// on the relevant powers of the sets
	res.ClaimedValues = make([][][]fr.Element, len(p))
	for i := 0; i < len(p); i++ {
		res.ClaimedValues[i] = make([][]fr.Element, len(p[i]))
		for j := 0; j < len(points[i]); j++ {
			res.ClaimedValues[i][j] = make([]fr.Element, len(points[i]))
			for k := 0; k < len(points[i]); k++ {
				res.ClaimedValues[i][j][k] = eval(p[i][j], pointsPowerM[i][k])
			}
		}
	}

	// step 2: fold polynomials
	foldedPolynomials := make([][]fr.Element, len(p))
	for i := 0; i < len(p); i++ {
		foldedPolynomials[i] = Fold(p[i])
	}

	// step 4: compute the associated roots, that is for each point p corresponding
	// to a pack i of polynomials, we extend to <p, ω p, .., ωᵗ⁻¹p> if
	// the i-th pack contains t polynomials where ω is a t-th root of 1
	var omega fr.Element
	zeroBigInt := big.NewInt(0)
	genFrStar := getGenFrStar()
	rMinusOneBigInt := fr.Modulus()
	oneBigInt := big.NewInt(1)
	rMinusOneBigInt.Sub(rMinusOneBigInt, oneBigInt)
	newPoints := make([][]fr.Element, len(points))
	for i := 0; i < len(p); i++ {
		tmpBigInt.SetUint64(uint64(len(p[i])))
		tmpBigInt.Mod(rMinusOneBigInt, &tmpBigInt)
		if tmpBigInt.Cmp(zeroBigInt) != 0 {
			return res, ErrRootsOne
		}
		tmpBigInt.SetUint64(uint64(len(p[i])))
		tmpBigInt.Div(rMinusOneBigInt, &tmpBigInt)
		omega.Exp(genFrStar, &tmpBigInt)
		t := len(p[i])
		newPoints[i] = make([]fr.Element, t*len(points[i]))
		for j := 0; j < len(points[i]); j++ {
			newPoints[i][j*t].Set(&points[i][j])
			for k := 1; k < t; k++ {
				newPoints[i][j*t+k].Mul(&newPoints[i][j*t+k-1], &omega)
			}
		}
	}

	// step 5: shplonk open the list of single polynomials on the new sets
	var err error
	res.SOpeningProof, err = shplonk.BatchOpen(foldedPolynomials, digests, points, hf, pk, dataTranscript...)

	return res, err

}

// utils

// getGenFrStar returns a generator of Fr^{*}
func getGenFrStar() fr.Element {
	var res fr.Element
	res.SetUint64(5)
	return res
}

func eval(f []fr.Element, x fr.Element) fr.Element {
	var y fr.Element
	for i := len(f) - 1; i >= 0; i-- {
		y.Mul(&y, &x).Add(&y, &f[i])
	}
	return y
}

// Open
// func Open(polynomials [][]fr.Element, digests []kzg.Digest, points []fr.Element, hf hash.Hash, pk kzg.ProvingKey, dataTranscript ...[]byte) (shplonk.OpeningProof, error) {

// }

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
