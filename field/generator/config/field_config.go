// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package config provides Golang code generation for efficient field arithmetic operations.
package config

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/internal/addchain"
)

var (
	errParseModulus = errors.New("can't parse modulus")
)

// Field precomputed values used in template for code generation of field element APIs
type Field struct {
	PackageName               string
	ElementName               string
	ModulusBig                *big.Int
	Modulus                   string
	ModulusHex                string
	NbWords                   int
	NbBits                    int
	NbBytes                   int
	NbWordsLastIndex          int
	NbWordsIndexesNoZero      []int
	NbWordsIndexesFull        []int
	P20InversionCorrectiveFac []uint64
	P20InversionNbIterations  int
	UsingP20Inverse           bool
	IsMSWSaturated            bool // indicates if the most significant word is 0xFFFFF...FFFF
	Q                         []uint64
	QInverse                  []uint64
	QMinusOneHalvedP          []uint64 // ((q-1) / 2 ) + 1
	Mu                        uint64   // mu = 2^288 / q for 4.5 word barrett reduction
	RSquare                   []uint64
	One, Eleven, Thirteen     []uint64
	LegendreExponent          string // big.Int to base16 string
	NoCarry                   bool
	NoCarrySquare             bool // used if NoCarry is set, but some op may overflow in square optimization
	SqrtQ3Mod4                bool
	SqrtAtkin                 bool
	SqrtTonelliShanks         bool
	SqrtE                     uint64
	SqrtS                     []uint64
	SqrtAtkinExponent         string   // big.Int to base16 string
	SqrtSMinusOneOver2        string   // big.Int to base16 string
	SqrtQ3Mod4Exponent        string   // big.Int to base16 string
	SqrtG                     []uint64 // NonResidue ^  SqrtR (montgomery form)
	NonResidue                big.Int  // (montgomery form)
	LegendreExponentData      *addchain.AddChainData
	SqrtAtkinExponentData     *addchain.AddChainData
	SqrtSMinusOneOver2Data    *addchain.AddChainData
	SqrtQ3Mod4ExponentData    *addchain.AddChainData
	UseAddChain               bool

	Word Word // 32 iff Q < 2^32, else 64
	F31  bool // 31 bits field

	// asm code generation
	GenerateOpsAMD64       bool
	GenerateOpsARM64       bool
	GenerateVectorOpsAMD64 bool
	GenerateVectorOpsARM64 bool

	ASMPackagePath string
}

type Word struct {
	BitSize   int    // 32 or 64
	ByteSize  int    // 4 or 8
	TypeLower string // uint32 or uint64
	TypeUpper string // Uint32 or Uint64
	Add       string // Add64 or Add32
	Sub       string // Sub64 or Sub32
	Len       string // Len64 or Len32
}

