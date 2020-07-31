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
	Q.X.Inverse(&p.ZZ).MulAssign(&p.X)
	Q.Y.Inverse(&p.ZZZ).MulAssign(&p.Y)
	return Q
}

// ToJac sets p in affine coords
func (p *{{ toLower .PointName }}JacExtended) ToJac(Q *{{ toUpper .PointName }}Jac) *{{ toUpper .PointName }}Jac {
	var zero {{.CoordType}}
	if p.ZZ.Equal(&zero) {
		Q.Set(&{{ toLower .PointName }}Infinity)
		return Q
	}
	Q.X.Mul(&p.ZZ, &p.X).MulAssign(&p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).MulAssign(&p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
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
	Q2.AddAssign(&Q).AddAssign(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).MulAssign(&R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.MulAssign(&PP)
	p.ZZZ.MulAssign(&PPP)

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
		AddAssign(&_M) // -> + a, but a=0 here
	p.X.Square(&M).
		SubAssign(&S).
		SubAssign(&S)
	Y3.Sub(&S, &p.X).MulAssign(&M)
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
		MulAssign(&Z2Z2)
	S2.Mul(&p.Y, &a.Z).
		MulAssign(&Z1Z1)

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
		SubAssign(&J).
		SubAssign(&V).
		SubAssign(&V)
	p.Y.Sub(&V, &p.X).
		MulAssign(&r)
	S1.MulAssign(&J).Double(&S1)
	p.Y.SubAssign(&S1)
	p.Z.AddAssign(&a.Z)
	p.Z.Square(&p.Z).
		SubAssign(&Z1Z1).
		SubAssign(&Z2Z2).
		MulAssign(&H)

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
		MulAssign(&Z1Z1)

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
		SubAssign(&J).
		SubAssign(&V).
		SubAssign(&V)
	J.MulAssign(&p.Y).Double(&J)
	p.Y.Sub(&V, &p.X).
		MulAssign(&r)
	p.Y.SubAssign(&J)
	p.Z.AddAssign(&H)
	p.Z.Square(&p.Z).
		SubAssign(&Z1Z1).
		SubAssign(&HH)

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
		SubAssign(&XX).
		SubAssign(&YYYY).
		Double(&S)
	M.Double(&XX).AddAssign(&XX)
	p.Z.AddAssign(&p.Y).
		Square(&p.Z).
		SubAssign(&YY).
		SubAssign(&ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.SubAssign(&T)
	p.Y.Sub(&S, &p.X).
		MulAssign(&M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.SubAssign(&YYYY)

	return p
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

	chTasks := []chan struct{}{
		make(chan struct{}),
		make(chan struct{}),
	}

	// s1 part (on phi({{ toLower .PointName}})=lambda*{{ toLower .PointName}})
	go func() {
		phi{{ toLower .PointName}}._doubleandadd(&phi{{ toLower .PointName}}Affine, &s1)
		chTasks[0] <- struct{}{}
	}()

	// s2 part (on {{ toLower .PointName}})
	go func() {
		{{ toLower .PointName}}._doubleandadd(a, &s2)
		chTasks[1] <- struct{}{}
	}()

	<-chTasks[0]
	res.AddAssign(&phi{{ toLower .PointName}})
	<-chTasks[1]
	res.AddAssign(&{{ toLower .PointName}})

	p.Set(&res)

	return p
}

// MultiExp complexity O(n)
func (p *{{ toUpper .PointName }}Jac) MultiExp(points []{{ toUpper .PointName }}Affine, scalars []fr.Element) chan {{ toUpper .PointName }}Jac {

	nbPoints := len(points)
	debug.Assert(nbPoints == len(scalars))

	chRes := make(chan {{ toUpper .PointName }}Jac, 1)

	// under 50 points, the windowed multi exp performs better
	const minPoints = 50
	if nbPoints <= minPoints {
		var tmp {{ toUpper .PointName }}Jac
		var s big.Int
		p.Set(&{{ toLower .PointName }}Infinity)
		for i := 0; i < nbPoints; i++ {
			scalars[i].ToBigInt(&s)
			tmp.ScalarMulGLV(&points[i], &s)
			p.AddAssign(&tmp)
		}
	}

	// empirical values
	var nbChunks, chunkSize int
	var mask uint64
	if nbPoints <= 10000 {
		chunkSize = 8
	} else if nbPoints <= 80000 {
		chunkSize = 11
	} else if nbPoints <= 400000 {
		chunkSize = 13
	} else if nbPoints <= 800000 {
		chunkSize = 14
	} else {
		chunkSize = 16
	}

	const sizeScalar = fr.ElementLimbs * 64

	var bitsForTask [][]int
	if sizeScalar%chunkSize == 0 {
		counter := sizeScalar - 1
		nbChunks = sizeScalar / chunkSize
		bitsForTask = make([][]int, nbChunks)
		for i := 0; i < nbChunks; i++ {
			bitsForTask[i] = make([]int, chunkSize)
			for j := 0; j < chunkSize; j++ {
				bitsForTask[i][j] = counter
				counter--
			}
		}
	} else {
		counter := sizeScalar - 1
		nbChunks = sizeScalar/chunkSize + 1
		bitsForTask = make([][]int, nbChunks)
		for i := 0; i < nbChunks; i++ {
			if i < nbChunks-1 {
				bitsForTask[i] = make([]int, chunkSize)
			} else {
				bitsForTask[i] = make([]int, sizeScalar%chunkSize)
			}
			for j := 0; j < chunkSize && counter >= 0; j++ {
				bitsForTask[i][j] = counter
				counter--
			}
		}
	}

	accumulators := make([]{{ toUpper .PointName }}Jac, nbChunks)
	chIndices := make([]chan struct{}, nbChunks)
	chPoints := make([]chan struct{}, nbChunks)
	for i := 0; i < nbChunks; i++ {
		chIndices[i] = make(chan struct{}, 1)
		chPoints[i] = make(chan struct{}, 1)
	}

	mask = (1 << chunkSize) - 1
	nbPointsPerSlots := nbPoints / int(mask)
	// [][] is more efficient than [][][] for storage, elements are accessed via i*nbChunks+k
	indices := make([][]int, int(mask)*nbChunks)
	for i := 0; i < int(mask)*nbChunks; i++ {
		indices[i] = make([]int, 0, nbPointsPerSlots)
	}

	// if chunkSize=8, nbChunks=32 (the scalars are chunkSize*nbChunks bits long)
	// for each 32 chunk, there is a list of 2**8=256 list of indices
	// for the i-th chunk, accumulateIndices stores in the k-th list all the indices of points
	// for which the i-th chunk of 8 bits is equal to k
	accumulateIndices := func(cpuID, nbTasks, n int) {
		for i := 0; i < nbTasks; i++ {
			task := cpuID + i*n
			idx := task*int(mask) - 1
			for j := 0; j < nbPoints; j++ {
				val := 0
				for k := 0; k < len(bitsForTask[task]); k++ {
					val = val << 1
					c := bitsForTask[task][k] / int(64)
					o := bitsForTask[task][k] % int(64)
					b := (scalars[j][c] >> o) & 1
					val += int(b)
				}
				if val != 0 {
					indices[idx+int(val)] = append(indices[idx+int(val)], j)
				}
			}
			chIndices[task] <- struct{}{}
			close(chIndices[task])
		}
	}

	// if chunkSize=8, nbChunks=32 (the scalars are chunkSize*nbChunks bits long)
	// for each chunk, sum up elements in index 0, add to current result, sum up elements
	// in index 1, add to current result, etc, up to 255=2**8-1
	accumulatePoints := func(cpuID, nbTasks, n int) {
		for i := 0; i < nbTasks; i++ {
			var tmp {{ toLower .PointName }}JacExtended
			var _tmp {{ toUpper .PointName }}Jac
			task := cpuID + i*n

			// init points
			tmp.SetInfinity()
			accumulators[task].Set(&{{ toLower .PointName }}Infinity)

			// wait for indices to be ready
			<-chIndices[task]

			for j := int(mask - 1); j >= 0; j-- {
				for _, k := range indices[task*int(mask)+j] {
					tmp.mAdd(&points[k])
				}
				tmp.ToJac(&_tmp)
				accumulators[task].AddAssign(&_tmp)
			}
			chPoints[task] <- struct{}{}
			close(chPoints[task])
		}
	}

	// double and add algo to collect all small reductions
	reduce := func() {
		var res {{ toUpper .PointName }}Jac
		res.Set(&{{ toLower .PointName }}Infinity)
		for i := 0; i < nbChunks; i++ {
			for j := 0; j < len(bitsForTask[i]); j++ {
				res.DoubleAssign()
			}
			<-chPoints[i]
			res.AddAssign(&accumulators[i])
		}
		p.Set(&res)
		chRes <- *p
	}

	nbCpus := runtime.NumCPU()
	nbTasksPerCpus := nbChunks / nbCpus
	remainingTasks := nbChunks % nbCpus
	for i := 0; i < nbCpus; i++ {
		if remainingTasks > 0 {
			go accumulateIndices(i, nbTasksPerCpus+1, nbCpus)
			go accumulatePoints(i, nbTasksPerCpus+1, nbCpus)
			remainingTasks--
		} else {
			go accumulateIndices(i, nbTasksPerCpus, nbCpus)
			go accumulatePoints(i, nbTasksPerCpus, nbCpus)
		}
	}

	go reduce()

	return chRes
}

`
