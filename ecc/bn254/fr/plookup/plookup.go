package plookup

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
)

var ErrNotInTable = errors.New("some value in the vector is not in the lookup table")

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

	// Commitments to h1, h2
	comh1, comh2 kzg.Digest

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
// * beta, gamma are the challenges
// * it returns h in canonical basis
func computeH(cz, ch1, ch2, ct, cf []fr.Element, beta, gamma fr.Element) []fr.Element {

	// create domain (2*len(h1) is enough, since we divide by x^n-1)
	s := len(ch1)
	domainH := fft.NewDomain(uint64(2*s), 1, false)
	fmt.Printf("g: %s\n", domainH.FinerGenerator.String())

	// compute the numerator
	num := make([]fr.Element, 2*s)
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

	var u, v, w, _g, m, n, one, t fr.Element
	t.SetUint64(2).
		Inverse(&t)
	_g.Square(&domainH.Generator).
		Exp(_g, big.NewInt(int64(s-1)))
	one.SetOne()
	v.Add(&one, &beta)
	w.Mul(&v, &gamma)

	g := make([]fr.Element, 2*s)
	g[0].Set(&domainH.FinerGenerator)
	for i := 1; i < 2*s; i++ {
		g[i].Mul(&g[i-1], &domainH.Generator)
	}

	nn := uint64(64 - bits.TrailingZeros64(domainH.Cardinality))

	for i := 0; i < 2*s; i++ {

		_i := int(bits.Reverse64(uint64(i)) >> nn)
		_is := int(bits.Reverse64(uint64((i+2)%(2*s))) >> nn)

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

		// g.Mul(&g, &domainH.Generator)

	}

	// DEBUG
	fft.BitReverse(num)
	fmt.Println("lH: ")
	fmt.Printf("[")
	for i := 0; i < len(num); i++ {
		fmt.Printf("%s, ", num[i].String())
	}
	fmt.Printf("]")
	fmt.Println("")
	// END DEBUG

	domainH.FFTInverse(num, fft.DIF, 1)

	return num

}

// Prove returns proof that the values in f are in t.
// func Prove(srs *kzg.SRS, f, t Table) (Proof, error) {
func Prove(f, t Table) (Proof, error) {

	// create domains
	var dNum *fft.Domain
	if len(t) <= len(f) {
		dNum = fft.NewDomain(uint64(len(f)+1), 0, false)
	} else {
		dNum = fft.NewDomain(uint64(len(t)), 0, false)
	}
	cardDNum := int(dNum.Cardinality)

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
	_cf := make([]fr.Element, cardDNum)
	copy(_cf, cf)
	fft.BitReverse(_cf)
	fmt.Printf("f: ")
	for i := 0; i < len(_cf); i++ {
		fmt.Printf("%s*x**%d+", _cf[i].String(), i)
	}
	fmt.Println("")
	_ct := make([]fr.Element, cardDNum)
	copy(_ct, ct)
	fft.BitReverse(_ct)
	fmt.Printf("t: ")
	for i := 0; i < len(_ct); i++ {
		fmt.Printf("%s*x**%d+", _ct[i].String(), i)
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

	//  DEBUG
	_ch1 := make([]fr.Element, cardDNum)
	_ch2 := make([]fr.Element, cardDNum)
	copy(_ch1, ch1)
	copy(_ch2, ch2)
	fft.BitReverse(_ch1)
	fft.BitReverse(_ch2)
	fmt.Printf("h1: ")
	for i := 0; i < len(_ch1); i++ {
		fmt.Printf("%s*x**%d+", _ch1[i].String(), i)
	}
	fmt.Println("")
	fmt.Printf("h2: ")
	for i := 0; i < len(_ch2); i++ {
		fmt.Printf("%s*x**%d+", _ch2[i].String(), i)
	}
	fmt.Println("")
	//  END DEBUG

	// comh1, err := kzg.Commit(h1, srs)
	// if err != nil {
	// 	return Proof{}, err
	// }
	// comh2, err := kzg.Commit(h2, srs)
	// if err != nil {
	// 	return Proof{}, err
	// }

	// derive beta, gamma
	var beta, gamma fr.Element
	// beta.SetRandom()
	// gamma.SetRandom()
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

	// DEBUG
	_cz := make([]fr.Element, len(lz))
	copy(_cz, cz)
	fft.BitReverse(_cz)
	fmt.Printf("cz: ")
	for i := 0; i < len(_ch2); i++ {
		fmt.Printf("%s*x**%d+", _cz[i].String(), i)
	}
	fmt.Println("")
	// END DEBUG

	// compute h
	fft.BitReverse(cz)
	fft.BitReverse(ch1)
	fft.BitReverse(ch2)
	fft.BitReverse(ct)
	fft.BitReverse(cf)
	h := computeH(cz, ch1, ch2, ct, cf, beta, gamma)
	fmt.Println("h: ")
	for i := 0; i < len(h); i++ {
		fmt.Printf("%s*x**%d+", h[i].String(), i)
	}
	fmt.Println("")

	// build the opening proofs

	return Proof{}, nil

}
