package pairing

const Pairing = `
// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func (curve *Curve) FinalExponentiation(z *PairingResult, _z ...*PairingResult) PairingResult {
	var result PairingResult
	result.Set(z)

	// if additional parameters are provided, multiply them into z
	for _, e := range _z {
		result.Mul(&result, e)
	}

	result.FinalExponentiation(&result)

	return result
}

// FinalExponentiation computes the final expo x**((p**12 - 1)/r)
func (z *PairingResult) FinalExponentiation(x *PairingResult) *PairingResult {

{{- /* TODO add a curve family parameter for BLS12, BN and use it here */}}
{{- if and (eq .Fp6NonResidue "9,1") (eq .FpModulus "21888242871839275222246405745257275088696311157297823662689037894645226208583") }}
	// For BN curves use Section 5 of https://eprint.iacr.org/2008/490.pdf; their x is our t

	// TODO modify sage test points script to include a factor of 3 in the final exponent for BLS curves but not BN curves
	var mt [4]e12 // mt[i] is m^(t^i)

	// set m[0] = x^((p^6-1)*(p^2+1))
	{
		mt[0].Set(x)
		var temp e12
		temp.FrobeniusCube(&mt[0]).
			FrobeniusCube(&temp)

		mt[0].Inverse(&mt[0])
		temp.Mul(&temp, &mt[0])

		mt[0].FrobeniusSquare(&temp).
			Mul(&mt[0], &temp)
	}

	// "hard part": set z = m[0]^((p^4-p^2+1)/r)

	mt[1].Expt(&mt[0])
	mt[2].Expt(&mt[1])
	mt[3].Expt(&mt[2])

	// prepare y
	var y [7]e12

	y[1].InverseUnitary(&mt[0])
	y[4].Set(&mt[1])
	y[5].InverseUnitary(&mt[2])
	y[6].Set(&mt[3])

	mt[0].Frobenius(&mt[0])
	mt[1].Frobenius(&mt[1])
	mt[2].Frobenius(&mt[2])
	mt[3].Frobenius(&mt[3])

	y[0].Set(&mt[0])
	y[3].InverseUnitary(&mt[1])
	y[4].Mul(&y[4], &mt[2]).InverseUnitary(&y[4])
	y[6].Mul(&y[6], &mt[3]).InverseUnitary(&y[6])

	mt[0].Frobenius(&mt[0])
	mt[2].Frobenius(&mt[2])

	y[0].Mul(&y[0], &mt[0])
	y[2].Set(&mt[2])

	mt[0].Frobenius(&mt[0])

	y[0].Mul(&y[0], &mt[0])

	// compute addition chain
	var t [2]e12

	t[0].Square(&y[6])
	t[0].Mul(&t[0], &y[4])
	t[0].Mul(&t[0], &y[5])
	t[1].Mul(&y[3], &y[5])
	t[1].Mul(&t[1], &t[0])
	t[0].Mul(&t[0], &y[2])
	t[1].Square(&t[1])
	t[1].Mul(&t[1], &t[0])
	t[1].Square(&t[1])
	t[0].Mul(&t[1], &y[1])
	t[1].Mul(&t[1], &y[0])
	t[0].Square(&t[0])
	z.Mul(&t[0], &t[1])

	return z

{{- else }}
	// For BLS curves use Section 3 of https://eprint.iacr.org/2016/130.pdf; "hard part" is Algorithm 1 of https://eprint.iacr.org/2016/130.pdf
	var result PairingResult
	result.Set(x)

	// memalloc
	var t [6]PairingResult

	// buf = x**(p^6-1)
	t[0].FrobeniusCube(&result).
		FrobeniusCube(&t[0])

	result.Inverse(&result)
	t[0].Mul(&t[0], &result)

	// x = (x**(p^6-1)) ^(p^2+1)
	result.FrobeniusSquare(&t[0]).
		Mul(&result, &t[0])

	// hard part (up to permutation)
	// performs the hard part of the final expo
	// Algorithm 1 of https://eprint.iacr.org/2016/130.pdf
	// The result is the same as p**4-p**2+1/r, but up to permutation (it's 3* (p**4 -p**2 +1 /r)), ok since r=1 mod 3)

	t[0].InverseUnitary(&result).Square(&t[0])
	t[5].Expt(&result)
	t[1].Square(&t[5])
	t[3].Mul(&t[0], &t[5])

	t[0].Expt(&t[3])
	t[2].Expt(&t[0])
	t[4].Expt(&t[2])

	t[4].Mul(&t[1], &t[4])
	t[1].Expt(&t[4])
	t[3].InverseUnitary(&t[3])
	t[1].Mul(&t[3], &t[1])
	t[1].Mul(&t[1], &result)

	t[0].Mul(&t[0], &result)
	t[0].FrobeniusCube(&t[0])

	t[3].InverseUnitary(&result)
	t[4].Mul(&t[3], &t[4])
	t[4].Frobenius(&t[4])

	t[5].Mul(&t[2], &t[5])
	t[5].FrobeniusSquare(&t[5])

	t[5].Mul(&t[5], &t[0])
	t[5].Mul(&t[5], &t[4])
	t[5].Mul(&t[5], &t[1])

	result.Set(&t[5])

	z.Set(&result)
	return z

{{- end }}
}

// MillerLoop Miller loop
func (curve *Curve) MillerLoop(P G1Affine, Q G2Affine, result *PairingResult) *PairingResult {

	// init result
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return result
	}

	// the line goes through QCur and QNext
	var QCur, QNext, QNextNeg G2Jac
	var QNeg G2Affine

	// Stores -Q
	QNeg.Neg(&Q)

	// init QCur with Q
	Q.ToJacobian(&QCur)

	var lEval lineEvalRes

	// Miller loop
	for i := len(curve.loopCounter) - 2; i >= 0; i-- {
		QNext.Set(&QCur)
		QNext.Double()
		QNextNeg.Neg(&QNext)

		result.Square(result)

		// evaluates line though Qcur,2Qcur at P
		lineEvalJac(QCur, QNextNeg, &P, &lEval)
		lEval.mulAssign(result)

		if curve.loopCounter[i] == 1 {
			// evaluates line through 2Qcur, Q at P
			lineEvalAffine(QNext, Q, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&Q)

		} else if curve.loopCounter[i] == -1 {
			// evaluates line through 2Qcur, -Q at P
			lineEvalAffine(QNext, QNeg, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&QNeg)
		}
		QCur.Set(&QNext)
	}

	{{template "ExtraWork" dict "all" . }}

	return result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
// The case in which Q=R=Infinity is not handled as this doesn't happen in the SNARK pairing
func lineEvalJac(Q, R G2Jac, P *G1Affine, result *lineEvalRes) {
	// converts Q and R to projective coords
	Q.ToProjFromJac()
	R.ToProjFromJac()

	// line eq: w^3*(QyRz-QzRy)x +  w^2*(QzRx - QxRz)y + w^5*(QxRy-QyRxz)
	// result.r1 = QyRz-QzRy
	// result.r0 = QzRx - QxRz
	// result.r2 = QxRy-QyRxz

	result.r1.Mul(&Q.Y, &R.Z)
	result.r0.Mul(&Q.Z, &R.X)
	result.r2.Mul(&Q.X, &R.Y)

	Q.Z.Mul(&Q.Z, &R.Y)
	Q.X.Mul(&Q.X, &R.Z)
	Q.Y.Mul(&Q.Y, &R.X)

	result.r1.Sub(&result.r1, &Q.Z)
	result.r0.Sub(&result.r0, &Q.X)
	result.r2.Sub(&result.r2, &Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
	//result.r2.MulByElement(&result.r2, &P.Z)
}

// Same as above but R is in affine coords
func lineEvalAffine(Q G2Jac, R G2Affine, P *G1Affine, result *lineEvalRes) {

	// converts Q and R to projective coords
	Q.ToProjFromJac()

	// line eq: w^3*(QyRz-QzRy)x +  w^2*(QzRx - QxRz)y + w^5*(QxRy-QyRxz)
	// result.r1 = QyRz-QzRy
	// result.r0 = QzRx - QxRz
	// result.r2 = QxRy-QyRxz

	result.r1.Set(&Q.Y)
	result.r0.Mul(&Q.Z, &R.X)
	result.r2.Mul(&Q.X, &R.Y)

	Q.Z.Mul(&Q.Z, &R.Y)
	Q.Y.Mul(&Q.Y, &R.X)

	result.r1.Sub(&result.r1, &Q.Z)
	result.r0.Sub(&result.r0, &Q.X)
	result.r2.Sub(&result.r2, &Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
	// result.r2.MulByElement(&result.r2, &P.Z)
}

type lineEvalRes struct {
	r0 {{.Fp2Name}} // c0.b1
	r1 {{.Fp2Name}} // c1.b1
	r2 {{.Fp2Name}} // c1.b2
}

func (l *lineEvalRes) mulAssign(z *PairingResult) *PairingResult {

	{{template "MulAssign" dict "all" . }}

	return z
}

`

