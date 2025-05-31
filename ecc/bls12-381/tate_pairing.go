package bls12381

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// g1Proj is a point in homogeneous projective coordinates
type g1Proj struct {
	x, y, z fp.Element
}

// FromAffine sets p = Q, p in homogenous projective, Q in affine
func (p *g1Proj) FromAffine(Q *G1Affine) *g1Proj {
	if Q.X.IsZero() && Q.Y.IsZero() {
		p.z.SetZero()
		p.x.SetOne()
		p.y.SetOne()
		return p
	}
	p.z.SetOne()
	p.x.Set(&Q.X)
	p.y.Set(&Q.Y)
	return p
}

// line is a projective line
type line struct {
	r0 fp.Element
	r1 fp.Element
	r2 fp.Element
}

// doubleStep doubles the homogeneous projective point p, and computes the
// tangent line through [2]p.
//
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g1Proj) doubleStep(line *line) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)

	E.Neg(&D)

	F.Double(&E).
		Add(&F, &E)
	G.Add(&B, &F)
	G.Halve()
	H.Add(&p.y, &p.z).
		Square(&H)
	t1.Add(&B, &C)
	H.Sub(&H, &t1)
	I.Sub(&E, &B)
	J.Square(&p.x)
	EE.Square(&E)
	K.Double(&EE).
		Add(&K, &EE)

	// X, Y, Z
	p.x.Sub(&B, &F).
		Mul(&p.x, &A)
	p.y.Square(&G).
		Sub(&p.y, &K)
	p.z.Mul(&B, &H)

	// Line evaluation
	line.r0.Neg(&H)
	line.r1.Double(&J).
		Add(&line.r1, &J)
	line.r2.Set(&I)
}

// addMixedStep adds a point in homogeneous projective and and a point in affine
// coordinates, and evaluates the line in the (Tate) Miller loop.
//
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g1Proj) AddMixedStep(line *line, a *G1Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, C, D, E, F, G, H, t0, t1, t2, J fp.Element
	Y2Z1.Mul(&a.Y, &p.z)
	O.Sub(&p.y, &Y2Z1)
	X2Z1.Mul(&a.X, &p.z)
	L.Sub(&p.x, &X2Z1)
	C.Square(&O)
	D.Square(&L)
	E.Mul(&L, &D)
	F.Mul(&p.z, &C)
	G.Mul(&p.x, &D)
	t0.Double(&G)
	H.Add(&E, &F).
		Sub(&H, &t0)
	t1.Mul(&p.y, &E)

	// X, Y, Z
	p.x.Mul(&L, &H)
	p.y.Sub(&G, &H).
		Mul(&p.y, &O).
		Sub(&p.y, &t1)
	p.z.Mul(&E, &p.z)

	t2.Mul(&L, &a.Y)
	J.Mul(&a.X, &O).
		Sub(&J, &t2)

	// Line evaluation
	line.r0.Set(&L)
	line.r1.Neg(&O)
	line.r2.Set(&J)
}

// MillerLoopTate...
func MillerLoopTate(P G1Affine, Q G1Affine) (fp.Element, error) {
	// projective points for P
	var pProj g1Proj
	var pNeg G1Affine
	pProj.FromAffine(&P)
	pNeg.Neg(&P)

	// f_{r,P}(Q)
	var result fp.Element
	result.SetOne()
	var l line

	for i := 313; i >= 0; i-- {
		result.Square(&result)

		pProj.doubleStep(&l)
		// line evaluation
		l.r1.Mul(&l.r1, &Q.X)
		l.r0.Mul(&l.r0, &Q.Y)
		result.MulBy034(&l.r0, &l.r1, &l.r2)

		if loopCounterTate[i] == 1 {
			pProj.addMixedStep(&l, &p)
			// line evaluation
			l.r1.Mul(&l.r1, &Q.X)
			l.r0.Mul(&l.r0, &Q.Y)
			result.MulBy034(&l.r0, &l.r1, &l.r2)

		} else if loopCounterTate[i] == -1 {
			pProj.addMixedStep(&l, &pNeg)
			// line evaluation
			l.r1.Mul(&l.r1, &q.X)
			l.r0.Mul(&l.r0, &q.Y)
			result.MulBy034(&l.r0, &l.r1, &l.r2)
		}
	}

	return result, nil
}
