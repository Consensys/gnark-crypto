package plookup

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
)

var (
	ErrNotInTable          = errors.New("some value in the vector is not in the lookup table")
	ErrPlookupVerification = errors.New("plookup verification failed")
	ErrIncompatibleSize    = errors.New("the tables in f and t are not of the same size")
	ErrFoldedCommitment    = errors.New("the folded commitment is malformed")
)

type Table []fr.Element

// Len is the number of elements in the collection.
func (t Table) Len() int {
	return len(t)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (t Table) Less(i, j int) bool {
	return t[i].Cmp(&t[j]) == -1
}

// Swap swaps the elements with indexes i and j.
func (t Table) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Proof Plookup proof, containing opening proofs
type ProofLookupVector struct {

	// size of the system
	size uint64

	// Commitments to h1, h2, t, z, f, h
	h1, h2, t, z, f, h kzg.Digest

	// Batch opening proof of h1, h2, z, t
	BatchedProof kzg.BatchOpeningProof

	// Batch opening proof of h1, h2, z shifted by g
	BatchedProofShifted kzg.BatchOpeningProof
}

// ProofLookupTables proofs that a list of tables
type ProofLookupTables struct {

	// commitments to the rows f and t
	fs []kzg.Digest
	ts []kzg.Digest

	// lookup proof for the f and t folded
	foldedProof ProofLookupVector
}

// computeZ computes Z, in Lagrange basis. Z is the accumulation of the partial
// ratios of 2 fully split polynomials (cf https://eprint.iacr.org/2020/315.pdf)
// * lf is the list of values that should be in lt
// * lt is the lookup table
// * lh1, lh2 is lf sorted by lt split in 2 overlapping slices
// * beta, gamma are challenges (Schwartz-zippel: they are the random evaluations point)
func computeZ(lf, lt, lh1, lh2 []fr.Element, beta, gamma fr.Element) []fr.Element {

	z := make([]fr.Element, len(lt))

	n := len(lt)
	d := make([]fr.Element, n-1)
	var u, c fr.Element
	c.SetOne().
		Add(&c, &beta).
		Mul(&c, &gamma)
	for i := 0; i < n-1; i++ {

		d[i].Mul(&beta, &lh1[i+1]).
			Add(&d[i], &lh1[i]).
			Add(&d[i], &c)

		u.Mul(&beta, &lh2[i+1]).
			Add(&u, &lh2[i]).
			Add(&u, &c)

		d[i].Mul(&d[i], &u)
	}
	d = fr.BatchInvert(d)

	z[0].SetOne()
	var a, b, e fr.Element
	e.SetOne().Add(&e, &beta)
	for i := 0; i < n-1; i++ {

		a.Add(&gamma, &lf[i])

		b.Mul(&beta, &lt[i+1]).
			Add(&b, &lt[i]).
			Add(&b, &c)

		a.Mul(&a, &b).
			Mul(&a, &e)

		z[i+1].Mul(&z[i], &a).
			Mul(&z[i+1], &d[i])
	}

	return z
}

// computeH computes the evaluation (shifted, bit reversed) of h where
// h = (x-1)*z*(1+beta)*(gamma+f)*(gamma(1+beta) + t+ beta*t(gX)) -
//		(x-1)*z(gX)*(gamma(1+beta) + h1 + beta*h1(gX))*(gamma(1+beta) + h2 + beta*h2(gX) )
//
// * cz, ch1, ch2, ct, cf are the polynomials z, h1, h2, t, f in canonical basis
// * _lz, _lh1, _lh2, _lt, _lf are the polynomials z, h1, h2, t, f in shifted Lagrange basis (domainH)
// * beta, gamma are the challenges
// * it returns h in canonical basis
func computeH(_lz, _lh1, _lh2, _lt, _lf []fr.Element, beta, gamma fr.Element, domainH *fft.Domain) []fr.Element {

	// result
	s := int(domainH.Cardinality)
	num := make([]fr.Element, domainH.Cardinality)

	var u, v, w, _g, m, n, one, t fr.Element
	t.SetUint64(2).
		Inverse(&t)
	_g.Square(&domainH.Generator).
		Exp(_g, big.NewInt(int64(s/2-1)))
	one.SetOne()
	v.Add(&one, &beta)
	w.Mul(&v, &gamma)

	// var d [2]fr.Element
	// d[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality>>1)))
	// d[1].Neg(&d[0])
	// d[0].Sub(&d[0], &one).Inverse(&d[0])
	// d[1].Sub(&d[1], &one).Inverse(&d[1])

	g := make([]fr.Element, s)
	g[0].Set(&domainH.FinerGenerator)
	for i := 1; i < s; i++ {
		g[i].Mul(&g[i-1], &domainH.Generator)
	}

	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < s; i++ {

		_i := int(bits.Reverse64(uint64(i)) >> nn)
		_is := int(bits.Reverse64(uint64((i+2)%s)) >> nn)

		// m = (x-g**(n-1))*z*(1+beta)*(gamma+f)*(gamma(1+beta) + t+ beta*t(gX))
		m.Mul(&v, &_lz[_i])
		u.Add(&gamma, &_lf[_i])
		m.Mul(&m, &u)
		u.Mul(&beta, &_lt[_is]).
			Add(&u, &_lt[_i]).
			Add(&u, &w)
		m.Mul(&m, &u)

		// n = (x-g**(n-1))*z(gX)*(gamma(1+beta) + h1 + beta*h1(gX))*(gamma(1+beta) + h2 + beta*h2(gX)
		n.Mul(&beta, &_lh1[_is]).
			Add(&n, &_lh1[_i]).
			Add(&n, &w)
		u.Mul(&beta, &_lh2[_is]).
			Add(&u, &_lh2[_i]).
			Add(&u, &w)
		n.Mul(&n, &u).
			Mul(&n, &_lz[_is])

		num[_i].Sub(&m, &n)
		u.Sub(&g[i], &_g)
		num[_i].Mul(&num[_i], &u)

	}

	return num
}

// computeH0 returns l0 * (z-1), in Lagrange basis and bit reversed order
func computeH0(lzCosetReversed []fr.Element, domainH *fft.Domain) []fr.Element {

	var one fr.Element
	one.SetOne()

	var g [2]fr.Element
	g[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality/2)))
	g[1].Neg(&g[0])
	g[0].Sub(&g[0], &one)
	g[1].Sub(&g[1], &one)

	var d fr.Element
	d.Set(&domainH.FinerGenerator)
	den := make([]fr.Element, len(lzCosetReversed))
	for i := 0; i < len(den); i++ {
		den[i].Sub(&d, &one)
		d.Mul(&d, &domainH.Generator)
	}
	den = fr.BatchInvert(den)

	res := make([]fr.Element, len(lzCosetReversed))
	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < len(lzCosetReversed); i++ {
		_i := int(bits.Reverse64(uint64(i)) >> nn)
		res[_i].Sub(&lzCosetReversed[_i], &one).
			Mul(&res[_i], &g[i%2]).Mul(&res[_i], &den[i])
	}

	return res
}

