package element

// Fuzz when build tag is provided, we expose Generic methods to be used by ecc/ package fuzzing functions
const Fuzz = `

import (
	"encoding/binary"
	"io"
	"math/bits"
	"math/big"
	"bytes"
)

const (
	fuzzInteresting = 1
	fuzzNormal      = 0
	fuzzDiscard     = -1
)

// Fuzz arithmetic operations fuzzer
func Fuzz(data []byte) int {
	r := bytes.NewReader(data)

	var e1, e2 {{.ElementName}}
	e1.SetRawBytes(r)
	e2.SetRawBytes(r)

	{
		// mul assembly 

		var c, _c {{.ElementName}}
		a, _a, b, _b := e1, e1, e2, e2
		c.Mul(&a, &b)
		_mulGeneric(&_c, &_a, &_b)

		if !c.Equal(&_c) {
			panic("mul asm != mul generic on {{.ElementName}}")
		}
	}

	{
		// inverse
		inv := e1 
		inv.Inverse(&inv)

		var bInv, b1, b2 big.Int 
		e1.ToBigIntRegular(&b1)
		bInv.ModInverse(&b1, Modulus())
		inv.ToBigIntRegular(&b2)

		if b2.Cmp(&bInv) != 0 {
			panic("inverse operation doesn't match big int result")
		}
	}

	{
		// a + -a == 0
		a, b := e1, e1
		b.Neg(&b)
		a.Add(&a, &b)
		if !a.IsZero() {
			panic("a + -a != 0")
		}
	}

	return fuzzNormal

}

// SetRawBytes reads up to Bytes (bytes needed to represent {{.ElementName}}) from reader
// and interpret it as big endian uint64
// used for fuzzing purposes only
func (z *{{.ElementName}}) SetRawBytes(r io.Reader) {

	buf := make([]byte, 8)
	
	for i := 0; i < len(z); i++ {
		if _, err := io.ReadFull(r, buf); err != nil {
			goto eof
		}
		z[i] = binary.BigEndian.Uint64(buf[:])
	}
eof:
	z[{{.NbWordsLastIndex}}] %= q{{.ElementName}}[{{.NbWordsLastIndex}}]

	if z.BiggerModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q{{$.ElementName}}[0], 0)
		{{- range $i := .NbWordsIndexesNoZero}}
			z[{{$i}}], b = bits.Sub64(z[{{$i}}], q{{$.ElementName}}[{{$i}}], b)
		{{- end}}
	}

	return 
}


func (z *{{.ElementName}}) BiggerModulus() bool {
	{{- range $i :=  reverse .NbWordsIndexesNoZero}}
	if z[{{$i}}] > q{{$.ElementName}}[{{$i}}] {
		return true
	}
	if z[{{$i}}] < q{{$.ElementName}}[{{$i}}] {
		return false
	}
	{{end}}
	
	return z[0] >= q{{.ElementName}}[0]
}

`
