package fri

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

func TestBuildProofOfProximity(t *testing.T) {

	size := uint64(8)

	p := polynomial.New(size)
	for i := 0; i < int(size); i++ {
		p[i].SetUint64(3 << i)
	}

	iop := RADIX_2_FRI.New(size, sha256.New())
	proof, err := iop.BuildProofOfProximity(p)
	if err != nil {
		t.Fatal(err)
	}

	err = iop.VerifyProofOfProximity(proof)
	if err != nil {
		t.Fatal(err)
	}

}

func TestDeriveQueriesPositions(t *testing.T) {

	_s := RADIX_2_FRI.New(8, sha256.New())
	s := _s.(radixTwoFri)
	var r, g fr.Element
	r.Mul(&s.domains[0].Generator, &s.domains[0].Generator).Mul(&r, &s.domains[0].Generator)
	pos := s.deriveQueriesPositions(r)
	g.Set(&s.domains[0].Generator)
	n := int(s.domains[0].Cardinality)

	// conversion of indices from ordered to canonical, _n is the size of the slice
	// _p is the index to convert. It returns g^u, g^v where {g^u, g^v} is the fiber
	// of g^(2*_p)
	convert := func(_p, _n int) (_u, _v big.Int) {
		if _p%2 == 0 {
			_u.SetInt64(int64(_p / 2))
			_v.SetInt64(int64(_p/2 + n/2))
		} else {
			l := (n - 1 - _p) / 2
			_u.SetInt64(int64(n - 1 - l))
			_v.SetInt64(int64(n - 1 - l - n/2))
		}
		return
	}

	for i := 0; i < len(pos); i++ {

		u, v := convert(pos[i], n)

		var g1, g2, r1, r2 fr.Element
		g1.Exp(g, &u).Square(&g1)
		g2.Exp(g, &v).Square(&g2)

		if !g1.Equal(&g2) {
			t.Fatal("g1 and g2 are not in the same fiber")
		}
		g.Square(&g)
		n = n >> 1
		if i < len(pos)-1 {
			u, v := convert(pos[i+1], n)
			r1.Exp(g, &u)
			r2.Exp(g, &v)
			if !g1.Equal(&r1) && !g2.Equal(&r2) {
				t.Fatal("g1 and g2 are not in the correct fiber")
			}
		}
	}
}

// Benchmarks

func BenchmarkProximityVerification(b *testing.B) {

	baseSize := 16

	for i := 0; i < 10; i++ {

		size := baseSize << i
		p := polynomial.New(uint64(size))
		for k := 0; k < size; k++ {
			p[k].SetRandom()
		}

		iop := RADIX_2_FRI.New(uint64(size), sha256.New())
		proof, _ := iop.BuildProofOfProximity(p)

		b.Run(fmt.Sprintf("Polynomial size %d", size), func(b *testing.B) {
			b.ResetTimer()
			for l := 0; l < b.N; l++ {
				iop.VerifyProofOfProximity(proof)
			}
		})

	}

}