// NewFieldConfig returns a data structure with needed information to generate apis for field element
//
// See field/generator package
func NewFieldConfig(packageName, elementName, modulus string, useAddChain bool) (*Field, error) {
	// parse modulus
	var bModulus big.Int
	if _, ok := bModulus.SetString(modulus, 0); !ok {
		return nil, errParseModulus
	}

	// field info
	F := &Field{
		PackageName: packageName,
		ElementName: elementName,
		Modulus:     bModulus.Text(10),
		ModulusHex:  bModulus.Text(16),
		ModulusBig:  new(big.Int).Set(&bModulus),
		UseAddChain: useAddChain,
	}
	// pre compute field constants
	F.NbBits = bModulus.BitLen()
	// note: here we set F31 only for BabyBear and KoalaBear;
	// we could do uint32 bit size for all fields with NbBits <= 31, but we keep it as is for now
	// to avoid breaking changes
	F.F31 = F.ModulusHex == "7f000001" || F.ModulusHex == "78000001" // F.NbBits <= 31
	F.NbWords = len(bModulus.Bits())
	F.NbWordsLastIndex = F.NbWords - 1

	// set q from big int repr
	F.Q = toUint64Slice(&bModulus)
	F.IsMSWSaturated = F.Q[len(F.Q)-1] == math.MaxUint64
	_qHalved := big.NewInt(0)
	bOne := new(big.Int).SetUint64(1)
	_qHalved.Sub(&bModulus, bOne).Rsh(_qHalved, 1).Add(_qHalved, bOne)
	F.QMinusOneHalvedP = toUint64Slice(_qHalved, F.NbWords)

	// Word size; we pick uint32 only if the modulus is less than 2^32
	F.Word.BitSize = 64
	F.Word.ByteSize = 8
	F.Word.TypeLower = "uint64"
	F.Word.TypeUpper = "Uint64"
	F.Word.Add = "Add64"
	F.Word.Sub = "Sub64"
	F.Word.Len = "Len64"
	if F.F31 {
		F.Word.BitSize = 32
		F.Word.ByteSize = 4
		F.Word.TypeLower = "uint32"
		F.Word.TypeUpper = "Uint32"
		F.Word.Add = "Add32"
		F.Word.Sub = "Sub32"
		F.Word.Len = "Len32"
	}

	F.NbBytes = F.NbWords * F.Word.ByteSize

	//  setting qInverse
	radix := uint(F.Word.BitSize)

	_r := big.NewInt(1)
	_r.Lsh(_r, uint(F.NbWords)*radix)
	_rInv := big.NewInt(1)
	_qInv := big.NewInt(0)
	extendedEuclideanAlgo(_r, &bModulus, _rInv, _qInv)
	_qInv.Mod(_qInv, _r)
	F.QInverse = toUint64Slice(_qInv, F.NbWords)

	// Pornin20 inversion correction factors
	k := 32 // Optimized for 64 bit machines, still works for 32

	p20InvInnerLoopNbIterations := 2*F.NbBits - 1
	// if constant time inversion then p20InvInnerLoopNbIterations-- (among other changes)
	F.P20InversionNbIterations = (p20InvInnerLoopNbIterations-1)/(k-1) + 1 // ⌈ (2 * field size - 1) / (k-1) ⌉
	F.P20InversionNbIterations += F.P20InversionNbIterations % 2           // "round up" to a multiple of 2

	kLimbs := k * F.NbWords
	p20InversionCorrectiveFacPower := kLimbs*6 + F.P20InversionNbIterations*(kLimbs-k+1)
	p20InversionCorrectiveFac := big.NewInt(1)
	p20InversionCorrectiveFac.Lsh(p20InversionCorrectiveFac, uint(p20InversionCorrectiveFacPower))
	p20InversionCorrectiveFac.Mod(p20InversionCorrectiveFac, &bModulus)
	F.P20InversionCorrectiveFac = toUint64Slice(p20InversionCorrectiveFac, F.NbWords)

	{
		c := F.NbWords * 64
		// TODO @gbotrel check inverse performance for 32 bits
		F.UsingP20Inverse = F.NbWords > 1 && F.NbBits < c && F.Word.BitSize == 64
	}

	// rsquare
	_rSquare := big.NewInt(1)
	_rSquare.Lsh(_rSquare, uint(F.NbWords)*radix*2).Mod(_rSquare, &bModulus)
	F.RSquare = toUint64Slice(_rSquare, F.NbWords)

	var one big.Int
	one.SetUint64(1)
	one.Lsh(&one, uint(F.NbWords)*radix).Mod(&one, &bModulus)
	F.One = toUint64Slice(&one, F.NbWords)

	{
		var n big.Int
		n.SetUint64(11)
		n.Lsh(&n, uint(F.NbWords)*radix).Mod(&n, &bModulus)
		F.Eleven = toUint64Slice(&n, F.NbWords)
	}

	{
		var n big.Int
		n.SetUint64(13)
		n.Lsh(&n, uint(F.NbWords)*radix).Mod(&n, &bModulus)
		F.Thirteen = toUint64Slice(&n, F.NbWords)
	}

	// indexes (template helpers)
	F.NbWordsIndexesFull = make([]int, F.NbWords)
	F.NbWordsIndexesNoZero = make([]int, F.NbWords-1)
	for i := 0; i < F.NbWords; i++ {
		F.NbWordsIndexesFull[i] = i
		if i > 0 {
			F.NbWordsIndexesNoZero[i-1] = i
		}
	}

	// See https://hackmd.io/@gnark/modular_multiplication
	// if the last word of the modulus is smaller or equal to B,
	// we can simplify the montgomery multiplication
	const B = (^uint64(0) >> 1) - 1
	F.NoCarry = (F.Q[len(F.Q)-1] <= B) && F.NbWords <= 12
	const BSquare = ^uint64(0) >> 2
	F.NoCarrySquare = F.Q[len(F.Q)-1] <= BSquare

	// Legendre exponent (p-1)/2
	var legendreExponent big.Int
	legendreExponent.SetUint64(1)
	legendreExponent.Sub(&bModulus, &legendreExponent)
	legendreExponent.Rsh(&legendreExponent, 1)
	F.LegendreExponent = legendreExponent.Text(16)
	if F.UseAddChain {
		F.LegendreExponentData = addchain.GetAddChain(&legendreExponent)
	}

	// Sqrt pre computes
	var qMod big.Int
	qMod.SetUint64(4)
	if qMod.Mod(&bModulus, &qMod).Cmp(new(big.Int).SetUint64(3)) == 0 {
		// q ≡ 3 (mod 4)
		// using  z ≡ ± x^((p+1)/4) (mod q)
		F.SqrtQ3Mod4 = true
		var sqrtExponent big.Int
		sqrtExponent.SetUint64(1)
		sqrtExponent.Add(&bModulus, &sqrtExponent)
		sqrtExponent.Rsh(&sqrtExponent, 2)
		F.SqrtQ3Mod4Exponent = sqrtExponent.Text(16)

		// add chain stuff
		if F.UseAddChain {
			F.SqrtQ3Mod4ExponentData = addchain.GetAddChain(&sqrtExponent)
		}

	} else {
		// q ≡ 1 (mod 4)
		qMod.SetUint64(8)
		if qMod.Mod(&bModulus, &qMod).Cmp(new(big.Int).SetUint64(5)) == 0 {
			// q ≡ 5 (mod 8)
			// use Atkin's algorithm
			// see modSqrt5Mod8Prime in math/big/int.go
			F.SqrtAtkin = true
			e := new(big.Int).Rsh(&bModulus, 3) // e = (q - 5) / 8
			F.SqrtAtkinExponent = e.Text(16)
			if F.UseAddChain {
				F.SqrtAtkinExponentData = addchain.GetAddChain(e)
			}
		} else {
			// use Tonelli-Shanks
			F.SqrtTonelliShanks = true

			// Write q-1 =2ᵉ * s , s odd
			var s big.Int
			one.SetUint64(1)
			s.Sub(&bModulus, &one)

			e := s.TrailingZeroBits()
			s.Rsh(&s, e)
			F.SqrtE = uint64(e)
			F.SqrtS = toUint64Slice(&s)

			// find non residue
			var nonResidue big.Int
			nonResidue.SetInt64(2)
			one.SetUint64(1)
			for big.Jacobi(&nonResidue, &bModulus) != -1 {
				nonResidue.Add(&nonResidue, &one)
			}

			// g = nonresidue ^ s
			var g big.Int
			g.Exp(&nonResidue, &s, &bModulus)
			// store g in montgomery form
			g.Lsh(&g, uint(F.NbWords)*radix).Mod(&g, &bModulus)
			F.SqrtG = toUint64Slice(&g, F.NbWords)

			// store non residue in montgomery form
			F.NonResidue = F.ToMont(nonResidue)

			// (s+1) /2
			s.Sub(&s, &one).Rsh(&s, 1)
			F.SqrtSMinusOneOver2 = s.Text(16)

			if F.UseAddChain {
				F.SqrtSMinusOneOver2Data = addchain.GetAddChain(&s)
			}
		}
	}

	// note: to simplify output files generated, we generated ASM code only for
	// moduli that meet the condition F.NoCarry
	// asm code generation for moduli with more than 6 words can be optimized further
	F.GenerateOpsAMD64 = F.F31 || (F.NoCarry && F.NbWords <= 12 && F.NbWords > 1)
	if F.NbWords == 4 && F.GenerateOpsAMD64 && F.NbBits <= 225 {
		// 4 words field with 225 bits or less have no vector ops
		// for now since we generate both in same file we disable
		// TODO @gbotrel
		F.GenerateOpsAMD64 = false
	}
	F.GenerateVectorOpsAMD64 = F.F31 || (F.GenerateOpsAMD64 && F.NbWords == 4 && F.NbBits > 225)
	F.GenerateOpsARM64 = F.F31 || (F.GenerateOpsAMD64 && (F.NbWords%2 == 0))
	F.GenerateVectorOpsARM64 = F.F31

	// setting Mu 2^288 / q
	if F.NbWords == 4 {
		_mu := big.NewInt(1)
		_mu.Lsh(_mu, 288)
		_mu.Div(_mu, &bModulus)
		muSlice := toUint64Slice(_mu, F.NbWords)
		F.Mu = muSlice[0]
	}

	return F, nil
}