// ExtraWork extra operations needed when the loop shortening is used (cf Vecauteren, Optimal Pairing)
const ExtraWork = `
{{define "ExtraWork" }}
	{{if eq $.all.Fpackage "bn256" }}
		// cf https://eprint.iacr.org/2010/354.pdf for instance for optimal Ate Pairing
		var Q1, Q2 G2Affine

		//Q1 = Frob(Q)
		Q1.X.Conjugate(&Q.X).MulByNonResiduePower2(&Q1.X)
		Q1.Y.Conjugate(&Q.Y).MulByNonResiduePower3(&Q1.Y)

		// Q2 = -Frob2(Q)
		Q2.X.MulByNonResiduePowerSquare2(&Q.X)
		Q2.Y.MulByNonResiduePowerSquare3(&Q.Y).Neg(&Q2.Y)

		lineEvalAffine(QCur, Q1, &P, &lEval)
		lEval.mulAssign(result)

		QCur.AddMixed(&Q1)

		lineEvalAffine(QCur, Q2, &P, &lEval)
		lEval.mulAssign(result)
	{{end}}
{{- end}}
`

// MulAssign multiplies the result of a line evalution to a e12 elmt.
// The line evaluation result is sparse therefore there is a special optimized method to handle this case.
const MulAssign = `
{{define "MulAssign" }}
	{{if eq $.all.Fpackage "bn256" }}	
		var a, b, c e12
		a.MulByVW(z, &l.r1)
		b.MulByV(z, &l.r0)
		c.MulByV2W(z, &l.r2)
		z.Add(&a, &b).Add(z, &c)
	{{else if eq $.all.Fpackage "bls377" }}
		var a, b, c e12
		a.MulByVW(z, &l.r1)
		b.MulByV(z, &l.r0)
		c.MulByV2W(z, &l.r2)
		z.Add(&a, &b).Add(z, &c)
	{{else if eq $.all.Fpackage "bls381" }}
		var a, b, c e12
		a.MulByVWNRInv(z, &l.r1)
		b.MulByV2NRInv(z, &l.r0)
		c.MulByWNRInv(z, &l.r2)
		z.Add(&a, &b).Add(z, &c)
	{{end}}
{{end}}
`
