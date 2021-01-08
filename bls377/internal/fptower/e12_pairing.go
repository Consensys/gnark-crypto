package fptower

// Expt set z to x^t in E12 and return z
func (z *E12) Expt(x *E12) *E12 {
	// const tAbsVal uint64 = 9586122913090633729
	// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
	// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
	// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

	var result, x33 E12

	// a shortest addition chain for 136227
	result.Set(x)                    // 0                1
	result.CyclotomicSquare(&result) // 1( 0)            2
	result.CyclotomicSquare(&result) // 2( 1)            4
	result.CyclotomicSquare(&result) // 3( 2)            8
	result.CyclotomicSquare(&result) // 4( 3)           16
	result.CyclotomicSquare(&result) // 5( 4)           32
	result.Mul(&result, x)           // 6( 5, 0)        33
	x33.Set(&result)                 // save x33 for step 14
	result.CyclotomicSquare(&result) // 7( 6)           66
	result.CyclotomicSquare(&result) // 8( 7)          132
	result.CyclotomicSquare(&result) // 9( 8)          264
	result.CyclotomicSquare(&result) // 10( 9)          528
	result.CyclotomicSquare(&result) // 11(10)         1056
	result.CyclotomicSquare(&result) // 12(11)         2112
	result.CyclotomicSquare(&result) // 13(12)         4224
	result.Mul(&result, &x33)        // 14(13, 6)      4257
	result.CyclotomicSquare(&result) // 15(14)         8514
	result.CyclotomicSquare(&result) // 16(15)        17028
	result.CyclotomicSquare(&result) // 17(16)        34056
	result.CyclotomicSquare(&result) // 18(17)        68112
	result.Mul(&result, x)           // 19(18, 0)     68113
	result.CyclotomicSquare(&result) // 20(19)       136226
	result.Mul(&result, x)           // 21(20, 0)    136227

	// the remaining 46 bits
	for i := 0; i < 46; i++ {
		result.CyclotomicSquare(&result)
	}
	result.Mul(&result, x)

	z.Set(&result)
	return z
}

// MulByVW set z to x*(y*v*w) and return z
// here y*v*w means the E12 element with C1.B1=y and all other components 0
func (z *E12) MulByVW(x *E12, y *E2) *E12 {

	var result E12
	var yNR E2

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C1.B1, &yNR)
	result.C0.B1.Mul(&x.C1.B2, &yNR)
	result.C0.B2.Mul(&x.C1.B0, y)
	result.C1.B0.Mul(&x.C0.B2, &yNR)
	result.C1.B1.Mul(&x.C0.B0, y)
	result.C1.B2.Mul(&x.C0.B1, y)
	z.Set(&result)
	return z
}

// MulByV set z to x*(y*v) and return z
// here y*v means the E12 element with C0.B1=y and all other components 0
func (z *E12) MulByV(x *E12, y *E2) *E12 {

	var result E12
	var yNR E2

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C0.B2, &yNR)
	result.C0.B1.Mul(&x.C0.B0, y)
	result.C0.B2.Mul(&x.C0.B1, y)
	result.C1.B0.Mul(&x.C1.B2, &yNR)
	result.C1.B1.Mul(&x.C1.B0, y)
	result.C1.B2.Mul(&x.C1.B1, y)
	z.Set(&result)
	return z
}

// MulByV2W set z to x*(y*v^2*w) and return z
// here y*v^2*w means the E12 element with C1.B2=y and all other components 0
func (z *E12) MulByV2W(x *E12, y *E2) *E12 {

	var result E12
	var yNR E2

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C1.B0, &yNR)
	result.C0.B1.Mul(&x.C1.B1, &yNR)
	result.C0.B2.Mul(&x.C1.B2, &yNR)
	result.C1.B0.Mul(&x.C0.B1, &yNR)
	result.C1.B1.Mul(&x.C0.B2, &yNR)
	result.C1.B2.Mul(&x.C0.B0, y)
	z.Set(&result)
	return z
}

// MulBy034 multiplication by sparse element
func (z *E12) MulBy034(c0, c3, c4 *E2) *E12 {

	var z0, z1, z2, z3, z4, z5, tmp1, tmp2 E2
	var t [12]E2

	z0 = z.C0.B0
	z1 = z.C0.B1
	z2 = z.C0.B2
	z3 = z.C1.B0
	z4 = z.C1.B1
	z5 = z.C1.B2

	tmp1.MulByNonResidue(c3)
	tmp2.MulByNonResidue(c4)

	t[0].Mul(&tmp1, &z5)
	t[1].Mul(&tmp2, &z4)
	t[2].Mul(c3, &z3)
	t[3].Mul(&tmp2, &z5)
	t[4].Mul(c3, &z4)
	t[5].Mul(c4, &z3)
	t[6].Mul(c3, &z0)
	t[7].Mul(&tmp2, &z2)
	t[8].Mul(c3, &z1)
	t[9].Mul(c4, &z0)
	t[10].Mul(c3, &z2)
	t[11].Mul(c4, &z1)

	z.C0.B0.Mul(c0, &z0).
		Add(&z.C0.B0, &t[0]).
		Add(&z.C0.B0, &t[1])
	z.C0.B1.Mul(c0, &z1).
		Add(&z.C0.B1, &t[2]).
		Add(&z.C0.B1, &t[3])
	z.C0.B2.Mul(c0, &z2).
		Add(&z.C0.B2, &t[4]).
		Add(&z.C0.B2, &t[5])
	z.C1.B0.Mul(c0, &z3).
		Add(&z.C1.B0, &t[6]).
		Add(&z.C1.B0, &t[7])
	z.C1.B1.Mul(c0, &z4).
		Add(&z.C1.B1, &t[8]).
		Add(&z.C1.B1, &t[9])
	z.C1.B2.Mul(c0, &z5).
		Add(&z.C1.B2, &t[10]).
		Add(&z.C1.B2, &t[11])

	return z
}
