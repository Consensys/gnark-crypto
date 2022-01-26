package element

const Base = `

// /!\ WARNING /!\
// this code has not been audited and is provided as-is. In particular,
// there is no security guarantees such as constant time implementation
// or side-channel attack resistance
// /!\ WARNING /!\

import (
	"math/big"
	"math/bits"
	"crypto/rand"
	"encoding/binary"
	"io"
	"sync"
	"strconv"
	"errors"
	"reflect"
	"strings"
)

// {{.ElementName}} represents a field element stored on {{.NbWords}} words (uint64)
// {{.ElementName}} are assumed to be in Montgomery form in all methods
// field modulus q =
//
// {{.Modulus}}
type {{.ElementName}} [{{.NbWords}}]uint64

// Limbs number of 64 bits words needed to represent {{.ElementName}}
const Limbs = {{.NbWords}}

// Bits number bits needed to represent {{.ElementName}}
const Bits = {{.NbBits}}

// Bytes number bytes needed to represent {{.ElementName}}
const Bytes = Limbs * 8

// field modulus stored as big.Int
var _modulus big.Int

// Modulus returns q as a big.Int
// q =
//
// {{.Modulus}}
func Modulus() *big.Int {
	return new(big.Int).Set(&_modulus)
}

// q (modulus)
{{- range $i := $.NbWordsIndexesFull}}
const q{{$.ElementName}}Word{{$i}} uint64 = {{index $.Q $i}} 
{{- end}}

var q{{.ElementName}} = {{.ElementName}}{
	{{- range $i := $.NbWordsIndexesFull}}
	q{{$.ElementName}}Word{{$i}},{{end}}
}

// Used for Montgomery reduction. (qInvNeg) q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
const qInvNegLsw uint64 = {{index .QInverse 0}}

// rSquare
var rSquare = {{.ElementName}}{
	{{- range $i := .RSquare}}
	{{$i}},{{end}}
}


var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

func init() {
	_modulus.SetString("{{.Modulus}}", 10)
}

// New{{.ElementName}} returns a new {{.ElementName}} from a uint64 value
//
// it is equivalent to
// 		var v New{{.ElementName}}
// 		v.SetUint64(...)
func New{{.ElementName}}(v uint64) {{.ElementName}} {
	z := {{.ElementName}}{v}
	z.Mul(&z, &rSquare)
	return z
}


// SetUint64 sets z to v and returns z
func (z *{{.ElementName}}) SetUint64(v uint64) *{{.ElementName}} {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	*z = {{.ElementName}}{v}
	return z.Mul(z, &rSquare) // z.ToMont()
}

// SetInt64 sets z to v and returns z
func (z *{{.ElementName}}) SetInt64(v int64) *{{.ElementName}} {

	// absolute value of v
	m := v >> 63
	z.SetUint64(uint64((v ^ m) - m))

	if m != 0 {
		// v is negative
		z.Neg(z)
	}

	return z
}

// Set z = x
func (z *{{.ElementName}}) Set(x *{{.ElementName}}) *{{.ElementName}} {
	{{- range $i := .NbWordsIndexesFull}}
		z[{{$i}}] = x[{{$i}}]
	{{- end}}
	return z
}

// SetInterface converts provided interface into {{.ElementName}}
// returns an error if provided type is not supported
// supported types: {{.ElementName}}, *{{.ElementName}}, uint64, int, string (interpreted as base10 integer),
// *big.Int, big.Int, []byte
func (z *{{.ElementName}}) SetInterface(i1 interface{}) (*{{.ElementName}}, error) {
	switch c1 := i1.(type) {
	case {{.ElementName}}:
		return z.Set(&c1), nil
	case *{{.ElementName}}:
		return z.Set(c1), nil
	case uint8:
		return z.SetUint64(uint64(c1)), nil
	case uint16:
		return z.SetUint64(uint64(c1)), nil
	case uint32:
		return z.SetUint64(uint64(c1)), nil
	case uint:
		return z.SetUint64(uint64(c1)), nil
	case uint64:
		return z.SetUint64(c1), nil
	case int8:
		return z.SetInt64(int64(c1)), nil
	case int16:
		return z.SetInt64(int64(c1)), nil
	case int32:
		return z.SetInt64(int64(c1)), nil
	case int64:
		return z.SetInt64(c1), nil
	case int:
		return z.SetInt64(int64(c1)), nil
	case string:
		return z.SetString(c1), nil
	case *big.Int:
		return z.SetBigInt(c1), nil
	case big.Int:
		return z.SetBigInt(&c1), nil
	case []byte:
		return z.SetBytes(c1), nil
	default:
		return nil, errors.New("can't set {{.PackageName}}.{{.ElementName}} from type " + reflect.TypeOf(i1).String())
	}
}

// SetZero z = 0
func (z *{{.ElementName}}) SetZero() *{{.ElementName}} {
	{{- range $i := .NbWordsIndexesFull}}
		z[{{$i}}] = 0
	{{- end}}
	return z
}

// SetOne z = 1 (in Montgomery form)
func (z *{{.ElementName}}) SetOne() *{{.ElementName}} {
	{{- range $i := .NbWordsIndexesFull}}
		z[{{$i}}] = {{index $.One $i}}
	{{- end}}
	return z
}


// Div z = x*y^-1 mod q
func (z *{{.ElementName}}) Div( x, y *{{.ElementName}}) *{{.ElementName}} {
	var yInv {{.ElementName}}
	yInv.Inverse( y)
	z.Mul( x, &yInv)
	return z
}

// Bit returns the i'th bit, with lsb == bit 0.
// It is the responsability of the caller to convert from Montgomery to Regular form if needed
func (z *{{.ElementName}}) Bit(i uint64) uint64 {
	j := i / 64
	if j >= {{.NbWords}} {
		return 0
	}
	return uint64(z[j] >> (i % 64) & 1)
}

// Equal returns z == x
func (z *{{.ElementName}}) Equal(x *{{.ElementName}}) bool {
	return {{- range $i :=  reverse .NbWordsIndexesNoZero}}(z[{{$i}}] == x[{{$i}}]) &&{{end}}(z[0] == x[0])
}

// IsZero returns z == 0
func (z *{{.ElementName}}) IsZero() bool {
	return ( {{- range $i :=  reverse .NbWordsIndexesNoZero}} z[{{$i}}] | {{end}}z[0]) == 0
}

// IsOne returns z == 1
func (z *{{.ElementName}}) IsOne() bool {
	return ( {{- range $i := reverse .NbWordsIndexesNoZero }} z[{{$i}}] ^ {{index $.One $i}} | {{- end}} z[0] ^ {{index $.One 0}} ) == 0
}

// IsUint64 reports whether z can be represented as an uint64.
func (z *{{.ElementName}}) IsUint64() bool {
	return ( {{- range $i :=  reverse .NbWordsIndexesNoZero}} z[{{$i}}] {{- if ne $i 1}}|{{- end}} {{end}}) == 0
}

// Cmp compares (lexicographic order) z and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
func (z *{{.ElementName}}) Cmp(x *{{.ElementName}}) int {
	_z := *z
	_x := *x
	_z.FromMont()
	_x.FromMont()
	{{- range $i := reverse $.NbWordsIndexesFull}}
	if _z[{{$i}}] > _x[{{$i}}] {
		return 1
	} else if _z[{{$i}}] < _x[{{$i}}] {
		return -1
	}
	{{- end}}
	return 0
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *{{.ElementName}}) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	// we check if the element is larger than (q-1) / 2
	// if z - (((q -1) / 2) + 1) have no underflow, then z > (q-1) / 2

	_z := *z
	_z.FromMont()

	var b uint64
	_, b = bits.Sub64(_z[0], {{index .QMinusOneHalvedP 0}}, 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		_, b = bits.Sub64(_z[{{$i}}], {{index $.QMinusOneHalvedP $i}}, b)
	{{- end}}

	return b == 0
}




// SetRandom sets z to a random element < q
func (z *{{.ElementName}}) SetRandom() (*{{.ElementName}}, error) {
	var bytes [{{mul 8 .NbWords}}]byte
	if _, err := io.ReadFull(rand.Reader, bytes[:]); err != nil {
		return nil, err 
	}
	{{- range $i :=  .NbWordsIndexesFull}}
		{{- $k := add $i 1}}
		z[{{$i}}] = binary.BigEndian.Uint64(bytes[{{mul $i 8}}:{{mul $k 8}}])
	{{- end}}
	z[{{$.NbWordsLastIndex}}] %= {{index $.Q $.NbWordsLastIndex}}

	{{ template "reduce" . }}

	return z, nil
}

// One returns 1 (in montgommery form)
func One() {{.ElementName}} {
	var one {{.ElementName}}
	one.SetOne()
	return one
}

// Halve sets z to z / 2 (mod p)
func (z *{{.ElementName}}) Halve()  {
	{{- if .NoCarry}}
		if z[0]&1 == 1 {
			var carry uint64
			{{ template "add_q" dict "all" . "V1" "z" }}
		}
		{{ rsh "z" .NbWords}}
	{{ else}}
		var twoInv {{.ElementName}}
		twoInv.SetOne().Double(&twoInv).Inverse(&twoInv)
		z.Mul(z, &twoInv)
	{{end}}
}


// API with assembly impl

// Mul z = x * y mod q
// see https://hackmd.io/@gnark/modular_multiplication
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	mul(z, x, y)
	return z
}

// Square z = x * x mod q
// see https://hackmd.io/@gnark/modular_multiplication
func (z *{{.ElementName}}) Square(x *{{.ElementName}}) *{{.ElementName}} {
	mul(z,x, x)
	return z
}

// FromMont converts z in place (i.e. mutates) from Montgomery to regular representation
// sets and returns z = z * 1
func (z *{{.ElementName}}) FromMont() *{{.ElementName}} {
	fromMont(z)
	return z
}

// Add z = x + y mod q
func (z *{{.ElementName}}) Add( x, y *{{.ElementName}}) *{{.ElementName}} {
	add(z, x, y)
	return z
}

// Double z = x + x mod q, aka Lsh 1
func (z *{{.ElementName}}) Double( x *{{.ElementName}}) *{{.ElementName}} {
	double(z, x)
	return z
}


// Sub  z = x - y mod q
func (z *{{.ElementName}}) Sub( x, y *{{.ElementName}}) *{{.ElementName}} {
	sub(z, x, y)
	return z
}

// Neg z = q - x
func (z *{{.ElementName}}) Neg( x *{{.ElementName}}) *{{.ElementName}} {
	neg(z, x)
	return z
}




// Generic (no ADX instructions, no AMD64) versions of multiplication and squaring algorithms

func _mulGeneric(z,x,y *{{.ElementName}}) {
	{{ if .NoCarry}}
		{{ template "mul_nocarry" dict "all" . "V1" "x" "V2" "y"}}
	{{ else }}
		{{ template "mul_cios" dict "all" . "V1" "x" "V2" "y" "NoReturn" true}}
	{{ end }}
	{{ template "reduce" . }}
}

func _mulWGeneric(z,x *{{.ElementName}}, y uint64) {
	{{ template "mul_nocarry_v2" dict "all" . "V2" "x"}}
	{{ template "reduce" . }}
}


func _fromMontGeneric(z *{{.ElementName}}) {
	// the following lines implement z = z * 1
	// with a modified CIOS montgomery multiplication
	{{- range $j := .NbWordsIndexesFull}}
	{
		// m = z[0]n'[0] mod W
		m := z[0] * {{index $.QInverse 0}}
		C := madd0(m, {{index $.Q 0}}, z[0])
		{{- range $i := $.NbWordsIndexesNoZero}}
			C, z[{{sub $i 1}}] = madd2(m, {{index $.Q $i}}, z[{{$i}}], C)
		{{- end}}
		z[{{sub $.NbWords 1}}] = C
	}
	{{- end}}

	{{ template "reduce" .}}
}



func _addGeneric(z,  x, y *{{.ElementName}}) {
	var carry uint64
	{{$k := sub $.NbWords 1}}
	z[0], carry = bits.Add64(x[0], y[0], 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		{{- if eq $i $.NbWordsLastIndex}}
		{{- else}}
			z[{{$i}}], carry = bits.Add64(x[{{$i}}], y[{{$i}}], carry)
		{{- end}}
	{{- end}}
	{{- if .NoCarry}}
		z[{{$k}}], _ = bits.Add64(x[{{$k}}], y[{{$k}}], carry)
	{{- else }}
		z[{{$k}}], carry = bits.Add64(x[{{$k}}], y[{{$k}}], carry)
		// if we overflowed the last addition, z >= q
		// if z >= q, z = z - q
		if carry != 0 {
			// we overflowed, so z >= q
			z[0], carry = bits.Sub64(z[0], {{index $.Q 0}}, 0)
			{{- range $i := .NbWordsIndexesNoZero}}
				z[{{$i}}], carry = bits.Sub64(z[{{$i}}], {{index $.Q $i}}, carry)
			{{- end}}
			return
		}
	{{- end}}

	{{ template "reduce" .}}
}

func _doubleGeneric(z,  x *{{.ElementName}}) {
	var carry uint64
	{{$k := sub $.NbWords 1}}
	z[0], carry = bits.Add64(x[0], x[0], 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		{{- if eq $i $.NbWordsLastIndex}}
		{{- else}}
			z[{{$i}}], carry = bits.Add64(x[{{$i}}], x[{{$i}}], carry)
		{{- end}}
	{{- end}}
	{{- if .NoCarry}}
		z[{{$k}}], _ = bits.Add64(x[{{$k}}], x[{{$k}}], carry)
	{{- else }}
		z[{{$k}}], carry = bits.Add64(x[{{$k}}], x[{{$k}}], carry)
		// if we overflowed the last addition, z >= q
		// if z >= q, z = z - q
		if carry != 0 {
			// we overflowed, so z >= q
			z[0], carry = bits.Sub64(z[0], {{index $.Q 0}}, 0)
			{{- range $i := .NbWordsIndexesNoZero}}
				z[{{$i}}], carry = bits.Sub64(z[{{$i}}], {{index $.Q $i}}, carry)
			{{- end}}
			return
		}
	{{- end}}

	{{ template "reduce" .}}
}


func _subGeneric(z,  x, y *{{.ElementName}}) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		z[{{$i}}], b = bits.Sub64(x[{{$i}}], y[{{$i}}], b)
	{{- end}}
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], {{index $.Q 0}}, 0)
		{{- range $i := .NbWordsIndexesNoZero}}
			{{- if eq $i $.NbWordsLastIndex}}
				z[{{$i}}], _ = bits.Add64(z[{{$i}}], {{index $.Q $i}}, c)
			{{- else}}
				z[{{$i}}], c = bits.Add64(z[{{$i}}], {{index $.Q $i}}, c)
			{{- end}}
		{{- end}}
	}
}

func _negGeneric(z,  x *{{.ElementName}}) {
	if x.IsZero() {
		z.SetZero()
		return
	}
	var borrow uint64
	z[0], borrow = bits.Sub64({{index $.Q 0}}, x[0], 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		{{- if eq $i $.NbWordsLastIndex}}
			z[{{$i}}], _ = bits.Sub64({{index $.Q $i}}, x[{{$i}}], borrow)
		{{- else}}
			z[{{$i}}], borrow = bits.Sub64({{index $.Q $i}}, x[{{$i}}], borrow)
		{{- end}}
	{{- end}}
}


func _reduceGeneric(z *{{.ElementName}})  {
	{{ template "reduce" . }}
}

func mulByConstant(z *{{.ElementName}}, c uint8) {
	switch c {
	case 0:
		z.SetZero()
		return
	case 1:
		return
	case 2:
		z.Double(z)
		return
	case 3:
		_z := *z
		z.Double(z).Add(z, &_z)
	case 5:
		_z := *z
		z.Double(z).Double(z).Add(z, &_z)
	case 11:
		_z := *z
		z.Double(z).Double(z).Add(z, &_z).Double(z).Add(z, &_z)
	default:
		var y {{.ElementName}}
		y.SetUint64(uint64(c))
		z.Mul(z, &y)
	}
}


// BatchInvert returns a new slice with every element inverted.
// Uses Montgomery batch inversion trick
func BatchInvert(a []{{.ElementName}}) []{{.ElementName}} {
	res := make([]{{.ElementName}}, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	accumulator := One()

	for i:=0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i] = accumulator
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}

func _butterflyGeneric(a, b *{{.ElementName}}) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

// BitLen returns the minimum number of bits needed to represent z
// returns 0 if z == 0
func (z *{{.ElementName}}) BitLen() int {
	{{- range $i := reverse .NbWordsIndexesNoZero}}
	if z[{{$i}}] != 0 {
		return {{mul $i 64}} + bits.Len64(z[{{$i}}])
	}
	{{- end}}
	return bits.Len64(z[0])
}

{{ define "add_q" }}
	// {{$.V1}} = {{$.V1}} + q 
	{{$.V1}}[0], carry = bits.Add64({{$.V1}}[0], {{index $.all.Q 0}}, 0)
	{{- range $i := .all.NbWordsIndexesNoZero}}
		{{- if eq $i $.all.NbWordsLastIndex}}
			{{$.V1}}[{{$i}}], _ = bits.Add64({{$.V1}}[{{$i}}], {{index $.all.Q $i}}, carry)
		{{- else}}
			{{$.V1}}[{{$i}}], carry = bits.Add64({{$.V1}}[{{$i}}], {{index $.all.Q $i}}, carry)
		{{- end}}
	{{- end}}
{{ end }}

{{ define "rsh V nbWords" }}
	// {{$.V}} = {{$.V}} >> 1
	{{$lastIndex := sub .nbWords 1}}
	{{- range $i :=  iterate .nbWords}}
		{{- if ne $i $lastIndex}}
			{{$.V}}[{{$i}}] = {{$.V}}[{{$i}}] >> 1 | {{$.V}}[{{(add $i 1)}}] << 63
		{{- end}}
	{{- end}}
	{{$.V}}[{{$lastIndex}}] >>= 1
{{ end }}


`
