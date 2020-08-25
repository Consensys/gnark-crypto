package point

const Point = `

import (
	"math/big"
	"runtime"

	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fp"
	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fr"
	"github.com/consensys/gurvy/utils/debug"
)

// {{ toUpper .PointName }}Jac is a point with {{.CoordType}} coordinates
type {{ toUpper .PointName }}Jac struct {
	X, Y, Z {{.CoordType}}
}

// {{ toUpper .PointName }}Proj point in projective coordinates
type {{ toUpper .PointName }}Proj struct {
	X, Y, Z {{.CoordType}}
}

// {{ toUpper .PointName }}Affine point in affine coordinates
type {{ toUpper .PointName }}Affine struct {
	X, Y {{.CoordType}}
}

//  {{ toLower .PointName }}JacExtended parameterized jacobian coordinates (x=X/ZZ, y=Y/ZZZ, ZZ**3=ZZZ**2)
type {{ toLower .PointName }}JacExtended struct {
	X, Y, ZZ, ZZZ {{.CoordType}}
}

// SetInfinity sets p to O
func (p *{{ toLower .PointName }}JacExtended) SetInfinity() *{{ toLower .PointName }}JacExtended {
	p.X.SetOne()
	p.Y.SetOne()
	p.ZZ.SetZero()
	p.ZZZ.SetZero()
	return p
}

// ToAffine sets p in affine coords
func (p *{{ toLower .PointName }}JacExtended) ToAffine(Q *{{ toUpper .PointName }}Affine) *{{ toUpper .PointName }}Affine {
	var zero {{.CoordType}}
	if p.ZZ.Equal(&zero) {
		Q.X.Set(&zero)
		Q.Y.Set(&zero)
		return Q
	}
	Q.X.Inverse(&p.ZZ).Mul(&Q.X, &p.X)
	Q.Y.Inverse(&p.ZZZ).Mul(&Q.Y, &p.Y)
	return Q
}

// ToJac sets p in affine coords
func (p *{{ toLower .PointName }}JacExtended) ToJac(Q *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	var zero {{.CoordType}}
	if p.ZZ.Equal(&zero) {
		Q.Set(&{{ toLower .PointName }}Infinity)
		return Q
	}
	Q.X.Mul(&p.ZZ, &p.X).Mul(&Q.X, &p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).Mul(&Q.Y, &p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
}

// unsafeToJac sets p in affine coords, but don't check for infinity
func (p *{{ toLower .PointName }}JacExtended) unsafeToJac(Q *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	Q.X.Mul(&p.ZZ, &p.X).Mul(&Q.X, &p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).Mul(&Q.Y, &p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
}


// mSub
// http://www.hyperelliptic.org/EFD/ {{ toLower .PointName }}p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *{{ toLower .PointName }}JacExtended) mSub(a *{{ toUpper .PointName }}Affine) *{{ toLower .PointName }}JacExtended {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.Y.Neg(&p.Y)
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var U2, S2, P, R, PP, PPP, Q, Q2, RR, X3, Y3 {{.CoordType}}

	// p2: a, p1: p
	U2.Mul(&a.X, &p.ZZ)
	S2.Mul(&a.Y, &p.ZZZ)
	S2.Neg(&S2)


	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)

	pIsZero := P.IsZero()
	rIsZero := R.IsZero()

	if pIsZero && rIsZero {
		return p.doubleNeg(a)
	} else if pIsZero {
		p.ZZ.SetZero()
		p.ZZZ.SetZero()
		return p
	} 



	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.Double(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).Mul(&Y3, &R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &PPP)

	return p
}


// mAdd
// http://www.hyperelliptic.org/EFD/ {{ toLower .PointName }}p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *{{ toLower .PointName }}JacExtended) mAdd(a *{{ toUpper .PointName }}Affine) *{{ toLower .PointName }}JacExtended {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var U2, S2, P, R, PP, PPP, Q, Q2, RR, X3, Y3 {{.CoordType}}

	// p2: a, p1: p
	U2.Mul(&a.X, &p.ZZ)
	S2.Mul(&a.Y, &p.ZZZ)

	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)

	pIsZero := P.IsZero()
	rIsZero := R.IsZero()

	if pIsZero && rIsZero {
		return p.double(a)
	} else if pIsZero {
		p.ZZ.SetZero()
		p.ZZZ.SetZero()
		return p
	} 

	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.Double(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).Mul(&Y3, &R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &PPP)

	return p
}

func (p *{{ toLower .PointName }}JacExtended) doubleNeg(q *{{ toUpper .PointName }}Affine) *{{ toLower .PointName }}JacExtended {

	var U, S, M, _M, Y3 {{.CoordType}}

	U.Double(&q.Y)
	U.Neg(&U)
	p.ZZ.Square(&U)
	p.ZZZ.Mul(&U, &p.ZZ)
	S.Mul(&q.X, &p.ZZ)
	_M.Square(&q.X)
	M.Double(&_M).
		Add(&M, &_M) // -> + a, but a=0 here
	p.X.Square(&M).
		Sub(&p.X, &S).
		Sub(&p.X, &S)
	Y3.Sub(&S, &p.X).Mul(&Y3, &M)
	U.Mul(&p.ZZZ, &q.Y)
	U.Neg(&U)
	p.Y.Sub(&Y3, &U)

	return p
}


// double point in ZZ coords
// http://www.hyperelliptic.org/EFD/ {{ toLower .PointName }}p/auto-shortw-xyzz.html#doubling-dbl-2008-s-1
func (p *{{ toLower .PointName }}JacExtended) double(q *{{ toUpper .PointName }}Affine) *{{ toLower .PointName }}JacExtended {

	var U, S, M, _M, Y3 {{.CoordType}}

	U.Double(&q.Y)
	p.ZZ.Square(&U)
	p.ZZZ.Mul(&U, &p.ZZ)
	S.Mul(&q.X, &p.ZZ)
	_M.Square(&q.X)
	M.Double(&_M).
		Add(&M, &_M) // -> + a, but a=0 here
	p.X.Square(&M).
		Sub(&p.X, &S).
		Sub(&p.X, &S)
	Y3.Sub(&S, &p.X).Mul(&Y3, &M)
	U.Mul(&p.ZZZ, &q.Y)
	p.Y.Sub(&Y3, &U)

	return p
}

// Set set p to the provided point
func (p *{{ toUpper .PointName }}Jac) Set(a *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	p.X.Set(&a.X)
	p.Y.Set(&a.Y)
	p.Z.Set(&a.Z)
	return p
}

// Equal tests if two points (in Jacobian coordinates) are equal
func (p *{{ toUpper .PointName }}Jac) Equal(a *{{ toUpper .PointName }}Jac) bool {

	if p.Z.IsZero() && a.Z.IsZero() {
		return true
	}
	_p := {{ toUpper .PointName }}Affine{}
	_p.FromJacobian(p)

	_a := {{ toUpper .PointName }}Affine{}
	_a.FromJacobian(a)

	return _p.X.Equal(&_a.X) && _p.Y.Equal(&_a.Y)
}

// Equal tests if two points (in Affine coordinates) are equal
func (p *{{ toUpper .PointName }}Affine) Equal(a *{{ toUpper .PointName }}Affine) bool {
	return p.X.Equal(&a.X) && p.Y.Equal(&a.Y)
}

// Neg computes -G
func (p *{{ toUpper .PointName }}Jac) Neg(a *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	p.Set(a)
	p.Y.Neg(&a.Y)
	return p
}

// Neg computes -G
func (p *{{ toUpper .PointName }}Affine) Neg(a *{{ toUpper .PointName }}Affine) *{{ toUpper .PointName }}Affine {
	p.X.Set(&a.X)
	p.Y.Neg(&a.Y)
	return p
}

// SubAssign substracts two points on the curve
func (p *{{ toUpper .PointName }}Jac) SubAssign(a {{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	a.Y.Neg(&a.Y)
	p.AddAssign(&a)
	return p
}

// FromJacobian rescale a point in Jacobian coord in z=1 plane
func (p *{{ toUpper .PointName }}Affine) FromJacobian(p1 *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Affine {

	var a, b {{.CoordType}}

	if p1.Z.IsZero() {
		p.X.SetZero()
		p.Y.SetZero()
		return p
	}

	a.Inverse(&p1.Z)
	b.Square(&a)
	p.X.Mul(&p1.X, &b)
	p.Y.Mul(&p1.Y, &b).Mul(&p.Y, &a)

	return p
}

// FromJacobian converts a point from Jacobian to projective coordinates
func (p *{{ toUpper .PointName }}Proj) FromJacobian(Q *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Proj {
	// memalloc
	var buf {{.CoordType}}
	buf.Square(&Q.Z)

	p.X.Mul(&Q.X, &Q.Z)
	p.Y.Set(&Q.Y)
	p.Z.Mul(&Q.Z, &buf)

	return p
}

func (p *{{ toUpper .PointName }}Jac) String() string {
	if p.Z.IsZero() {
		return "O"
	}
	_p := {{ toUpper .PointName }}Affine{}
	_p.FromJacobian(p)
	return "E([" + _p.X.String() + "," + _p.Y.String() + "]),"
}

// FromAffine sets p = Q, p in Jacboian, Q in affine
func (p *{{ toUpper .PointName }}Jac) FromAffine(Q *{{ toUpper .PointName }}Affine) *{{ toUpper .PointName }}Jac {
	if Q.X.IsZero() && Q.Y.IsZero() {
		p.Z.SetZero()
		p.X.SetOne()
		p.Y.SetOne()
		return p
	}
	p.Z.SetOne()
	p.X.Set(&Q.X)
	p.Y.Set(&Q.Y)
	return p
}

func (p *{{ toUpper .PointName }}Affine) String() string {
	var x, y {{.CoordType}}
	x.Set(&p.X)
	y.Set(&p.Y)
	return "E([" + x.String() + "," + y.String() + "]),"
}

// IsInfinity checks if the point is infinity (in affine, it's encoded as (0,0))
func (p *{{ toUpper .PointName }}Affine) IsInfinity() bool {
	return p.X.IsZero() && p.Y.IsZero()
}

// IsOnCurve returns true if p in on the curve
func (p *{{ toUpper .PointName}}Proj) IsOnCurve() bool {
	var left, right, tmp  {{.CoordType}}
	left.Square(&p.Y).
		Mul(&left, &p.Z)
	right.Square(&p.X).
		Mul(&right, &p.X)
	tmp.Square(&p.Z).
		Mul(&tmp, &p.Z).
		{{- if eq .PointName "g1"}}
			Mul(&tmp, &B)
		{{- else}}
			Mul(&tmp, &Btwist)
		{{- end}}
	right.Add(&right, &tmp)
	return left.Equal(&right)
}

// IsOnCurve returns true if p in on the curve
func (p *{{ toUpper .PointName}}Jac) IsOnCurve() bool {
	var left, right, tmp  {{.CoordType}}
	left.Square(&p.Y)
	right.Square(&p.X).Mul(&right, &p.X)
	tmp.Square(&p.Z).
		Square(&tmp).
		Mul(&tmp, &p.Z).
		Mul(&tmp, &p.Z).
		{{- if eq .PointName "g1"}}
			Mul(&tmp, &B)
		{{- else}}
			Mul(&tmp, &Btwist)
		{{- end}}
	right.Add(&right, &tmp)
	return left.Equal(&right)
}

// IsOnCurve returns true if p in on the curve
func (p *{{ toUpper .PointName}}Affine) IsOnCurve() bool {
	var point {{ toUpper .PointName}}Jac
	point.FromAffine(p)
	return point.IsOnCurve() // call this function to handle infinity point
}


// AddAssign point addition in montgomery form
// https://hyperelliptic.org/EFD/{{ toLower .PointName }}p/auto-shortw-jacobian-3.html#addition-add-2007-bl
func (p *{{ toUpper .PointName }}Jac) AddAssign(a *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {

	// p is infinity, return a
	if p.Z.IsZero() {
		p.Set(a)
		return p
	}

	// a is infinity, return p
	if a.Z.IsZero() {
		return p
	}

	var Z1Z1, Z2Z2, U1, U2, S1, S2, H, I, J, r, V {{.CoordType}}
	Z1Z1.Square(&a.Z)
	Z2Z2.Square(&p.Z)
	U1.Mul(&a.X, &Z2Z2)
	U2.Mul(&p.X, &Z1Z1)
	S1.Mul(&a.Y, &p.Z).
		Mul(&S1, &Z2Z2)
	S2.Mul(&p.Y, &a.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U1.Equal(&U2) && S1.Equal(&S2) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &U1)
	I.Double(&H).
		Square(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &S1).Double(&r)
	V.Mul(&U1, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	S1.Mul(&S1, &J).Double(&S1)
	p.Y.Sub(&p.Y, &S1)
	p.Z.Add(&p.Z, &a.Z)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &Z2Z2).
		Mul(&p.Z, &H)

	return p
}

// AddMixed point addition
// http://www.hyperelliptic.org/EFD/{{ toLower .PointName }}p/auto-shortw-jacobian-0.html#addition-madd-2007-bl
func (p *{{ toUpper .PointName }}Jac) AddMixed(a *{{ toUpper .PointName }}Affine) *{{ toUpper .PointName }}Jac {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.Z.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.Z.SetOne()
		return p
	}

	// get some Element from our pool
	var Z1Z1, U2, S2, H, HH, I, J, r, V {{.CoordType}}
	Z1Z1.Square(&p.Z)
	U2.Mul(&a.X, &Z1Z1)
	S2.Mul(&a.Y, &p.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &p.X)
	HH.Square(&H)
	I.Double(&HH).Double(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &p.Y).Double(&r)
	V.Mul(&p.X, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	J.Mul(&J, &p.Y).Double(&J)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	p.Y.Sub(&p.Y, &J)
	p.Z.Add(&p.Z, &H)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &HH)

	return p
}

// Double doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/{{ toLower .PointName }}p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *{{ toUpper .PointName }}Jac) Double(q *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	p.Set(q)
	p.DoubleAssign()
	return p
}

// DoubleAssign doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/{{ toLower .PointName }}p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *{{ toUpper .PointName }}Jac) DoubleAssign() *{{ toUpper .PointName }}Jac {

	// get some Element from our pool
	var XX, YY, YYYY, ZZ, S, M, T {{.CoordType}}

	XX.Square(&p.X)
	YY.Square(&p.Y)
	YYYY.Square(&YY)
	ZZ.Square(&p.Z)
	S.Add(&p.X, &YY)
	S.Square(&S).
		Sub(&S, &XX).
		Sub(&S, &YYYY).
		Double(&S)
	M.Double(&XX).Add(&M, &XX)
	p.Z.Add(&p.Z, &p.Y).
		Square(&p.Z).
		Sub(&p.Z, &YY).
		Sub(&p.Z, &ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.Sub(&p.X, &T)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.Sub(&p.Y, &YYYY)

	return p
}


// ScalarMulByGen multiplies given scalar by generator
func (p *{{ toUpper .PointName }}Jac) ScalarMulByGen(s *big.Int) *{{ toUpper .PointName }}Jac {
	{{- if .GLV}}
		return p.ScalarMulGLV(&{{ toLower .PointName }}GenAff, s)
	{{- else}}
		return p.ScalarMultiplication(&{{ toLower .PointName }}GenAff, s)
	{{- end}}
}

// ScalarMultiplication 2-bits windowed exponentiation
func (p *{{ toUpper .PointName }}Jac) ScalarMultiplication(a *{{ toUpper .PointName }}Affine, s *big.Int) *{{ toUpper .PointName }}Jac {

	var res, tmp {{toUpper .PointName}}Jac
	var ops [3]{{toUpper .PointName}}Affine

	res.Set(&{{toLower .PointName}}Infinity)
	ops[0] = *a
	tmp.FromAffine(a).DoubleAssign()
	ops[1].FromJacobian(&tmp)
	tmp.AddMixed(a)
	ops[2].FromJacobian(&tmp)

	b := s.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0xc0)
		for j := 0; j < 4; j++ {
			res.DoubleAssign().DoubleAssign()
			c := (w & mask) >> (6 - 2*j)
			if c != 0 {
				res.AddMixed(&ops[c-1])
			}
			mask = mask >> 2
		}
	}
	p.Set(&res)

	return p

}

{{ if .GLV}}

// phi assigns p to phi(a) where phi: (x,y)->(ux,y), and returns p
func (p *{{toUpper .PointName}}Jac) phi(a *{{toUpper .PointName}}Affine) *{{toUpper .PointName}}Jac {
	p.FromAffine(a)
	{{if eq .CoordType "E2"}}
		p.X.MulByElement(&p.X, &thirdRootOne{{toUpper .PointName}})
	{{else}}
		p.X.Mul(&p.X, &thirdRootOne{{toUpper .PointName}})
	{{end}}
	return p
}

// ScalarMulGLV performs scalar multiplication using GLV
func (p *{{toUpper .PointName}}Jac) ScalarMulGLV(a *{{toUpper .PointName}}Affine, s *big.Int) *{{toUpper .PointName}}Jac {

	var table [3]{{toUpper .PointName}}Jac
	var zero big.Int
	var res {{toUpper .PointName}}Jac
	var k1, k2 fr.Element

	res.Set(&{{toLower .PointName}}Infinity)

	// table stores [+-a, +-phi(a), +-a+-phi(a)]
	table[0].FromAffine(a)
	table[1].phi(a)

	// split the scalar, modifies +-a, phi(a) accordingly
	k := utils.SplitScalar(s, &glvBasis)

	if k[0].Cmp(&zero) == -1 {
		k[0].Neg(&k[0])
		table[0].Neg(&table[0])
	}
	if k[1].Cmp(&zero) == -1 {
		k[1].Neg(&k[1])
		table[1].Neg(&table[1])
	}
	table[2].Set(&table[0]).AddAssign(&table[1])

	// bounds on the lattice base vectors guarantee that k1, k2 are len(r)/2 bits long max
	k1.SetBigInt(&k[0]).FromMont()
	k2.SetBigInt(&k[1]).FromMont()

	// loop starts from len(k1)/2 due to the bounds
	for i := len(k1)/2 - 1; i >= 0; i-- {
		mask := uint64(1) << 63
		for j := 0; j < 64; j++ {
			res.Double(&res)
			b1 := (k1[i] & mask) >> (63 - j)
			b2 := (k2[i] & mask) >> (63 - j)
			if b1|b2 != 0 {
				s := (b2<<1 | b1)
				res.AddAssign(&table[s-1])
			}
			mask = mask >> 1
		}
	}

	p.Set(&res)
	return p
}

{{ end }}


// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf 
func (p *{{ toUpper .PointName }}Jac) MultiExp(points []{{ toUpper .PointName }}Affine, scalars []fr.Element) *{{ toUpper .PointName }}Jac {
	// note: 
	// each of the multiExpcX method is the same, except for the c constant it declares
	// duplicating (through template generation) these methods allows to declare the buckets on the stack
	// the choice of c needs to be improved: 
	// there is a theoritical value that gives optimal asymptotics
	// but in practice, other factors come into play, including:
	// * if c doesn't divide 64, the word size, then we're bound to select bits over 2 words of our scalars, instead of 1
	// * number of CPUs 
	// * cache friendliness (which depends on the host, G1 or G2... )
	//	--> for example, on BN256, a G1 point fits into one cache line of 64bytes, but a G2 point don't. 

	// approximate cost (in group operations)
	// cost = bits/c * (nbPoints + 2^{c-1})
	// this needs to be verified empirically. 
	// for example, on a MBP 2016, for G2 MultiExp > 8M points, hand picking c gives better results
	implementedCs := []int{
		{{- range $c :=  .CRange}} {{$c}},{{- end}}
	}

	nbPoints := len(points)
	min := math.MaxFloat64
	bestC := 0
	for _, c := range implementedCs {
		cc := fr.Limbs * 64 * (nbPoints + (1 << (c-1)))
		cost := float64(cc) / float64(c)
		if cost < min {
			min = cost
			bestC = c 
		}
	}

	// semaphore to limit number of cpus
	numCpus := runtime.NumCPU()
	chCpus := make(chan struct{}, numCpus)
	for i:=0; i < numCpus; i++ {
		chCpus <- struct{}{}
	}


	switch bestC {
	{{range $c :=  .CRange}}
	case {{$c}}:
		return p.multiExpc{{$c}}(points, scalars, chCpus)	
	{{end}}
	default:
		panic("unimplemented")
	}
}

// chunkReduce{{ toUpper .PointName }} reduces the weighted sum of the buckets into the result of the multiExp
func chunkReduce{{ toUpper .PointName }}(p *{{ toUpper .PointName }}Jac, c int, chTotals []chan {{ toUpper .PointName }}Jac)  *{{ toUpper .PointName }}Jac {
	totalj := <-chTotals[len(chTotals)-1]
	p.Set(&totalj)
	for j := len(chTotals) - 2; j >= 0; j-- {
		for l := 0; l < c; l++ {
			p.DoubleAssign()
		}
		totalj := <-chTotals[j]
		p.AddAssign(&totalj)
	}
	return p
}



{{range $c :=  .CRange}}

// multiExpc{{$c}} implements the multi exp  (section 4 of https://eprint.iacr.org/2012/549.pdf  )
// with c = {{$c}}
// all the multiExpcXX are the same (generated with templates) except for this const c = xx declaration
// that enables to declares array and allocate on the stack, but generates a lot of duplicate code in our gXX.go files.
// chCpus is a semaphore to limit number of CPUs running the bucket accumulation and iterating through the points at the same time
func (p *{{ toUpper $.PointName }}Jac) multiExpc{{$c}}(points []{{ toUpper $.PointName }}Affine, scalars []fr.Element, chCpus chan struct{}) *{{ toUpper $.PointName }}Jac {
	{{$cDividesBits := divides $c $.RBitLen}}
	const c  = {{$c}} 							// scalars partitioned into c-bit radixes
	const t = fr.Limbs * 64 / c        			// number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1	// low c bits are 1
	const nbChunks = t {{if not $cDividesBits }} + 1 {{end}} // note: if c doesn't divide fr.Bits, nbChunks != t)
	const msbWindow uint32 = (1 << (c -1)) 


	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan {{ toUpper $.PointName }}Jac
	for i:= 0; i< nbChunks; i++ {
		chTotals[i] = make(chan {{ toUpper $.PointName }}Jac, 1)
	}

	// step 1: we compute, for each scalars over c-bit wide windows, nbChunk digits
	// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and substract
	// 2^{c} to the current digit, making it negative.
	// negative digits will be processed in the next step as adding -G into the bucket instead of G
	// (computing -G is cheap, and this saves us half of the buckets)
	scalarsToDigits := func(scalars []fr.Element) (digits [][nbChunks]uint32) {
		const max = int(msbWindow)
		const twoc = (1 << c ) 
		digits = make([][nbChunks]uint32, len(scalars))
		
		// process the scalars in parallel
		parallel.Execute(len(scalars), func(start, end int) {
			for i:=start; i < end; i++ {
				var carry int

				// for each chunk in the scalar, compute the current digit, and an eventual carry
				for chunk := 0; chunk < nbChunks; chunk++ {
		
					// compute offset and word selector / shift to select the right bits of our windows
					jc := uint64(chunk * c)
					selectorIndex := jc / 64
					selectorShift := jc - (selectorIndex * 64)

					// init with carry if any
					digit := carry
					carry = 0

					// digit = value of the c-bit window
					digit += int((scalars[i][selectorIndex] & (selectorMask << selectorShift)) >> selectorShift)
					
					{{$cDivides64 := divides $c 64}}
					{{if not $cDivides64}}
						// c doesn't divide 64, which means we may need to select bits over 2 words
						multiWordSelect := int(selectorShift) > (64-c) && selectorIndex < (fr.Limbs - 1 )
						if multiWordSelect {
							// we are selecting bits over 2 words
							selectorIndexNext := selectorIndex+1
							nbBitsHigh := selectorShift - uint64(64-c)
							highShift := 64 - nbBitsHigh
							highShiftRight := highShift - (64 - selectorShift)
							digit += int((scalars[i][selectorIndexNext] << highShift) >> highShiftRight)
						}
					{{end}}

					// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and substract
					// 2^{c} to the current digit, making it negative.
					if digit >= max {
						digit -= twoc
						carry = 1 
					}

					if digit == 0 {
						continue // digit[i][chunk] = 0
					}

					if digit > 0 {
						digits[i][chunk] = uint32(digit)
					} else {
						// mark negative sign using msbWindow mask, a bit we know is not used. 
						digits[i][chunk] = uint32(-digit - 1) | msbWindow
					}
					
				}
			}
		})
		return 
	}
	digits := scalarsToDigits(scalars)

	// step 2
	// bucketAccumulate places points into buckets base on their selector and return the weighted bucket sum in given channel
	bucketAccumulate := func(chunk int, chRes chan<- {{ toUpper $.PointName }}Jac) {
		<-chCpus // wait and decrement avaiable CPUs on the semaphore

		// declare our buckets on the stack
		// notice that we have 2^{c-1} buckets instead of 2^{c} (see step1)
		// we use jacobian extended formulas here as they are faster than mixed addition
		var buckets [1<<(c-1)]{{ toLower $.PointName }}JacExtended
		for i := 0 ; i < len(buckets); i++ {
			buckets[i].SetInfinity()
		}


		// for each scalars, get the digit corresponding to the chunk we're processing. 
		for i := 0; i < len(digits); i++ {
			bits := digits[i][chunk]
			if bits == 0 {
				continue
			}
			
			// if msbWindow bit is set, we need to substract
			if bits & msbWindow == 0 {
				// add 
				buckets[bits-1].mAdd(&points[i])
			} else {
				// sub
				buckets[bits & ^msbWindow].mSub(&points[i])
			}
		}

		
		// reduce buckets into total
		// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

		var runningSum, tj, total {{ toUpper $.PointName }}Jac
		runningSum.Set(&{{ toLower $.PointName }}Infinity)
		total.Set(&{{ toLower $.PointName }}Infinity)
		for k := len(buckets) - 1; k >= 0; k-- {
			if !buckets[k].ZZ.IsZero() {
				runningSum.AddAssign(buckets[k].unsafeToJac(&tj))
			}
			total.AddAssign(&runningSum)
		}
		

		chRes <- total
		close(chRes)
		chCpus <- struct{}{} // increment avaiable CPUs into the semaphore
	} 

	// run step2 
	for chunk := nbChunks - 1; chunk >= 0; chunk-- {
		go bucketAccumulate(chunk, chTotals[chunk])
	}

	// reduce the buckets weigthed sums into our result
	return chunkReduce{{ toUpper $.PointName }}(p, c, chTotals[:])
	
}
{{end}}

{{/* note batch inversion for g2 elements with e2 that is curve specific is a bit more troublesome to implement */}}
{{if eq .PointName "g1"}}

// BatchJacobianToAffine{{ toUpper .PointName }} converts points in Jacobian coordinates to Affine coordinates
// performing a single field inversion (Montgomery batch inversion trick)
// result must be allocated with len(result) == len(points)
func BatchJacobianToAffine{{ toUpper .PointName }}(points []{{ toUpper .PointName}}Jac, result []{{ toUpper .PointName}}Affine) {
	debug.Assert(len(result) == len(points))
	zeroes := make([]bool, len(points))
	accumulator := fp.One()

	// mark all zero points to ignore them. 
	for i:=0; i < len(points); i++ {
		if points[i].Z.IsZero() {
			zeroes[i] = true
			continue
		}
		result[i].Y = accumulator
		accumulator.Mul(&accumulator, &points[i].Z)
	}

	var accInverse fp.Element
	accInverse.Inverse(&accumulator)

	for i := len(points) - 1; i >= 0; i-- {
		if zeroes[i] {
			// do nothing, X and Y are zeroes in affine.
			continue
		}
		result[i].Y.Mul(&result[i].Y, &accInverse)
		accInverse.Mul(&accInverse, &points[i].Z)
	}

	parallel.Execute( len(points), func(start, end int) {
		for i:=start; i < end; i++ {
			if zeroes[i] {
				// do nothing, X and Y are zeroes in affine.
				continue
			}
			var a, b fp.Element
			a = result[i].Y
			b.Square(&a)
			result[i].X.Mul(&points[i].X, &b)
			result[i].Y.Mul(&points[i].Y, &b).
				Mul(&result[i].Y, &a)
		}
	})

}
{{end}}

// BatchScalarMultiplication{{ toUpper .PointName }} multiplies the same base (generator) by all scalars
// and return resulting points in affine coordinates
// currently uses a simple windowed-NAF like exponentiation algorithm, and use fixed windowed size (16 bits)
// TODO : implement variable window size depending on input size
func BatchScalarMultiplication{{ toUpper .PointName }}(base *{{ toUpper .PointName}}Affine, scalars []fr.Element) []{{ toUpper .PointName }}Affine {
	const c = 16 // window size
	const nbChunks = fr.Limbs * 64 / c
	const selectorMask uint64 = (1 << c) - 1	// low c bits are 1
	const msbWindow uint32 = (1 << (c -1)) 

	// precompute all powers of base for our window
	var baseTable [(1<<(c-1))]{{ toUpper .PointName }}Jac
	baseTable[0].Set(&{{ toLower .PointName}}Infinity)
	baseTable[0].AddMixed(base)
	for i:=1;i<len(baseTable);i++ {
		baseTable[i] = baseTable[i-1]
		baseTable[i].AddMixed(base)
	}

	// convert our scalars to digits
	scalarsToDigits := func(scalars []fr.Element) (digits [][nbChunks]uint32) {
		const max = int(msbWindow)
		const twoc = (1 << c ) 
		digits = make([][nbChunks]uint32, len(scalars))
		
		// process the scalars in parallel
		parallel.Execute( len(scalars), func(start, end int) {
			for i:=start; i < end; i++ {
				var carry int

				// for each chunk in the scalar, compute the current digit, and an eventual carry
				for chunk := 0; chunk < nbChunks; chunk++ {
		
					// compute offset and word selector / shift to select the right bits of our windows
					jc := uint64(chunk * c)
					selectorIndex := jc / 64
					selectorShift := jc - (selectorIndex * 64)

					// init with carry if any
					digit := carry
					carry = 0

					// digit = value of the c-bit window
					digit += int((scalars[i][selectorIndex] & (selectorMask << selectorShift)) >> selectorShift)
					
					// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and substract
					// 2^{c} to the current digit, making it negative.
					if digit >= max {
						digit -= twoc
						carry = 1 
					}

					if digit == 0 {
						continue // digit[i][chunk] = 0
					}

					if digit > 0 {
						digits[i][chunk] = uint32(digit)
					} else {
						// mark negative sign using msbWindow mask, a bit we know is not used. 
						digits[i][chunk] = uint32(-digit - 1) | msbWindow
					}
					
				}
			}
		})
		return 
	}
	digits := scalarsToDigits(scalars)

	{{if eq .PointName "g1"}}
		toReturn := make([]{{ toUpper .PointName }}Jac, len(scalars))
	{{else}}
		toReturn := make([]{{ toUpper .PointName }}Affine, len(scalars))
	{{end}}

	// for each digit, take value in the base table, double it c time, voila.
	parallel.Execute( len(digits), func(start, end int) {
		var p {{ toUpper .PointName }}Jac
		for i:=start; i < end; i++ {
			p.Set(&{{ toLower .PointName}}Infinity)
			
			for chunk := nbChunks - 1; chunk >=0; chunk-- {
				if chunk != nbChunks -1 {
					for j:=0; j<c; j++ {
						p.DoubleAssign()
					}
				}

				bits := digits[i][chunk]
				if bits != 0 {
					if bits & msbWindow == 0 {
						// add 
						p.AddAssign(&baseTable[bits-1])
					} else {
						// sub 
						t := baseTable[bits & ^msbWindow]
						t.Neg(&t)
						p.AddAssign(&t)
					}
				}
			}

			// set our result point 
			{{if eq .PointName "g1"}}
				toReturn[i] = p
			{{else}}
				toReturn[i].FromJacobian(&p)
			{{end}}
			
		}
	})

	{{if eq .PointName "g1"}}
		toReturnAff := make([]{{ toUpper .PointName }}Affine, len(scalars))
		BatchJacobianToAffine{{ toUpper .PointName }}(toReturn, toReturnAff)
		return toReturnAff
	{{else}}
		return toReturn
	{{end}}
}

`