// computeHn returns ln * (z-1), in Lagrange basis and bit reversed order
func computeHn(lzCosetReversed []fr.Element, domainH *fft.Domain) []fr.Element {

	var one fr.Element
	one.SetOne()

	var g [2]fr.Element
	g[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality/2)))
	g[1].Neg(&g[0])
	g[0].Sub(&g[0], &one)
	g[1].Sub(&g[1], &one)

	var _g, d fr.Element
	one.SetOne()
	d.Set(&domainH.FinerGenerator)
	_g.Square(&domainH.Generator).Exp(_g, big.NewInt(int64(domainH.Cardinality/2-1)))
	den := make([]fr.Element, len(lzCosetReversed))
	for i := 0; i < len(lzCosetReversed); i++ {
		den[i].Sub(&d, &_g)
		d.Mul(&d, &domainH.Generator)
	}
	den = fr.BatchInvert(den)

	res := make([]fr.Element, len(lzCosetReversed))
	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < len(lzCosetReversed); i++ {
		_i := int(bits.Reverse64(uint64(i)) >> nn)
		res[_i].Sub(&lzCosetReversed[_i], &one).
			Mul(&res[_i], &g[i%2]).
			Mul(&res[_i], &den[i])
	}

	return res
}

// computeHh1h2 returns ln * (h1 - h2(g.x)), in Lagrange basis and bit reversed order
func computeHh1h2(_lh1, _lh2 []fr.Element, domainH *fft.Domain) []fr.Element {

	var one fr.Element
	one.SetOne()

	var g [2]fr.Element
	g[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality/2)))
	g[1].Neg(&g[0])
	g[0].Sub(&g[0], &one)
	g[1].Sub(&g[1], &one)

	var _g, d fr.Element
	d.Set(&domainH.FinerGenerator)
	_g.Square(&domainH.Generator).Exp(_g, big.NewInt(int64(domainH.Cardinality/2-1)))
	den := make([]fr.Element, len(_lh1))
	for i := 0; i < len(_lh1); i++ {
		den[i].Sub(&d, &_g)
		d.Mul(&d, &domainH.Generator)
	}
	den = fr.BatchInvert(den)

	res := make([]fr.Element, len(_lh1))
	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	s := len(_lh1)
	for i := 0; i < s; i++ {

		_i := int(bits.Reverse64(uint64(i)) >> nn)
		_is := int(bits.Reverse64(uint64((i+2)%s)) >> nn)

		res[_i].Sub(&_lh1[_i], &_lh2[_is]).
			Mul(&res[_i], &g[i%2]).
			Mul(&res[_i], &den[i])
	}

	return res
}

