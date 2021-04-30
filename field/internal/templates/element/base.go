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
var q{{.ElementName}} = {{.ElementName}}{
	{{- range $i := .NbWordsIndexesFull}}
	{{index $.Q $i}},{{end}}
}

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


// SetUint64 z = v, sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
func (z *{{.ElementName}}) SetUint64(v uint64) *{{.ElementName}} {
	*z = {{.ElementName}}{v}
	return z.Mul(z, &rSquare) // z.ToMont()
}

// Set z = x
func (z *{{.ElementName}}) Set(x *{{.ElementName}}) *{{.ElementName}} {
	{{- range $i := .NbWordsIndexesFull}}
		z[{{$i}}] = x[{{$i}}]
	{{- end}}
	return z
}

// SetInterface converts i1 from uint64, int, string, or {{.ElementName}}, big.Int into {{.ElementName}}
// panic if provided type is not supported
func (z *{{.ElementName}}) SetInterface(i1 interface{}) *{{.ElementName}} {
	switch c1 := i1.(type) {
	case {{.ElementName}}:
		return z.Set(&c1)
	case *{{.ElementName}}:
		return z.Set(c1)
	case uint64:
		return z.SetUint64(c1)
	case int:
		return z.SetString(strconv.Itoa(c1))
	case string:
		return z.SetString(c1)
	case *big.Int:
		return z.SetBigInt(c1)
	case big.Int:
		return z.SetBigInt(&c1)
	case []byte:
		return z.SetBytes(c1)
	default:
		panic("invalid type")
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

// Equal returns z == x
func (z *{{.ElementName}}) Equal(x *{{.ElementName}}) bool {
	return {{- range $i :=  reverse .NbWordsIndexesNoZero}}(z[{{$i}}] == x[{{$i}}]) &&{{end}}(z[0] == x[0])
}

// IsZero returns z == 0
func (z *{{.ElementName}}) IsZero() bool {
	return ( {{- range $i :=  reverse .NbWordsIndexesNoZero}} z[{{$i}}] | {{end}}z[0]) == 0
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


// API with assembly impl

// Mul z = x * y mod q 
// see https://hackmd.io/@zkteam/modular_multiplication
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	mul(z, x, y)
	return z
}

// Square z = x * x mod q
// see https://hackmd.io/@zkteam/modular_multiplication
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
	default:
		var y {{.ElementName}}
		y.SetUint64(uint64(c))
		z.Mul(z, &y)
	}
}

`
