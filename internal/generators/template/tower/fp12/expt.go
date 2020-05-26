package fp12

const Expt = `
const tAbsVal uint64 = {{.T}} {{ if .TNeg }}// negative{{- end }}

// Expt set z to x^t in {{.Fp12Name}} and return z
// TODO make a ExptAssign method that assigns the result to self; then this method can assert fail if z != x
// TODO Expt is the only method that depends on tAbsVal.  The rest of the tower does not depend on this value.  Logically, Expt should be separated from the rest of the tower.
func (z *{{.Fp12Name}}) Expt(x *{{.Fp12Name}}) *{{.Fp12Name}} {
	// TODO what if x==0?
	// TODO make this match Element.Exp: x is a non-pointer?
	{{- if (eq .T "9586122913090633729" ) }}

		// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
		// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
		// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

		var result, x33 {{.Fp12Name}}

		// a shortest addition chain for 136227
		result.Set(x)             // 0                1
		result.Square(&result)    // 1( 0)            2
		result.Square(&result)    // 2( 1)            4
		result.Square(&result)    // 3( 2)            8
		result.Square(&result)    // 4( 3)           16
		result.Square(&result)    // 5( 4)           32
		result.Mul(&result, x)    // 6( 5, 0)        33
		x33.Set(&result)          // save x33 for step 14
		result.Square(&result)    // 7( 6)           66
		result.Square(&result)    // 8( 7)          132
		result.Square(&result)    // 9( 8)          264
		result.Square(&result)    // 10( 9)          528
		result.Square(&result)    // 11(10)         1056
		result.Square(&result)    // 12(11)         2112
		result.Square(&result)    // 13(12)         4224
		result.Mul(&result, &x33) // 14(13, 6)      4257
		result.Square(&result)    // 15(14)         8514
		result.Square(&result)    // 16(15)        17028
		result.Square(&result)    // 17(16)        34056
		result.Square(&result)    // 18(17)        68112
		result.Mul(&result, x)    // 19(18, 0)     68113
		result.Square(&result)    // 20(19)       136226
		result.Mul(&result, x)    // 21(20, 0)    136227
	
		// the remaining 46 bits
		for i := 0; i < 46; i++ {
			result.Square(&result)
		}
		result.Mul(&result, x)
	
	{{- else }}
		var result {{.Fp12Name}}
		result.Set(x)

		l := bits.Len64(tAbsVal) - 2
		for i := l; i >= 0; i-- {
			result.Square(&result)
			if tAbsVal&(1<<uint(i)) != 0 {
				result.Mul(&result, x)
			}
		}
	{{- end }}

	{{- if .TNeg }}
		result.Conjugate(&result) // because tAbsVal is negative
	{{- end }}

	z.Set(&result)
	return z
}
`