// computeQuotient computes the full quotient of the plookup protocol.
// * alpha is the challenge to fold the numerator
// * lh, lh0, lhn, lh1h2 are the various pieces of the numerator (Lagrange shifted form, bit reversed order)
// * domainH fft domain
// It returns the quotient, in canonical basis
func computeQuotient(alpha fr.Element, lh, lh0, lhn, lh1h2 []fr.Element, domainH *fft.Domain) []fr.Element {

	s := len(lh)
	res := make([]fr.Element, s)

	var one fr.Element
	one.SetOne()

	var d [2]fr.Element
	d[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality>>1)))
	d[1].Neg(&d[0])
	d[0].Sub(&d[0], &one).Inverse(&d[0])
	d[1].Sub(&d[1], &one).Inverse(&d[1])

	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < s; i++ {

		_i := int(bits.Reverse64(uint64(i)) >> nn)

		res[_i].Mul(&lh1h2[_i], &alpha).
			Add(&res[_i], &lhn[_i]).
			Mul(&res[_i], &alpha).
			Add(&res[_i], &lh0[_i]).
			Mul(&res[_i], &alpha).
			Add(&res[_i], &lh[_i]).
			Mul(&res[_i], &d[i%2])
	}

	domainH.FFTInverse(res, fft.DIT, 1)

	return res
}

