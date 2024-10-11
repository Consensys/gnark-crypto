package poseidon2

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// poseidon
// https://github.com/argumentcomputer/neptune/blob/main/spec/poseidon_spec.pdf

// poseidon2 ref implem
// https://github.com/HorizenLabs/poseidon2/blob/main/plain_implementations/src/poseidon2/poseidon2.rs

// M ∈ {80,128,256}, security level in bits

// Hash stores the state of the poseidon2 permutation and provides poseidon2 permutation
// methods on the state
type Hash struct {

	// len(preimage)+len(digest)=len(preimage)+ceil(log(2*<security_level>/r))
	t int

	// sbox degree
	d int

	// state
	state []fr.Element

	// number of full rounds (even number)
	rF int

	// number of partial rounds
	rP int
}

// sBox applies the sBox on state[index]
func (h *Hash) sBox(index int) {
	var tmp fr.Element
	tmp.Set(&h.state[index])
	if h.d == 3 {
		h.state[index].Square(&h.state[index]).
			Mul(&h.state[index], &tmp)
	} else if h.d == 5 {
		h.state[index].Square(&h.state[index]).
			Square(&h.state[index]).
			Mul(&h.state[index], &tmp)
	} else if h.d == 7 {
		h.state[index].Square(&h.state[index]).
			Mul(&h.state[index], &tmp).
			Square(&h.state[index]).
			Mul(&h.state[index], &tmp)
	}
}

// matMulM4 computes
// s <- M4*s
// where M4=
// (5 7 1 3)
// (4 6 1 1)
// (1 3 5 7)
// (1 1 4 6)
// on chunks of 4 elemts on each part of the state
// see https://eprint.iacr.org/2023/323.pdf appendix B for the addition chain
func (h *Hash) matMulM4InPlace(s []fr.Element) {
	c := len(s) / 4
	for i := 0; i < c; i++ {
		var t0, t1, t2, t3, t4, t5, t6, t7 fr.Element
		t0.Add(&s[4*i], &s[4*i+1])               // s0+s1
		t1.Add(&s[4*i+2], &s[4*i+3])             // s2+s3
		t2.Double(&s[4*i+1]).Add(&t2, &t1)       // 2s1+t1
		t3.Double(&s[4*i+3]).Add(&t3, &t0)       // 2s3+t0
		t4.Double(&t1).Double(&t4).Add(&t4, &t3) // 4t1+t3
		t5.Double(&t0).Double(&t5).Add(&t5, &t2) // 4t0+t2
		t6.Add(&t3, &t5)                         // t3+t4
		t7.Add(&t2, &t4)                         // t2+t4
		s[4*i].Set(&t6)
		s[4*i+1].Set(&t5)
		s[4*i+2].Set(&t7)
		s[4*i+3].Set(&t4)
	}
}

// when t=2,3 the state is multiplied by circ(2,1) and circ(2,1,1)
// see https://eprint.iacr.org/2023/323.pdf page 15, case t=2,3
//
// when t=0[4], the state is multiplied by circ(2M4,M4,..,M4)
// see https://eprint.iacr.org/2023/323.pdf
func (s *Hash) matMulExternalInPlace() {

	if s.t == 2 {
		var tmp fr.Element
		tmp.Add(&s.state[0], &s.state[1])
		s.state[0].Add(&tmp, &s.state[0])
		s.state[1].Add(&tmp, &s.state[1])
	} else if s.t == 3 {
		var tmp fr.Element
		tmp.Add(&s.state[0], &s.state[1]).
			Add(&tmp, &s.state[2])
		s.state[0].Add(&tmp, &s.state[0])
		s.state[1].Add(&tmp, &s.state[1])
		s.state[2].Add(&tmp, &s.state[2])
	} else if s.t == 4 {
		s.matMulM4InPlace(s.state)
	} else {
		// at this stage t is supposed to be a multiple of 4
		// the MDS matrix is circ(2M4,M4,..,M4)
		s.matMulM4InPlace(s.state)
		tmp := make([]fr.Element, 4)
		for i := 0; i < s.t/4; i++ {
			tmp[0].Add(&tmp[0], &s.state[4*i])
			tmp[1].Add(&tmp[1], &s.state[4*i+1])
			tmp[2].Add(&tmp[2], &s.state[4*i+2])
			tmp[3].Add(&tmp[3], &s.state[4*i+3])
		}
		for i := 0; i < s.t/4; i++ {
			s.state[4*i].Add(&s.state[4*i], &tmp[0])
			s.state[4*i+1].Add(&s.state[4*i], &tmp[1])
			s.state[4*i+2].Add(&s.state[4*i], &tmp[2])
			s.state[4*i+3].Add(&s.state[4*i], &tmp[3])
		}
	}
}
