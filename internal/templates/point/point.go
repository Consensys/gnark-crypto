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
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.doubleNeg(a)
	}
	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)
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
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.double(a)
	}
	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)
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

// Clone returns a copy of self
func (p *{{ toUpper .PointName }}Jac) Clone() *{{ toUpper .PointName }}Jac {
	return &{{ toUpper .PointName }}Jac{
		p.X, p.Y, p.Z,
	}
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
	return p.ScalarMulGLV(&{{ toLower .PointName }}GenAff, s)
}

// ScalarMultiplication algo for exponentiation
func (p *{{ toUpper .PointName }}Jac) ScalarMultiplication(a *{{ toUpper .PointName }}Affine, s *big.Int) *{{ toUpper .PointName }}Jac {

	var res {{ toUpper .PointName }}Jac
	res.Set(&{{ toLower .PointName }}Infinity)
	b := s.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 0; j < 8; j++ {
			res.DoubleAssign()
			if (w&mask)>>(7-j) != 0 {
				res.AddMixed(a)
			}
			mask = mask >> 1
		}
	}
	p.Set(&res)

	return p
}

// ScalarMulGLV performs scalar multiplication using GLV (without the lattice reduction)
func (p *{{ toUpper .PointName }}Jac) ScalarMulGLV(a *{{ toUpper .PointName }}Affine, s *big.Int) *{{ toUpper .PointName }}Jac {

	var {{ toLower .PointName}}, phi{{ toLower .PointName}}, res {{ toUpper .PointName}}Jac
	var phi{{ toLower .PointName}}Affine {{ toUpper .PointName}}Affine
	res.Set(&{{ toLower .PointName}}Infinity)
	{{ toLower .PointName}}.FromAffine(a)
	phi{{ toLower .PointName}}.Set(&{{ toLower .PointName}})
	{{- if eq .CoordType "fp.Element" }}
		phi{{ toLower .PointName}}.X.Mul(&phi{{ toLower .PointName}}.X, &thirdRootOne{{ toUpper .PointName}})
	{{- else if eq .CoordType "E2" }}
		phi{{ toLower .PointName}}.X.MulByElement(&phi{{ toLower .PointName}}.X, &thirdRootOne{{ toUpper .PointName}})
	{{- end }}

	phi{{ toLower .PointName}}Affine.FromJacobian(&phi{{ toLower .PointName}})

	// s = s1*lambda+s2
	var s1, s2 big.Int
	s1.DivMod(s, &lambdaGLV, &s2)


	// s1 part (on phi({{ toLower .PointName}})=lambda*{{ toLower .PointName}})
	phi{{ toLower .PointName}}.ScalarMultiplication(&phi{{ toLower .PointName}}Affine, &s1)

	// s2 part (on {{ toLower .PointName}})
	{{ toLower .PointName}}.ScalarMultiplication(a, &s2)

	res.AddAssign(&phi{{ toLower .PointName}})
	res.AddAssign(&{{ toLower .PointName}})

	p.Set(&res)

	return p
}


// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf 
func (p *{{ toUpper .PointName }}Jac) MultiExp(points []{{ toUpper .PointName }}Affine, scalars []fr.Element) chan {{ toUpper .PointName }}Jac {
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
	nbPoints := len(points)
	if nbPoints <= 100 {
		return p.multiExpc4(points, scalars)
	} else if nbPoints <= 10000 {
		return p.multiExpc8(points, scalars)
	} else if nbPoints <= 80000 {
		return p.multiExpc11(points, scalars)
	} else if nbPoints <= 400000 {
		return p.multiExpc13(points, scalars)
	} else if nbPoints < 8388608 {
		return p.multiExpc16(points, scalars)
	} else {
		return p.multiExpc18(points, scalars)
	}

}


{{ template "multiexp" dict "all" . "C" "4"}}
{{ template "multiexp" dict "all" . "C" "8"}}
{{ template "multiexp" dict "all" . "C" "11"}}
{{ template "multiexp" dict "all" . "C" "13"}}
{{ template "multiexp" dict "all" . "C" "14"}}
{{ template "multiexp" dict "all" . "C" "15"}}
{{ template "multiexp" dict "all" . "C" "16"}}
{{ template "multiexp" dict "all" . "C" "17"}}
{{ template "multiexp" dict "all" . "C" "18"}}