// Prove returns proof that the values in f are in t.
func ProveLookupVector(srs *kzg.SRS, f, t Table) (ProofLookupVector, error) {

	// res
	var proof ProofLookupVector
	var err error

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "beta", "gamma", "alpha", "nu")

	// create domains
	var dNum *fft.Domain
	if len(t) <= len(f) {
		dNum = fft.NewDomain(uint64(len(f)+1), 0, false)
	} else {
		dNum = fft.NewDomain(uint64(len(t)), 0, false)
	}
	cardDNum := int(dNum.Cardinality)

	// set the size
	proof.size = dNum.Cardinality

	// sort f and t
	// sort.Sort(f)
	// sort.Sort(t)

	// resize f and t
	// note: the last element of lf does not matter
	lf := make([]fr.Element, cardDNum)
	lt := make([]fr.Element, cardDNum)
	cf := make([]fr.Element, cardDNum)
	ct := make([]fr.Element, cardDNum)
	copy(lt, t)
	copy(lf, f)
	for i := len(f); i < cardDNum; i++ {
		lf[i] = f[len(f)-1]
		lt[i] = t[len(t)-1]
	}
	copy(ct, lt)
	copy(cf, lf)
	dNum.FFTInverse(ct, fft.DIF, 0)
	dNum.FFTInverse(cf, fft.DIF, 0)
	fft.BitReverse(ct)
	fft.BitReverse(cf)
	proof.t, err = kzg.Commit(ct, srs)
	if err != nil {
		return proof, err
	}
	proof.f, err = kzg.Commit(cf, srs)
	if err != nil {
		return proof, err
	}

	// write f sorted by t
	lfSortedByt := make(Table, 2*dNum.Cardinality-1)
	copy(lfSortedByt, lt)
	copy(lfSortedByt[dNum.Cardinality:], lf)
	sort.Sort(lfSortedByt)

	// compute h1, h2, commit to them
	lh1 := make([]fr.Element, cardDNum)
	lh2 := make([]fr.Element, cardDNum)
	ch1 := make([]fr.Element, cardDNum)
	ch2 := make([]fr.Element, cardDNum)
	copy(lh1, lfSortedByt[:cardDNum])
	copy(lh2, lfSortedByt[cardDNum-1:])

	copy(ch1, lfSortedByt[:cardDNum])
	copy(ch2, lfSortedByt[cardDNum-1:])
	dNum.FFTInverse(ch1, fft.DIF, 0)
	dNum.FFTInverse(ch2, fft.DIF, 0)
	fft.BitReverse(ch1)
	fft.BitReverse(ch2)

	proof.h1, err = kzg.Commit(ch1, srs)
	if err != nil {
		return proof, err
	}
	proof.h2, err = kzg.Commit(ch2, srs)
	if err != nil {
		return proof, err
	}

	// derive beta, gamma
	beta, err := deriveRandomness(&fs, "beta", &proof.t, &proof.f, &proof.h1, &proof.h2)
	if err != nil {
		return proof, err
	}
	gamma, err := deriveRandomness(&fs, "gamma")
	if err != nil {
		return proof, err
	}

	// Compute to Z
	lz := computeZ(lf, lt, lh1, lh2, beta, gamma)
	cz := make([]fr.Element, len(lz))
	copy(cz, lz)
	dNum.FFTInverse(cz, fft.DIF, 0)
	fft.BitReverse(cz)
	proof.z, err = kzg.Commit(cz, srs)
	if err != nil {
		return proof, err
	}

	// prepare data for computing the quotient
	// compute the numerator
	s := dNum.Cardinality
	domainH := fft.NewDomain(uint64(2*s), 1, false)
	_lz := make([]fr.Element, 2*s)
	_lh1 := make([]fr.Element, 2*s)
	_lh2 := make([]fr.Element, 2*s)
	_lt := make([]fr.Element, 2*s)
	_lf := make([]fr.Element, 2*s)
	copy(_lz, cz)
	copy(_lh1, ch1)
	copy(_lh2, ch2)
	copy(_lt, ct)
	copy(_lf, cf)
	domainH.FFT(_lz, fft.DIF, 1)
	domainH.FFT(_lh1, fft.DIF, 1)
	domainH.FFT(_lh2, fft.DIF, 1)
	domainH.FFT(_lt, fft.DIF, 1)
	domainH.FFT(_lf, fft.DIF, 1)

	// compute h
	lh := computeH(_lz, _lh1, _lh2, _lt, _lf, beta, gamma, domainH)

	// compute h0
	lh0 := computeH0(_lz, domainH)

	// compute hn
	lhn := computeHn(_lz, domainH)

	// compute hh1h2
	lh1h2 := computeHh1h2(_lh1, _lh2, domainH)

	// compute the quotient
	alpha, err := deriveRandomness(&fs, "alpha", &proof.z)
	if err != nil {
		return proof, err
	}
	ch := computeQuotient(alpha, lh, lh0, lhn, lh1h2, domainH)
	proof.h, err = kzg.Commit(ch, srs)
	if err != nil {
		return proof, err
	}

	// build the opening proofs
	nu, err := deriveRandomness(&fs, "nu", &proof.h)
	if err != nil {
		return proof, err
	}
	proof.BatchedProof, err = kzg.BatchOpenSinglePoint(
		[]polynomial.Polynomial{
			ch1,
			ch2,
			ct,
			cz,
			cf,
			ch,
		},
		[]kzg.Digest{
			proof.h1,
			proof.h2,
			proof.t,
			proof.z,
			proof.f,
			proof.h,
		},
		&nu,
		hFunc,
		dNum,
		srs,
	)
	if err != nil {
		return proof, err
	}

	nu.Mul(&nu, &dNum.Generator)
	proof.BatchedProofShifted, err = kzg.BatchOpenSinglePoint(
		[]polynomial.Polynomial{
			ch1,
			ch2,
			ct,
			cz,
		},
		[]kzg.Digest{
			proof.h1,
			proof.h2,
			proof.t,
			proof.z,
		},
		&nu,
		hFunc,
		dNum,
		srs,
	)
	if err != nil {
		return proof, err
	}

	return proof, nil

}

