package plookup

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

var (
	ErrNotInTable          = errors.New("some value in the vector is not in the lookup table")
	ErrPlookupVerification = errors.New("plookup verification failed")
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
type Proof struct {

	// size of the system
	size uint64

	// Commitments to h1, h2, t, z, f, h
	h1, h2, t, z, f, h kzg.Digest

	// Batch opening proof of h1, h2, z, t
	BatchedProof kzg.BatchOpeningProof

	// Batch opening proof of h1, h2, z shifted by g
	BatchedProofShifted kzg.BatchOpeningProof
}

// sortByT sorts f by t, and put the result in res.
// f and t are supposed to be sorted by increasing order.
// and are supposed to be of correct size, that is len(t)=len(f)+1
func sortByT(f, t []fr.Element) []fr.Element {

	res := make([]fr.Element, 2*len(t)-1)
	c := 0
	i := 0
	for i < len(f) {
		if t[c].Cmp(&f[i]) < 0 {
			res[i+c] = t[c]
			c++
		} else {
			res[i+c] = f[i]
			i++
		}
	}
	for c < len(t) {
		res[i+c] = t[c]
		c++
	}
	return res
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

// computeH computes the quotient h where
// h = (x-1)*z*(1+beta)*(gamma+f)*(gamma(1+beta) + t+ beta*t(gX)) -
//		(x-1)*z(gX)*(gamma(1+beta) + h1 + beta*h1(gX))*(gamma(1+beta) + h2 + beta*h2(gX) )
//
// * cz, ch1, ch2, ct, cf are the polynomials z, h1, h2, t, f in canonical basis
// * _lz, _lh1, _lh2, _lt, _lf are the polynomials z, h1, h2, t, f in shifted Lagrange basis (domainH)
// * beta, gamma are the challenges
// * it returns h in canonical basis
// func computeH(cz, ch1, ch2, ct, cf []fr.Element, beta, gamma fr.Element, domainH *fft.Domain) []fr.Element {
func computeH(_lz, _lh1, _lh2, _lt, _lf []fr.Element, beta, gamma fr.Element, domainH *fft.Domain) []fr.Element {

	// result
	s := int(domainH.Cardinality)
	num := make([]fr.Element, domainH.Cardinality)

	// create domain (2*len(h1) is enough, since we divide by x^n-1)
	fmt.Printf("domainH.FinerGenerator: %s\n", domainH.FinerGenerator.String())

	var u, v, w, _g, m, n, one, t fr.Element
	t.SetUint64(2).
		Inverse(&t)
	_g.Square(&domainH.Generator).
		Exp(_g, big.NewInt(int64(s/2-1)))
	one.SetOne()
	v.Add(&one, &beta)
	w.Mul(&v, &gamma)

	var d [2]fr.Element
	d[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality>>1)))
	d[1].Neg(&d[0])
	d[0].Sub(&d[0], &one).Inverse(&d[0])
	d[1].Sub(&d[1], &one).Inverse(&d[1])

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
		fmt.Printf("--- _i %d ----\n", _i)
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

		num[_i].Mul(&num[_i], &d[i%2])
	}

	// DEBUG
	fmt.Println("lH: ")
	fmt.Printf("[")
	for i := 0; i < len(num); i++ {
		fmt.Printf("%s, ", num[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	// END DEBUG

	domainH.FFTInverse(num, fft.DIT, 1)

	return num
}

// computeH0 returns l0 * (z-1), in bit reversed order
func computeH0(lzCosetReversed []fr.Element, domainH *fft.Domain) []fr.Element {

	var d, one fr.Element
	one.SetOne()

	var g [2]fr.Element
	g[0].Exp(domainH.FinerGenerator, big.NewInt(int64(domainH.Cardinality)))
	g[1].Neg(&g[0])
	g[0].Sub(&g[0], &one)
	g[1].Sub(&g[1], &one)
	d.Set(&domainH.FinerGenerator)

	den := make([]fr.Element, len(lzCosetReversed))
	den[0].Set(&domainH.FinerGenerator)
	for i := 1; i < len(den); i++ {
		den[i].Mul(&den[i-1], &domainH.Generator)
	}
	den = fr.BatchInvert(den)

	res := make([]fr.Element, len(lzCosetReversed))
	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < len(lzCosetReversed); i++ {
		_i := int(bits.Reverse64(uint64(i)) >> nn)
		res[_i].Mul(&lzCosetReversed[_i], &g[i%2]).Mul(&res[_i], &den[i])
	}

	return res

}

// Prove returns proof that the values in f are in t.
func Prove(srs *kzg.SRS, f, t Table) (Proof, error) {

	// res
	var proof Proof
	var err error

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

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
	sort.Sort(f)
	sort.Sort(t)

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

	//cf[len(cf)-1].SetZero()

	// DEBUG
	fmt.Println("lf: ")
	fmt.Printf("[\n")
	for i := 0; i < len(lf); i++ {
		fmt.Printf("%s, ", lf[i].String())
	}
	fmt.Printf("]\n")
	fmt.Println("lt: ")
	fmt.Printf("[\n")
	for i := 0; i < len(lt); i++ {
		fmt.Printf("%s, ", lt[i].String())
	}
	fmt.Printf("]\n")
	fmt.Printf("f: ")
	for i := 0; i < len(cf); i++ {
		fmt.Printf("%s*x**%d+", cf[i].String(), i)
	}
	fmt.Println("")
	fmt.Printf("t: ")
	for i := 0; i < len(ct); i++ {
		fmt.Printf("%s*x**%d+", ct[i].String(), i)
	}
	fmt.Println("")
	// END DEBUG

	// commit to _f
	// cf := make([]fr.Element, cardDNum)
	// copy(cf, lf)
	// dNum.FFTInverse(cf, fft.DIF, 0)
	// fft.BitReverse(cf)

	// write f sorted by t
	lfSortedByt := sortByT(lf[:len(lf)-1], t)

	// DEBUG
	fmt.Println("lfSortedByt")
	fmt.Printf("[")
	for i := 0; i < len(lfSortedByt); i++ {
		fmt.Printf("%s, ", lfSortedByt[i].String())
	}
	fmt.Printf("]\n")
	// END DEBUG

	// compute h1, h2, commit to them
	lh1 := make([]fr.Element, cardDNum)
	lh2 := make([]fr.Element, cardDNum)
	ch1 := make([]fr.Element, cardDNum)
	ch2 := make([]fr.Element, cardDNum)
	copy(lh1, lfSortedByt[:cardDNum])
	copy(lh2, lfSortedByt[cardDNum-1:])

	// DEBUG
	fmt.Println("lh1:")
	fmt.Printf("[")
	for i := 0; i < len(lh1); i++ {
		fmt.Printf("%s, ", lh1[i].String())
	}
	fmt.Printf("]\n")
	fmt.Println("lh2:")
	fmt.Printf("[")
	for i := 0; i < len(lh2); i++ {
		fmt.Printf("%s, ", lh2[i].String())
	}
	fmt.Printf("]\n")
	// END DEBUG

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

	//  DEBUG
	fmt.Printf("h1: ")
	for i := 0; i < len(ch1); i++ {
		fmt.Printf("%s*x**%d+", ch1[i].String(), i)
	}
	fmt.Println("")
	fmt.Printf("h2: ")
	for i := 0; i < len(ch2); i++ {
		fmt.Printf("%s*x**%d+", ch2[i].String(), i)
	}
	fmt.Println("")
	//  END DEBUG

	// derive beta, gamma
	var beta, gamma fr.Element
	beta.SetUint64(13)
	gamma.SetUint64(23)

	// Compute to Z
	lz := computeZ(lf, lt, lh1, lh2, beta, gamma)

	// DEBUG
	fmt.Println("lz:")
	fmt.Println("[")
	for i := 0; i < len(lz); i++ {
		fmt.Printf("%s,\n", lz[i].String())
	}
	fmt.Println("]")
	// END DEBUG

	cz := make([]fr.Element, len(lz))
	copy(cz, lz)
	dNum.FFTInverse(cz, fft.DIF, 0)
	fft.BitReverse(cz)
	proof.z, err = kzg.Commit(cz, srs)
	if err != nil {
		return proof, err
	}

	// DEBUG
	fmt.Printf("cz: ")
	for i := 0; i < len(cz); i++ {
		fmt.Printf("%s*x**%d+", cz[i].String(), i)
	}
	fmt.Println("")
	// END DEBUG

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

	// DEBUG
	fmt.Println("_lz")
	fmt.Printf("[")
	for i := 0; i < len(_lz); i++ {
		fmt.Printf("%s, ", _lz[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	fmt.Println("_lh1")
	fmt.Printf("[")
	for i := 0; i < len(_lh1); i++ {
		fmt.Printf("%s, ", _lh1[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	fmt.Println("_lh2")
	fmt.Printf("[")
	for i := 0; i < len(_lh2); i++ {
		fmt.Printf("%s, ", _lh2[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	fmt.Println("_lt")
	fmt.Printf("[")
	for i := 0; i < len(_lt); i++ {
		fmt.Printf("%s, ", _lt[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	fmt.Println("_lf")
	fmt.Printf("[")
	for i := 0; i < len(_lf); i++ {
		fmt.Printf("%s, ", _lf[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	// END DEBUG

	// compute h
	ch := computeH(_lz, _lh1, _lh2, _lt, _lf, beta, gamma, domainH)
	fmt.Printf("h: (len(h)= %d)\n", len(ch))
	for i := 0; i < len(ch); i++ {
		fmt.Printf("%s*x**%d+", ch[i].String(), i)
	}
	fmt.Println("")
	proof.h, err = kzg.Commit(ch, srs)
	if err != nil {
		return proof, err
	}

	// compute h0

	// compute ho

	// compute hn

	// build the opening proofs
	var nu fr.Element
	nu.SetUint64(234)
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

	// DEBUG
	fmt.Printf("h1(nu): %s\n", proof.BatchedProof.ClaimedValues[0].String())
	fmt.Printf("h2(nu): %s\n", proof.BatchedProof.ClaimedValues[1].String())
	fmt.Printf("t(nu): %s\n", proof.BatchedProof.ClaimedValues[2].String())
	fmt.Printf("z(nu): %s\n", proof.BatchedProof.ClaimedValues[3].String())
	fmt.Printf("f(nu): %s\n", proof.BatchedProof.ClaimedValues[4].String())
	fmt.Println("")
	// END DEBUG

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
	// DEBUG
	fmt.Printf("h1(g*nu): %s\n", proof.BatchedProofShifted.ClaimedValues[0].String())
	fmt.Printf("h2(g*nu): %s\n", proof.BatchedProofShifted.ClaimedValues[1].String())
	fmt.Printf("t(g*nu): %s\n", proof.BatchedProofShifted.ClaimedValues[2].String())
	fmt.Printf("z(g*nu): %s\n", proof.BatchedProofShifted.ClaimedValues[3].String())
	fmt.Println("")
	// END DEBUG

	return proof, nil

}

// Verify verifies that a plookup proof is correct
func Verify(srs *kzg.SRS, proof Proof) error {

	// hash function that is used for Fiat Shamir
	hFunc := sha256.New()

	// check opening proofs
	err := kzg.BatchVerifySinglePoint(
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

	// DEBUG
	fmt.Printf("hnu: %s\n", proof.BatchedProof.ClaimedValues[5].String())
	// END DEBUG

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
	var lhs, rhs, nu, nun, g, _g, e, a, v, w, beta, gamma, one fr.Element
	nu.SetUint64(234)
	d := fft.NewDomain(proof.size, 0, false) // only there to access to root of 1...
	one.SetOne()
	g.Exp(d.Generator, big.NewInt(int64(d.Cardinality-1)))

	beta.SetUint64(13)
	gamma.SetUint64(23)
	v.Add(&one, &beta)
	w.Mul(&v, &gamma)

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

	// DEBUG
	fmt.Printf("lhs: %s\n", lhs.String())
	fmt.Printf("rhs: %s\n", rhs.String())
	// END DEBUG

	lhs.Sub(&lhs, &rhs)
	nun.Exp(nu, big.NewInt(int64(d.Cardinality)))
	_g.Sub(&nun, &one)
	e.Mul(&proof.BatchedProof.ClaimedValues[5], &_g)
	if !lhs.Equal(&e) {
		return ErrPlookupVerification
	}

	// check consistancy of bounds
	var l0, ln, d1, d2 fr.Element
	l0.Exp(nu, big.NewInt(int64(d.Cardinality))).Sub(&l0, &one)
	ln.Set(&l0)
	d1.Sub(&nu, &one)
	d2.Sub(&nu, &g)
	l0.Div(&l0, &d1)
	ln.Div(&ln, &d2)
	fmt.Printf("l0: %s\n", l0.String())
	fmt.Printf("ln: %s\n", ln.String())

	return nil
}