func chunkReduce{{ toUpper .PointName }}(p *{{ toUpper .PointName }}Jac, c int, chTotals []chan {{ toUpper .PointName }}Jac)  chan {{ toUpper .PointName }}Jac {
	chRes := make(chan {{ toUpper .PointName }}Jac, 1)
	debug.Assert(len(chTotals) >= 2)
	go func() {
		totalj := <-chTotals[len(chTotals)-1]
		p.Set(&totalj)
		for j := len(chTotals) - 2; j >= 0; j-- {
			for l := 0; l < c; l++ {
				p.DoubleAssign()
			}
			totalj := <-chTotals[j]
			p.AddAssign(&totalj)
		}
		
		chRes <- *p
		close(chRes)
	}()

	
	return chRes
}




{{ define "multiexp" }}
func (p *{{ toUpper .all.PointName }}Jac) multiExpc{{$.C}}(points []{{ toUpper .all.PointName }}Affine, scalars []fr.Element) chan {{ toUpper .all.PointName }}Jac {
	{{$cDividesBits := divides $.C $.all.RBitLen}}
	const c  = {{$.C}} 							// scalars partitioned into c-bit radixes
	const t = fr.Bits / c        			// number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1	// low c bits are 1
	const nbChunks = t {{if not $cDividesBits }} + 1 {{end}} // note: if c doesn't divide fr.Bits, nbChunks != t)

	
	scalarsToDigits := func(scalars []fr.Element) [][nbChunks]int {
		const max  = (1 << (c -1)) 
		const twoc = (1 << c ) 
		res := make([][nbChunks]int, len(scalars))
		
		parallel.Execute(0, len(scalars), func(start, end int) {
			for i:=start; i < end; i++ {
				var carry int
				// for each chunk, compute the current digit
				for chunk := 0; chunk < nbChunks; chunk++ {
		
					jc := uint64(chunk * c)
					selectorIndex := jc / 64
					selectorShift := jc - (selectorIndex * 64)
					selectedBits := selectorMask << selectorShift

					digit := carry
					carry = 0


					digit += int((scalars[i][selectorIndex] & selectedBits) >> selectorShift)
					{{$cDivides64 := divides $.C 64}}
					{{if not $cDivides64}}
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
					
					if digit >= (max) {
						digit -= twoc
						carry = 1 
					}
					res[i][chunk] = digit
				}
			}
		}, true)
		
		
		return res
	}


	// bucketAccumulate places points into buckets base on their selector and return the weighted bucket sum in given channel
	bucketAccumulate := func(chunk, c int, selectorMask uint64, points []{{ toUpper .all.PointName }}Affine, digits [][nbChunks]int, buckets []{{ toLower .all.PointName }}JacExtended, chRes chan<- {{ toUpper .all.PointName }}Jac) {
			
		for i := 0 ; i < len(buckets); i++ {
			buckets[i].SetInfinity()
		}

		// place points into buckets based on their selector
		for i := 0; i < len(digits); i++ {
			selector := (digits[i][chunk])
			if selector == 0 {
				continue
			} else if selector > 0 {
				buckets[selector-1].mAdd(&points[i])
			} else {
				buckets[-selector-1].mSub(&points[i])
			}
			
		}

		
		// reduce buckets into total
		// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

		var runningSum, tj, total {{ toUpper .all.PointName }}Jac
		runningSum.Set(&{{ toLower .all.PointName }}Infinity)
		total.Set(&{{ toLower .all.PointName }}Infinity)
		for k := len(buckets) - 1; k >= 0; k-- {
			if !buckets[k].ZZ.IsZero() {
				runningSum.AddAssign(buckets[k].unsafeToJac(&tj))
			}
			total.AddAssign(&runningSum)
		}
		

		chRes <- total
		close(chRes)
	} 


	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan {{ toUpper .all.PointName }}Jac
	for i:= 0; i< nbChunks; i++ {
		chTotals[i] = make(chan {{ toUpper .all.PointName }}Jac, 1)
	}

	digits := scalarsToDigits(scalars)

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [1<<(c-1)]{{ toLower .all.PointName }}JacExtended
			bucketAccumulate(chunk, c, selectorMask, points, digits, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduce{{ toUpper .all.PointName }}(p, c, chTotals[:])
	
}
{{ end }}

`

const Backup = `
// note: keeping that around for now, in case we need to explore 64%c != 0

`