// ProveLookupTables generates a proof that f, seen as a multi dimensional table,
// consists of vectors that are in t. In other words for each i, f[:][i] must be one
// of the t[:][j].
//
// For instance, if t is the truth table of the XOR function, t will be populated such
// that t[:][i] contains the i-th entry of the truth table, so t[0][i] XOR t[1][i] = t[2][i].
//
// The Table in f and t are supposed to be of the same size constant size.
func ProveLookupTables(srs *kzg.SRS, f, t []Table) (ProofLookupTables, error) {

	// res
	proof := ProofLookupTables{}
	var err error

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// check the sizes
	if len(f) != len(t) {
		return proof, ErrIncompatibleSize
	}
	s := len(f[0])
	for i := 1; i < len(f); i++ {
		if len(f[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}
	s = len(t[0])
	for i := 1; i < len(t); i++ {
		if len(t[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}

	// commit to the tables in f and t
	sizeTable := len(t)
	proof.fs = make([]kzg.Digest, sizeTable)
	proof.ts = make([]kzg.Digest, sizeTable)
	m := len(f[0]) + 1
	if m < len(t[0]) {
		m = len(t[0])
	}
	d := fft.NewDomain(uint64(m), 0, false)
	lfs := make([][]fr.Element, sizeTable)
	lts := make([][]fr.Element, sizeTable)
	cfs := make([][]fr.Element, sizeTable)
	cts := make([][]fr.Element, sizeTable)

	for i := 0; i < sizeTable; i++ {

		cfs[i] = make([]fr.Element, d.Cardinality)
		lfs[i] = make([]fr.Element, d.Cardinality)
		copy(cfs[i], f[i])
		copy(lfs[i], f[i])
		for j := len(f[i]); j < int(d.Cardinality); j++ {
			cfs[i][j] = f[i][len(f[i])-1]
			lfs[i][j] = f[i][len(f[i])-1]
		}
		d.FFTInverse(cfs[i], fft.DIF, 0)
		fft.BitReverse(cfs[i])
		proof.fs[i], err = kzg.Commit(cfs[i], srs)
		if err != nil {
			return proof, err
		}

		cts[i] = make([]fr.Element, d.Cardinality)
		lts[i] = make([]fr.Element, d.Cardinality)
		copy(cts[i], t[i])
		copy(lts[i], t[i])
		for j := len(t[i]); j < int(d.Cardinality); j++ {
			cts[i][j] = t[i][len(t[i])-1]
			lts[i][j] = t[i][len(t[i])-1]
		}
		d.FFTInverse(cts[i], fft.DIF, 0)
		fft.BitReverse(cts[i])
		proof.ts[i], err = kzg.Commit(cts[i], srs)
		if err != nil {
			return proof, err
		}
	}

	// fold f and t
	comms := make([]*kzg.Digest, 2*sizeTable)
	for i := 0; i < sizeTable; i++ {
		comms[i] = new(kzg.Digest)
		comms[sizeTable+i] = new(kzg.Digest)
		comms[i].Set(&proof.fs[i])
		comms[sizeTable+i].Set(&proof.ts[i])
	}
	lambda, err := deriveRandomness(&fs, "lambda", comms...)
	if err != nil {
		return proof, err
	}
	// lambda.SetUint64(238293208029)
	lambda.SetString("1535610991669198651944444444444444444444")
	fmt.Printf("lambda (prover):   %s\n", lambda.String())
	foldedf := make(Table, d.Cardinality)
	foldedt := make(Table, d.Cardinality)
	for i := 0; i < len(cfs[0]); i++ {
		for j := sizeTable - 1; j >= 0; j-- {
			foldedf[i].Mul(&foldedf[i], &lambda).
				Add(&foldedf[i], &lfs[j][i])
			foldedt[i].Mul(&foldedt[i], &lambda).
				Add(&foldedt[i], &lts[j][i])
		}
	}

	// call plookupVector, on foldedf[:len(foldedf)-1] to ensure that the domain size
	// in ProveLookupVector is the same as d's
	fmt.Println("folded f")
	for i := 0; i < len(foldedf)-1; i++ {
		fmt.Printf("fvector[%d].SetString(\"%s\")\n", i, foldedf[i].String())
	}
	fmt.Println("")
	fmt.Println("folded t")
	for i := 0; i < len(foldedf)-1; i++ {
		fmt.Printf("lookupVector[%d].SetString(\"%s\")\n", i, foldedt[i].String())
	}
	proof.foldedProof, err = ProveLookupVector(srs, foldedf[:len(foldedf)-1], foldedt)

	return proof, err
}

// VerifyLookupTables verifies that a ProofLookupTables proof is correct.
func VerifyLookupTables(srs *kzg.SRS, proof ProofLookupTables) error {

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// fold the commitments
	sizeTable := len(proof.fs)
	comms := make([]*kzg.Digest, 2*sizeTable)
	for i := 0; i < sizeTable; i++ {
		comms[i] = &proof.fs[i]
		comms[sizeTable+i] = &proof.ts[i]
	}
	lambda, err := deriveRandomness(&fs, "lambda", comms...)
	if err != nil {
		return err
	}
	// lambda.SetUint64(238293208029)
	lambda.SetString("1535610991669198651944444444444444444444")
	fmt.Printf("lambda (verifier): %s\n", lambda.String())

	// verify that the commitments in the inner proof are consistant
	// with the folded commitments.
	var comt, comf kzg.Digest
	comf.Set(&proof.fs[sizeTable-1])
	comt.Set(&proof.ts[sizeTable-1])
	var blambda big.Int
	lambda.ToBigIntRegular(&blambda)
	for i := sizeTable - 2; i >= 0; i-- {
		comf.ScalarMultiplication(&comf, &blambda).
			Add(&comf, &proof.fs[i])
		comt.ScalarMultiplication(&comt, &blambda).
			Add(&comt, &proof.ts[i])
	}

	if !comf.Equal(&proof.foldedProof.f) {
		return ErrFoldedCommitment
	}
	if !comt.Equal(&proof.foldedProof.t) {
		return ErrFoldedCommitment
	}

	// verify the inner proof
	return VerifyLookupVector(srs, proof.foldedProof)
}

// VerifyLookupVector verifies that a plookup proof is correct
func VerifyLookupVector(srs *kzg.SRS, proof ProofLookupVector) error {

	// hash function that is used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "beta", "gamma", "alpha", "nu")

	// derive the various challenges
	beta, err := deriveRandomness(&fs, "beta", &proof.t, &proof.f, &proof.h1, &proof.h2)
	if err != nil {
		return err
	}

	gamma, err := deriveRandomness(&fs, "gamma")
	if err != nil {
		return err
	}

	alpha, err := deriveRandomness(&fs, "alpha", &proof.z)
	if err != nil {
		return err
	}

	nu, err := deriveRandomness(&fs, "nu", &proof.h)
	if err != nil {
		return err
	}

	// check opening proofs
	err = kzg.BatchVerifySinglePoint(
		[]kzg.Digest{
			proof.h1,
			proof.h2,
			proof.t,
			proof.z,
			proof.f,
			proof.h,
		},
		&proof.BatchedProof,
		hFunc,
		srs,
	)
	if err != nil {
		return err
	}

	err = kzg.BatchVerifySinglePoint(
		[]kzg.Digest{
			proof.h1,
			proof.h2,
			proof.t,
			proof.z,
		},
		&proof.BatchedProofShifted,
		hFunc,
		srs,
	)
	if err != nil {
		return err
	}

	// check polynomial relation using Schwartz Zippel
	var lhs, rhs, nun, g, _g, a, v, w, one fr.Element
	d := fft.NewDomain(proof.size, 0, false) // only there to access to root of 1...
	one.SetOne()
	g.Exp(d.Generator, big.NewInt(int64(d.Cardinality-1)))

	// beta.SetUint64(13)
	// gamma.SetUint64(23)

	v.Add(&one, &beta)
	w.Mul(&v, &gamma)

	// h(nu) where
	// h = (x-1)*z*(1+beta)*(gamma+f)*(gamma(1+beta) + t+ beta*t(gX)) -
	//		(x-1)*z(gX)*(gamma(1+beta) + h1 + beta*h1(gX))*(gamma(1+beta) + h2 + beta*h2(gX) )
	lhs.Sub(&nu, &g).
		Mul(&lhs, &proof.BatchedProof.ClaimedValues[3]).
		Mul(&lhs, &v)
	a.Add(&gamma, &proof.BatchedProof.ClaimedValues[4])
	lhs.Mul(&lhs, &a)
	a.Mul(&beta, &proof.BatchedProofShifted.ClaimedValues[2]).
		Add(&a, &proof.BatchedProof.ClaimedValues[2]).
		Add(&a, &w)
	lhs.Mul(&lhs, &a)

	rhs.Sub(&nu, &g).
		Mul(&rhs, &proof.BatchedProofShifted.ClaimedValues[3])
	a.Mul(&beta, &proof.BatchedProofShifted.ClaimedValues[0]).
		Add(&a, &proof.BatchedProof.ClaimedValues[0]).
		Add(&a, &w)
	rhs.Mul(&rhs, &a)
	a.Mul(&beta, &proof.BatchedProofShifted.ClaimedValues[1]).
		Add(&a, &proof.BatchedProof.ClaimedValues[1]).
		Add(&a, &w)
	rhs.Mul(&rhs, &a)

	lhs.Sub(&lhs, &rhs)

	// check consistancy of bounds
	var l0, ln, d1, d2 fr.Element
	l0.Exp(nu, big.NewInt(int64(d.Cardinality))).Sub(&l0, &one)
	ln.Set(&l0)
	d1.Sub(&nu, &one)
	d2.Sub(&nu, &g)
	l0.Div(&l0, &d1)
	ln.Div(&ln, &d2)

	// l0*(z-1)
	var l0z fr.Element
	l0z.Sub(&proof.BatchedProof.ClaimedValues[3], &one).
		Mul(&l0z, &l0)

	// ln*(z-1)
	var lnz fr.Element
	lnz.Sub(&proof.BatchedProof.ClaimedValues[3], &one).
		Mul(&ln, &lnz)

	// ln*(h1 - h2(g.x))
	var lnh1h2 fr.Element
	lnh1h2.Sub(&proof.BatchedProof.ClaimedValues[0], &proof.BatchedProofShifted.ClaimedValues[1]).
		Mul(&lnh1h2, &ln)

	// fold the numerator
	lnh1h2.Mul(&lnh1h2, &alpha).
		Add(&lnh1h2, &lnz).
		Mul(&lnh1h2, &alpha).
		Add(&lnh1h2, &l0z).
		Mul(&lnh1h2, &alpha).
		Add(&lnh1h2, &lhs)

	// (x**n-1) * h(x) evaluated at nu
	nun.Exp(nu, big.NewInt(int64(d.Cardinality)))
	_g.Sub(&nun, &one)
	_g.Mul(&proof.BatchedProof.ClaimedValues[5], &_g)
	if !lnh1h2.Equal(&_g) {
		return ErrPlookupVerification
	}

	return nil
}

// TODO put that in fiat-shamir package
func deriveRandomness(fs *fiatshamir.Transcript, challenge string, points ...*bn254.G1Affine) (fr.Element, error) {

	var buf [bn254.SizeOfG1AffineUncompressed]byte
	var r fr.Element

	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}

	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}