func toUint64Slice(b *big.Int, nbWords ...int) (s []uint64) {
	if len(nbWords) > 0 && nbWords[0] > len(b.Bits()) {
		s = make([]uint64, nbWords[0])
	} else {
		s = make([]uint64, len(b.Bits()))
	}

	for i, v := range b.Bits() {
		s[i] = (uint64)(v)
	}
	return
}

// https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm
// r > q, modifies rinv and qinv such that rinv.r - qinv.q = 1
func extendedEuclideanAlgo(r, q, rInv, qInv *big.Int) {
	var s1, s2, t1, t2, qi, tmpMuls, riPlusOne, tmpMult, a, b big.Int
	t1.SetUint64(1)
	rInv.Set(big.NewInt(1))
	qInv.Set(big.NewInt(0))
	a.Set(r)
	b.Set(q)

	// r_i+1 = r_i-1 - q_i.r_i
	// s_i+1 = s_i-1 - q_i.s_i
	// t_i+1 = t_i-1 - q_i.s_i
	for b.Sign() > 0 {
		qi.Div(&a, &b)
		riPlusOne.Mod(&a, &b)

		tmpMuls.Mul(&s1, &qi)
		tmpMult.Mul(&t1, &qi)

		s2.Set(&s1)
		t2.Set(&t1)

		s1.Sub(rInv, &tmpMuls)
		t1.Sub(qInv, &tmpMult)
		rInv.Set(&s2)
		qInv.Set(&t2)

		a.Set(&b)
		b.Set(&riPlusOne)
	}
	qInv.Neg(qInv)
}

