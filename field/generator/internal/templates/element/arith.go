package element

const Arith = `
import (
	"math/bits"
)

// madd0 hi = a*b + c (discards lo bits)
func madd0(a, b, c {{$.Word.TypeLower}}) (hi {{$.Word.TypeLower}}) {
	var carry, lo {{$.Word.TypeLower}}
	hi, lo = bits.{{$.Word.Mul}}(a, b)
	_, carry = bits.{{$.Word.Add}}(lo, c, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, 0, carry)
	return
}

// madd1 hi, lo = a*b + c
func madd1(a, b, c {{$.Word.TypeLower}}) (hi {{$.Word.TypeLower}}, lo {{$.Word.TypeLower}}) {
	var carry {{$.Word.TypeLower}}
	hi, lo = bits.{{$.Word.Mul}}(a, b)
	lo, carry = bits.{{$.Word.Add}}(lo, c, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, 0, carry)
	return
}

// madd2 hi, lo = a*b + c + d
func madd2(a, b, c, d {{$.Word.TypeLower}}) (hi {{$.Word.TypeLower}}, lo {{$.Word.TypeLower}}) {
	var carry {{$.Word.TypeLower}}
	hi, lo = bits.{{$.Word.Mul}}(a, b)
	c, carry = bits.{{$.Word.Add}}(c, d, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, 0, carry)
	lo, carry = bits.{{$.Word.Add}}(lo, c, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, 0, carry)
	return
}


func madd3(a, b, c, d, e {{$.Word.TypeLower}}) (hi {{$.Word.TypeLower}}, lo {{$.Word.TypeLower}}) {
	var carry {{$.Word.TypeLower}}
	hi, lo = bits.{{$.Word.Mul}}(a, b)
	c, carry = bits.{{$.Word.Add}}(c, d, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, 0, carry)
	lo, carry = bits.{{$.Word.Add}}(lo, c, 0)
	hi, _ = bits.{{$.Word.Add}}(hi, e, carry)
	return
}


`
