package fptower

import "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"

// Expt set z to x^t in E6 and return z
func (z *E6) Expt(x *E6) *E6 {

	// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
	// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
	// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

	var result, x33 E6

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

// MulByVMinusThree set z to x*(y*v**-3) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-3 = u**-1 = (-4)**-1*u)
func (z *E6) MulByVMinusThree(x *E6, y *fp.Element) *E6 {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp E2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	z.MulByE2(x, &tmp)

	return z
}

// MulByVminusTwo set z to x*(y*v**-2) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-2 = (-4)**-1*u*v)
func (z *E6) MulByVminusTwo(x *E6, y *fp.Element) *E6 {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp E2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	var a E2
	a.MulByElement(&x.B2, y)
	z.B2.Mul(&x.B1, &tmp)
	z.B1.Mul(&x.B0, &tmp)
	z.B0.Set(&a)

	return z
}

// MulByVminusFive set z to x*(y*v**-5) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-5 = (-4)**-1*v)
func (z *E6) MulByVminusFive(x *E6, y *fp.Element) *E6 {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp E2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	var a E2
	a.Mul(&x.B2, &tmp)
	z.B2.MulByElement(&x.B1, &tmp.A1)
	z.B1.MulByElement(&x.B0, &tmp.A1)
	z.B0.Set(&a)

	return z
}