// StringToMont takes an element written in string form, and returns it in Montgomery form
// Useful for hard-coding in implementation field elements from standards documents
func (f *Field) StringToMont(str string) big.Int {

	var i big.Int
	i.SetString(str, 0)
	i = f.ToMont(i)

	return i
}

func (f *Field) ToMont(nonMont big.Int) big.Int {
	var mont big.Int
	mont.Lsh(&nonMont, uint(f.NbWords)*uint(f.Word.BitSize))
	mont.Mod(&mont, f.ModulusBig)
	return mont
}

func (f *Field) FromMont(nonMont *big.Int, mont *big.Int) *Field {

	if f.NbWords == 0 {
		nonMont.SetInt64(0)
		return f
	}
	f.halve(nonMont, mont)
	for i := 1; i < f.NbWords*f.Word.BitSize; i++ {
		f.halve(nonMont, nonMont)
	}

	return f
}

func (f *Field) Exp(res *big.Int, x *big.Int, pow *big.Int) *Field {
	res.SetInt64(1)

	for i := pow.BitLen() - 1; ; {

		if pow.Bit(i) == 1 {
			res.Mul(res, x)
		}

		if i == 0 {
			break
		}
		i--

		res.Mul(res, res).Mod(res, f.ModulusBig)
	}

	res.Mod(res, f.ModulusBig)
	return f
}

func (f *Field) halve(res *big.Int, x *big.Int) {
	var z big.Int
	if x.Bit(0) == 0 {
		z.Set(x)
	} else {
		z.Add(x, f.ModulusBig)
	}
	res.Rsh(&z, 1)
}

func (f *Field) Mul(z *big.Int, x *big.Int, y *big.Int) *Field {
	z.Mul(x, y).Mod(z, f.ModulusBig)
	return f
}

func (f *Field) Add(z *big.Int, x *big.Int, y *big.Int) *Field {
	z.Add(x, y).Mod(z, f.ModulusBig)
	return f
}

func (f *Field) ToMontSlice(x []big.Int) []big.Int {
	z := make(Element, len(x))
	for i := 0; i < len(x); i++ {
		z[i] = f.ToMont(x[i])
	}
	return z
}

// TODO: Spaghetti Alert: Okay to have codegen functions here?
func CoordNameForExtensionDegree(degree uint8) string {
	switch degree {
	case 1:
		return ""
	case 2:
		return "A"
	case 6:
		return "B"
	case 12:
		return "C"
	}
	panic(fmt.Sprint("unknown extension degree", degree))
}

func (f *Field) WriteElement(element Element) string {
	var builder strings.Builder

	builder.WriteString("{")
	length := len(element)
	var subElementNames string
	if length > 1 {
		builder.WriteString("\n")
		subElementNames = CoordNameForExtensionDegree(uint8(length))
	}
	for i, e := range element {
		if length > 1 {
			builder.WriteString(subElementNames)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(": fp.Element{")
		}
		mont := f.ToMont(e)
		bavard.WriteBigIntAsUint64Slice(&builder, &mont)
		if length > 1 {
			builder.WriteString("},\n")
		}
	}
	builder.WriteString("}")
	return builder.String()
}
