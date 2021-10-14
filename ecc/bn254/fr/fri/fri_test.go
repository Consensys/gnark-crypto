package fri

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

// func TestMerkleTree(t *testing.T) {

// 	h := sha256.New()

// 	var b1, b2 bytes.Buffer
// 	for i := 0; i < 16; i++ {
// 		var a fr.Element
// 		a.SetRandom()
// 		b1.Write(a.Marshal())
// 		b2.Write(a.Marshal())
// 	}

// 	var p1, p2 merkleProof
// 	t1 := merkletree.New(h)
// 	err := t1.SetIndex(0)
// 	t1.ReadAll(&b1, 32)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	mr, proofSet, proofIndex, numLeaves := t1.Prove()
// 	fmt.Printf("num leaves: %d\n", numLeaves)
// 	p1 = merkleProof{mr, proofSet, proofIndex, numLeaves}

// 	t2 := merkletree.New(h)
// 	err = t2.SetIndex(1)
// 	t2.ReadAll(&b2, 32)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	mr, proofSet, proofIndex, numLeaves = t2.Prove()
// 	p2 = merkleProof{mr, proofSet, proofIndex, numLeaves}

// 	for i := 0; i < len(p1.proofSet[2]); i++ {
// 		fmt.Printf("%x", p1.proofSet[2][i])
// 	}
// 	fmt.Println("")
// 	for i := 0; i < len(p2.proofSet[2]); i++ {
// 		fmt.Printf("%x", p2.proofSet[2][i])
// 	}
// 	fmt.Println("")

// 	fmt.Printf("p1 index: %d\n", p1.proofIndex)
// 	fmt.Printf("p2 index: %d\n", p2.proofIndex)

// 	h.Reset()
// 	h.Write(p1.proofSet[0])
// 	bs1 := h.Sum(nil)
// 	h.Reset()
// 	h.Write(bs1)
// 	h.Write(p1.proofSet[1])
// 	bs1 = h.Sum(nil)

// 	h.Reset()
// 	h.Write(p2.proofSet[0])
// 	bs2 := h.Sum(nil)
// 	h.Reset()
// 	h.Write(p2.proofSet[1])
// 	h.Write(bs2)
// 	bs2 = h.Sum(nil)

// 	for i := 0; i < len(bs1); i++ {
// 		fmt.Printf("%x", bs1[i])
// 	}
// 	fmt.Println("")
// 	for i := 0; i < len(bs2); i++ {
// 		fmt.Printf("%x", bs2[i])
// 	}
// }

func TestBuildProofOfProximity(t *testing.T) {

	p := polynomial.New(16)

	iop := RADIX_2_FRI.New(16, sha256.New())
	proof, err := iop.BuildProofOfProximity(p)
	if err != nil {
		t.Fatal(err)
	}

	// for i:=0; i<len(proof.interactions); i++{
	// 	fmt.Printf("%x\n", proof.interactions[i])
	// }

	err = iop.VerifyProofOfProximity(proof)
	if err != nil {
		t.Fatal(err)
	}

}

func TestDeriveQueriesPositions(t *testing.T) {

	_s := RADIX_2_FRI.New(16, sha256.New())
	s := _s.(radixTwoFri)
	var r, g fr.Element
	r.Mul(&s.domains[0].Generator, &s.domains[0].Generator).Mul(&r, &s.domains[0].Generator)
	g.Set(&s.domains[0].Generator)
	pos := s.deriveQueriesPositions(r)
	for i := 0; i < len(pos); i++ {
		fmt.Printf("%d\n", pos[i])
	}
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
		// var g1, g2 fr.Element
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
