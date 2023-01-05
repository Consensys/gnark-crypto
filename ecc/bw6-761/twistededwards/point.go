// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package twistededwards

import (
	"crypto/subtle"
	"io"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// PointAffine point on a twisted Edwards curve
type PointAffine struct {
	X, Y fr.Element
}

// PointProj point in projective coordinates
type PointProj struct {
	X, Y, Z fr.Element
}

// PointExtended point in extended coordinates
type PointExtended struct {
	X, Y, Z, T fr.Element
}

const (
	//following https://tools.ietf.org/html/rfc8032#section-3.1,
	// an fr element x is negative if its binary encoding is
	// lexicographically larger than -x.
	mCompressedNegative = 0x80
	mCompressedPositive = 0x00
	mUnmask             = 0x7f

	// size in byte of a compressed point (point.Y --> fr.Element)
	sizePointCompressed = fr.Bytes
)

// Bytes returns the compressed point as a byte array
// Follows https://tools.ietf.org/html/rfc8032#section-3.1,
// as the twisted Edwards implementation is primarily used
// for eddsa.
func (p *PointAffine) Bytes() [sizePointCompressed]byte {

	var res [sizePointCompressed]byte
	var mask uint

	y := p.Y.Bytes()

	if p.X.LexicographicallyLargest() {
		mask = mCompressedNegative
	} else {
		mask = mCompressedPositive
	}
	// p.Y must be in little endian
	y[0] |= byte(mask) // msb of y
	for i, j := 0, sizePointCompressed-1; i < j; i, j = i+1, j-1 {
		y[i], y[j] = y[j], y[i]
	}
	subtle.ConstantTimeCopy(1, res[:], y[:])
	return res
}

// Marshal converts p to a byte slice
func (p *PointAffine) Marshal() []byte {
	b := p.Bytes()
	return b[:]
}

func computeX(y *fr.Element) (x fr.Element) {
	initOnce.Do(initCurveParams)

	var one, num, den fr.Element
	one.SetOne()
	num.Square(y)
	den.Mul(&num, &curveParams.D)
	num.Sub(&one, &num)
	den.Sub(&curveParams.A, &den)
	x.Div(&num, &den)
	x.Sqrt(&x)
	return
}

// SetBytes sets p from buf
// len(buf) >= sizePointCompressed
// buf contains the Y coordinate masked with a parity bit to recompute the X coordinate
// from the curve equation. See Bytes() and https://tools.ietf.org/html/rfc8032#section-3.1
// Returns the number of read bytes and an error if the buffer is too short.
func (p *PointAffine) SetBytes(buf []byte) (int, error) {

	if len(buf) < sizePointCompressed {
		return 0, io.ErrShortBuffer
	}
	bufCopy := make([]byte, sizePointCompressed)
	subtle.ConstantTimeCopy(1, bufCopy, buf[:sizePointCompressed])
	for i, j := 0, sizePointCompressed-1; i < j; i, j = i+1, j-1 {
		bufCopy[i], bufCopy[j] = bufCopy[j], bufCopy[i]
	}
	isLexicographicallyLargest := (mCompressedNegative&bufCopy[0])>>7 == 1
	bufCopy[0] &= mUnmask
	p.Y.SetBytes(bufCopy)
	p.X = computeX(&p.Y)
	if isLexicographicallyLargest {
		if !p.X.LexicographicallyLargest() {
			p.X.Neg(&p.X)
		}
	} else {
		if p.X.LexicographicallyLargest() {
			p.X.Neg(&p.X)
		}
	}

	return sizePointCompressed, nil
}

// Unmarshal alias to SetBytes()
func (p *PointAffine) Unmarshal(b []byte) error {
	_, err := p.SetBytes(b)
	return err
}

// Set sets p to p1 and return it
func (p *PointAffine) Set(p1 *PointAffine) *PointAffine {
	p.X.Set(&p1.X)
	p.Y.Set(&p1.Y)
	return p
}

// Equal returns true if p=p1 false otherwise
func (p *PointAffine) Equal(p1 *PointAffine) bool {
	return p.X.Equal(&p1.X) && p.Y.Equal(&p1.Y)
}

// IsZero returns true if p=0 false otherwise
func (p *PointAffine) IsZero() bool {
	var one fr.Element
	one.SetOne()
	return p.X.IsZero() && p.Y.Equal(&one)
}

// NewPointAffine creates a new instance of PointAffine
func NewPointAffine(x, y fr.Element) PointAffine {
	return PointAffine{x, y}
}

// IsOnCurve checks if a point is on the twisted Edwards curve
func (p *PointAffine) IsOnCurve() bool {

	ecurve := GetEdwardsCurve()

	var lhs, rhs, tmp fr.Element

	tmp.Mul(&p.Y, &p.Y)
	lhs.Mul(&p.X, &p.X)
	mulByA(&lhs)
	lhs.Add(&lhs, &tmp)

	tmp.Mul(&p.X, &p.X).
		Mul(&tmp, &p.Y).
		Mul(&tmp, &p.Y).
		Mul(&tmp, &ecurve.D)
	rhs.SetOne().Add(&rhs, &tmp)

	return lhs.Equal(&rhs)
}

// Neg sets p to -p1 and returns it
func (p *PointAffine) Neg(p1 *PointAffine) *PointAffine {
	p.Set(p1)
	p.X.Neg(&p.X)
	return p
}

// Add adds two points (x,y), (u,v) on a twisted Edwards curve with parameters a, d
// modifies p
func (p *PointAffine) Add(p1, p2 *PointAffine) *PointAffine {

	ecurve := GetEdwardsCurve()

	var xu, yv, xv, yu, dxyuv, one, denx, deny fr.Element
	pRes := new(PointAffine)
	xv.Mul(&p1.X, &p2.Y)
	yu.Mul(&p1.Y, &p2.X)
	pRes.X.Add(&xv, &yu)

	xu.Mul(&p1.X, &p2.X)
	mulByA(&xu)
	yv.Mul(&p1.Y, &p2.Y)
	pRes.Y.Sub(&yv, &xu)

	dxyuv.Mul(&xv, &yu).Mul(&dxyuv, &ecurve.D)
	one.SetOne()
	denx.Add(&one, &dxyuv)
	deny.Sub(&one, &dxyuv)

	p.X.Div(&pRes.X, &denx)
	p.Y.Div(&pRes.Y, &deny)

	return p
}

// Double doubles point (x,y) on a twisted Edwards curve with parameters a, d
// modifies p
func (p *PointAffine) Double(p1 *PointAffine) *PointAffine {

	p.Set(p1)
	var xx, yy, xy, denum, two fr.Element

	xx.Square(&p.X)
	yy.Square(&p.Y)
	xy.Mul(&p.X, &p.Y)
	mulByA(&xx)
	denum.Add(&xx, &yy)

	p.X.Double(&xy).Div(&p.X, &denum)

	two.SetOne().Double(&two)
	denum.Neg(&denum).Add(&denum, &two)

	p.Y.Sub(&yy, &xx).Div(&p.Y, &denum)

	return p
}

// FromProj sets p in affine from p in projective
func (p *PointAffine) FromProj(p1 *PointProj) *PointAffine {
	var I fr.Element
	I.Inverse(&p1.Z)
	p.X.Mul(&p1.X, &I)
	p.Y.Mul(&p1.Y, &I)
	return p
}

// FromExtended sets p in affine from p in extended coordinates
func (p *PointAffine) FromExtended(p1 *PointExtended) *PointAffine {
	var I fr.Element
	I.Inverse(&p1.Z)
	p.X.Mul(&p1.X, &I)
	p.Y.Mul(&p1.Y, &I)
	return p
}

// ScalarMultiplication scalar multiplication of a point
// p1 in affine coordinates with a scalar in big.Int
func (p *PointAffine) ScalarMultiplication(p1 *PointAffine, scalar *big.Int) *PointAffine {

	var p1Extended, resExtended PointExtended
	p1Extended.FromAffine(p1)
	resExtended.ScalarMultiplication(&p1Extended, scalar)
	p.FromExtended(&resExtended)

	return p
}

// setInfinity sets p to O (0:1)
func (p *PointAffine) setInfinity() *PointAffine {
	p.X.SetZero()
	p.Y.SetOne()
	return p
}

//-------- Projective coordinates

// Set sets p to p1 and return it
func (p *PointProj) Set(p1 *PointProj) *PointProj {
	p.X.Set(&p1.X)
	p.Y.Set(&p1.Y)
	p.Z.Set(&p1.Z)
	return p
}

// setInfinity sets p to O (0:1:1)
func (p *PointProj) setInfinity() *PointProj {
	p.X.SetZero()
	p.Y.SetOne()
	p.Z.SetOne()
	return p
}

// Equal returns true if p=p1 false otherwise
// If one point is on the affine chart Z=0 it returns false
func (p *PointProj) Equal(p1 *PointProj) bool {
	if p.Z.IsZero() || p1.Z.IsZero() {
		return false
	}
	var pAffine, p1Affine PointAffine
	pAffine.FromProj(p)
	p1Affine.FromProj(p1)
	return pAffine.Equal(&p1Affine)
}

// IsZero returns true if p=0 false otherwise
func (p *PointProj) IsZero() bool {
	return p.X.IsZero() && p.Y.Equal(&p.Z)
}

// Neg negates point (x,y) on a twisted Edwards curve with parameters a, d
// modifies p
func (p *PointProj) Neg(p1 *PointProj) *PointProj {
	p.Set(p1)
	p.X.Neg(&p.X)
	return p
}

// FromAffine sets p in projective from p in affine
func (p *PointProj) FromAffine(p1 *PointAffine) *PointProj {
	p.X.Set(&p1.X)
	p.Y.Set(&p1.Y)
	p.Z.SetOne()
	return p
}

// MixedAdd adds a point in projective to a point in affine coordinates
// cf https://hyperelliptic.org/EFD/g1p/auto-twisted-projective.html#addition-madd-2008-bbjlp
func (p *PointProj) MixedAdd(p1 *PointProj, p2 *PointAffine) *PointProj {

	ecurve := GetEdwardsCurve()

	var B, C, D, E, F, G, H, I fr.Element
	B.Square(&p1.Z)
	C.Mul(&p1.X, &p2.X)
	D.Mul(&p1.Y, &p2.Y)
	E.Mul(&ecurve.D, &C).Mul(&E, &D)
	F.Sub(&B, &E)
	G.Add(&B, &E)
	H.Add(&p1.X, &p1.Y)
	I.Add(&p2.X, &p2.Y)
	p.X.Mul(&H, &I).
		Sub(&p.X, &C).
		Sub(&p.X, &D).
		Mul(&p.X, &p1.Z).
		Mul(&p.X, &F)
	mulByA(&C)
	p.Y.Sub(&D, &C).
		Mul(&p.Y, &p1.Z).
		Mul(&p.Y, &G)
	p.Z.Mul(&F, &G)

	return p
}

// Double adds points in projective coordinates
// cf https://hyperelliptic.org/EFD/g1p/auto-twisted-projective.html#doubling-dbl-2008-bbjlp
func (p *PointProj) Double(p1 *PointProj) *PointProj {

	var B, C, D, E, F, H, J fr.Element

	B.Add(&p1.X, &p1.Y).Square(&B)
	C.Square(&p1.X)
	D.Square(&p1.Y)
	E.Set(&C)
	mulByA(&E)
	F.Add(&E, &D)
	H.Square(&p1.Z)
	J.Sub(&F, &H).Sub(&J, &H)
	p.X.Sub(&B, &C).
		Sub(&p.X, &D).
		Mul(&p.X, &J)
	p.Y.Sub(&E, &D).Mul(&p.Y, &F)
	p.Z.Mul(&F, &J)

	return p
}

// Add adds points in projective coordinates
// cf https://hyperelliptic.org/EFD/g1p/auto-twisted-projective.html#addition-add-2008-bbjlp
func (p *PointProj) Add(p1, p2 *PointProj) *PointProj {

	ecurve := GetEdwardsCurve()

	var A, B, C, D, E, F, G, H, I fr.Element
	A.Mul(&p1.Z, &p2.Z)
	B.Square(&A)
	C.Mul(&p1.X, &p2.X)
	D.Mul(&p1.Y, &p2.Y)
	E.Mul(&ecurve.D, &C).Mul(&E, &D)
	F.Sub(&B, &E)
	G.Add(&B, &E)
	H.Add(&p1.X, &p1.Y)
	I.Add(&p2.X, &p2.Y)
	p.X.Mul(&H, &I).
		Sub(&p.X, &C).
		Sub(&p.X, &D).
		Mul(&p.X, &A).
		Mul(&p.X, &F)
	mulByA(&C)
	C.Neg(&C)
	p.Y.Add(&D, &C).
		Mul(&p.Y, &A).
		Mul(&p.Y, &G)
	p.Z.Mul(&F, &G)

	return p
}

// ScalarMultiplication scalar multiplication of a point
// p1 in projective coordinates with a scalar in big.Int
func (p *PointProj) ScalarMultiplication(p1 *PointProj, scalar *big.Int) *PointProj {
	var _scalar big.Int
	_scalar.Set(scalar)
	p.Set(p1)
	if _scalar.Sign() == -1 {
		_scalar.Neg(&_scalar)
		p.Neg(p)
	}
	var resProj PointProj
	resProj.setInfinity()
	const wordSize = bits.UintSize
	sWords := _scalar.Bits()

	for i := len(sWords) - 1; i >= 0; i-- {
		ithWord := sWords[i]
		for k := 0; k < wordSize; k++ {
			resProj.Double(&resProj)
			kthBit := (ithWord >> (wordSize - 1 - k)) & 1
			if kthBit == 1 {
				resProj.Add(&resProj, p)
			}
		}
	}

	p.Set(&resProj)
	return p
}

// ------- Extended coordinates

// Set sets p to p1 and return it
func (p *PointExtended) Set(p1 *PointExtended) *PointExtended {
	p.X.Set(&p1.X)
	p.Y.Set(&p1.Y)
	p.T.Set(&p1.T)
	p.Z.Set(&p1.Z)
	return p
}

// IsZero returns true if p=0 false otherwise
func (p *PointExtended) IsZero() bool {
	return p.X.IsZero() && p.Y.Equal(&p.Z) && p.T.IsZero()
}

// Equal returns true if p=p1 false otherwise
// If one point is on the affine chart Z=0 it returns false
func (p *PointExtended) Equal(p1 *PointExtended) bool {
	if p.Z.IsZero() || p1.Z.IsZero() {
		return false
	}
	var pAffine, p1Affine PointAffine
	pAffine.FromExtended(p)
	p1Affine.FromExtended(p1)
	return pAffine.Equal(&p1Affine)
}

// Neg negates point (x,y) on a twisted Edwards curve with parameters a, d
// modifies p
func (p *PointExtended) Neg(p1 *PointExtended) *PointExtended {
	p.Set(p1)
	p.X.Neg(&p.X)
	p.T.Neg(&p.T)
	return p
}

// FromAffine sets p in projective from p in affine
func (p *PointExtended) FromAffine(p1 *PointAffine) *PointExtended {
	p.X.Set(&p1.X)
	p.Y.Set(&p1.Y)
	p.Z.SetOne()
	p.T.Mul(&p1.X, &p1.Y)
	return p
}

// Add adds points in extended coordinates
// See https://hyperelliptic.org/EFD/g1p/auto-twisted-extended.html#addition-add-2008-hwcd-2
func (p *PointExtended) Add(p1, p2 *PointExtended) *PointExtended {
	if p1.Equal(p2) {
		p.Double(p1)
		return p
	}

	var A, B, C, D, E, F, G, H, tmp fr.Element
	A.Mul(&p1.X, &p2.X)
	B.Mul(&p1.Y, &p2.Y)
	C.Mul(&p1.Z, &p2.T)
	D.Mul(&p1.T, &p2.Z)
	E.Add(&D, &C)
	tmp.Sub(&p1.X, &p1.Y)
	F.Add(&p2.X, &p2.Y).
		Mul(&F, &tmp).
		Add(&F, &B).
		Sub(&F, &A)
	G.Set(&A)
	mulByA(&G)
	G.Add(&G, &B)
	H.Sub(&D, &C)

	p.X.Mul(&E, &F)
	p.Y.Mul(&G, &H)
	p.T.Mul(&E, &H)
	p.Z.Mul(&F, &G)

	return p
}

// MixedAdd adds a point in extended coordinates to a point in affine coordinates
// See https://hyperelliptic.org/EFD/g1p/auto-twisted-extended.html#addition-madd-2008-hwcd-2
func (p *PointExtended) MixedAdd(p1 *PointExtended, p2 *PointAffine) *PointExtended {
	var A, B, C, D, E, F, G, H, tmp fr.Element

	A.Mul(&p2.X, &p1.Z)
	B.Mul(&p2.Y, &p1.Z)

	if p1.X.Equal(&A) && p1.Y.Equal(&B) {
		p.MixedDouble(p1)
		return p
	}

	A.Mul(&p1.X, &p2.X)
	B.Mul(&p1.Y, &p2.Y)
	C.Mul(&p1.Z, &p2.X).
		Mul(&C, &p2.Y)
	D.Set(&p1.T)
	E.Add(&D, &C)
	tmp.Sub(&p1.X, &p1.Y)
	F.Add(&p2.X, &p2.Y).
		Mul(&F, &tmp).
		Add(&F, &B).
		Sub(&F, &A)
	G.Set(&A)
	mulByA(&G)
	G.Add(&G, &B)
	H.Sub(&D, &C)

	p.X.Mul(&E, &F)
	p.Y.Mul(&G, &H)
	p.T.Mul(&E, &H)
	p.Z.Mul(&F, &G)

	return p
}

// Double adds points in extended coordinates
// Dedicated doubling
// https://hyperelliptic.org/EFD/g1p/auto-twisted-extended-1.html#doubling-dbl-2008-hwcd
func (p *PointExtended) Double(p1 *PointExtended) *PointExtended {

	var A, B, C, D, E, F, G, H fr.Element

	A.Square(&p1.X)
	B.Square(&p1.Y)
	C.Square(&p1.Z).
		Double(&C)
	D.Set(&A)
	mulByA(&D)
	E.Add(&p1.X, &p1.Y).
		Square(&E).
		Sub(&E, &A).
		Sub(&E, &B)
	G.Add(&D, &B)
	F.Sub(&G, &C)
	H.Sub(&D, &B)

	p.X.Mul(&E, &F)
	p.Y.Mul(&G, &H)
	p.T.Mul(&H, &E)
	p.Z.Mul(&F, &G)

	return p
}

// MixedDouble adds points in extended coordinates
// Dedicated mixed doubling
// https://hyperelliptic.org/EFD/g1p/auto-twisted-extended-1.html#doubling-mdbl-2008-hwcd
func (p *PointExtended) MixedDouble(p1 *PointExtended) *PointExtended {

	var A, B, D, E, G, H, two fr.Element
	two.SetUint64(2)

	A.Square(&p1.X)
	B.Square(&p1.Y)
	D.Set(&A)
	mulByA(&D)
	E.Add(&p1.X, &p1.Y).
		Square(&E).
		Sub(&E, &A).
		Sub(&E, &B)
	G.Add(&D, &B)
	H.Sub(&D, &B)

	p.X.Sub(&G, &two).
		Mul(&p.X, &E)
	p.Y.Mul(&G, &H)
	p.T.Mul(&H, &E)
	p.Z.Square(&G).
		Sub(&p.Z, &G).
		Sub(&p.Z, &G)

	return p
}

// setInfinity sets p to O (0:1:1:0)
func (p *PointExtended) setInfinity() *PointExtended {
	p.X.SetZero()
	p.Y.SetOne()
	p.Z.SetOne()
	p.T.SetZero()
	return p
}

// ScalarMultiplication scalar multiplication of a point
// p1 in extended coordinates with a scalar in big.Int
func (p *PointExtended) ScalarMultiplication(p1 *PointExtended, scalar *big.Int) *PointExtended {
	var _scalar big.Int
	_scalar.Set(scalar)
	p.Set(p1)
	if _scalar.Sign() == -1 {
		_scalar.Neg(&_scalar)
		p.Neg(p)
	}
	var resExtended PointExtended
	resExtended.setInfinity()
	const wordSize = bits.UintSize
	sWords := _scalar.Bits()

	for i := len(sWords) - 1; i >= 0; i-- {
		ithWord := sWords[i]
		for k := 0; k < wordSize; k++ {
			resExtended.Double(&resExtended)
			kthBit := (ithWord >> (wordSize - 1 - k)) & 1
			if kthBit == 1 {
				resExtended.Add(&resExtended, p)
			}
		}
	}

	p.Set(&resExtended)
	return p
}
