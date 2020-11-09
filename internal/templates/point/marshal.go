package point

// Marshal ...
const Marshal = `
import (
	"io"
	"math/big"
	"runtime"
	"encoding/binary"

	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fp"
	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fr"
	"github.com/consensys/gurvy/utils/debug"
)


// To encode G1 and G2 points, we mask the most significant bits with these bits to specify without ambiguity
// metadata needed for point (de)compression
{{- if gt .UnusedBits 3}}
// we follow the BLS381 style encoding as specified in ZCash and now IETF
// The most significant bit, when set, indicates that the point is in compressed form. Otherwise, the point is in uncompressed form.
// The second-most significant bit indicates that the point is at infinity. If this bit is set, the remaining bits of the group element's encoding should be set to zero.
// The third-most significant bit is set if (and only if) this point is in compressed form and it is not the point at infinity and its y-coordinate is the lexicographically largest of the two associated with the encoded x-coordinate.
const (
	mMask                 uint64 = 0b111 << 61
	mUncompressed         uint64 = 0b000 << 61
	mUncompressedInfinity uint64 = 0b010 << 61
	mCompressedSmallest   uint64 = 0b100 << 61
	mCompressedLargest    uint64 = 0b101 << 61
	mCompressedInfinity   uint64 = 0b110 << 61
)
{{- else}}
// we have less than 3 bits available on the msw, so we can't follow BLS381 style encoding.
// the difference is the case where a point is infinity and uncompressed is not flagged
const (
	mMask               uint64 = 0b11 << 62
	mUncompressed       uint64 = 0b00 << 62
	mCompressedSmallest uint64 = 0b10 << 62
	mCompressedLargest  uint64 = 0b11 << 62
	mCompressedInfinity uint64 = 0b01 << 62
)
{{- end}}



// Encoder writes {{.CurveName}} object values to an output stream
type Encoder struct {
	w io.Writer
	n int64 		// written bytes
	raw bool 		// raw vs compressed encoding 
}

// Decoder reads {{.CurveName}} object values from an inbound stream
type Decoder struct {
	r io.Reader
	n int64 // read bytes
}

// NewDecoder returns a binary decoder supporting curve {{.CurveName}} objects in both 
// compressed and uncompressed (raw) forms
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}


// Decode reads the binary encoding of v from the stream
// type must be *uint64, *fr.Element, *fp.Element, *G1Affine, *G2Affine, *[]G1Affine or *[]G2Affine
func (dec *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() || !rv.Elem().CanSet() {
		return errors.New("{{.CurveName}} decoder: unsupported type, need pointer")
	}

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap
	// TODO double check memory usage and factorize this

	var buf [SizeOfG2Uncompressed]byte
	var read int
	var msw uint64

	switch t := v.(type) {
	case *uint64:
		msw, err = dec.readUint64()
		if err != nil {
			return
		}
		*t = msw
		return
	case *fr.Element:
		read, err = io.ReadFull(dec.r, buf[:fr.Limbs * 8])
		dec.n += int64(read)
		if err != nil {
			return
		}
		t.SetBytes(buf[:fr.Limbs * 8])
		return
	case *fp.Element:
		read, err = io.ReadFull(dec.r, buf[:fp.Limbs * 8])
		dec.n += int64(read)
		if err != nil {
			return
		}
		t.SetBytes(buf[:fp.Limbs * 8])
		return
	case *G1Affine:
		// read the most significant word
		read, err = io.ReadFull(dec.r, buf[:8])
		dec.n += int64(read)
		if err != nil {
			return
		}
		msw = binary.BigEndian.Uint64(buf[:8])
		nbBytes := SizeOfG1Uncompressed
		if isCompressed(msw) {
			nbBytes = SizeOfG1Compressed
		}
		read, err = io.ReadFull(dec.r, buf[8:nbBytes])
		dec.n += int64(read)
		if err != nil {
			return
		}
		_, err = t.SetBytes(buf[:nbBytes])
		return 
	case *G2Affine:
		read, err = io.ReadFull(dec.r, buf[:8])
		dec.n += int64(read)
		if err != nil {
			return
		}
		msw = binary.BigEndian.Uint64(buf[:8])
		nbBytes := SizeOfG2Uncompressed
		if isCompressed(msw) {
			nbBytes = SizeOfG2Compressed
		}
		read, err = io.ReadFull(dec.r, buf[8:nbBytes])
		dec.n += int64(read)
		if err != nil {
			return
		}
		_ , err = t.SetBytes(buf[:nbBytes])
		return 
	case *[]G1Affine:
		msw, err = dec.readUint64()
		if err != nil {
			return
		}
		if len(*t) != int(msw) {
			*t = make([]G1Affine, msw)
		}
		for i := 0; i < len(*t); i++ {
			// read the most significant word
			read, err = io.ReadFull(dec.r, buf[:8])
			dec.n += int64(read)
			if err != nil {
				return
			}
			msw = binary.BigEndian.Uint64(buf[:8])
			nbBytes := SizeOfG1Uncompressed
			if isCompressed(msw) {
				nbBytes = SizeOfG1Compressed
			}
			read, err = io.ReadFull(dec.r, buf[8:nbBytes])
			dec.n += int64(read)
			if err != nil {
				return
			}
			_, err = (*t)[i].SetBytes(buf[:nbBytes])
			if err != nil {
				return
			}
		}
		return nil
	case *[]G2Affine:
		msw, err = dec.readUint64()
		if err != nil {
			return
		}
		if len(*t) != int(msw) {
			*t = make([]G2Affine, msw)
		}
		for i := 0; i < len(*t); i++ {
			read, err = io.ReadFull(dec.r, buf[:8])
			dec.n += int64(read)
			if err != nil {
				return
			}
			msw = binary.BigEndian.Uint64(buf[:8])
			nbBytes := SizeOfG2Uncompressed
			if isCompressed(msw) {
				nbBytes = SizeOfG2Compressed
			}
			read, err = io.ReadFull(dec.r, buf[8:nbBytes])
			dec.n += int64(read)
			if err != nil {
				return
			}
			_ , err = (*t)[i].SetBytes(buf[:nbBytes])
			if err != nil {
				return
			}
		}
		return nil
	default:
		return errors.New("{{.CurveName}} encoder: unsupported type")
	}
}

func (dec *Decoder) readUint64() (r uint64, err error) {
	var read int
	var buf [8]byte
	read, err = io.ReadFull(dec.r, buf[:8])
	dec.n += int64(read)
	if err != nil {
		return
	}
	r = binary.BigEndian.Uint64(buf[:8])
	return 
}


func isCompressed(msw uint64) bool {
	mData := msw & mMask
	return !((mData == mUncompressed){{- if gt .UnusedBits 3}}||(mData == mUncompressedInfinity) {{- end}})
}


// NewEncoder returns a binary encoder supporting curve {{.CurveName}} objects
func NewEncoder(w io.Writer, options ...func(*Encoder)) *Encoder {
	// default settings
	enc := &Encoder {
		w: w,
		n: 0,
		raw: false,
	}

	// handle options
	for _, option := range options {
		option(enc)
	}

	return enc
}


// Encode writes the binary encoding of v to the stream
// type must be uint64, *fr.Element, *fp.Element, *G1Affine, *G2Affine, []G1Affine or []G2Affine
func (enc *Encoder) Encode(v interface{}) (err error) {
	if enc.raw {
		return enc.encodeRaw(v)
	}
	return enc.encode(v)
}

// WrittenBytes return total bytes written on writer
func (enc *Encoder) WrittenBytes() int64 {
	return enc.n
}


// RawEncoding returns an option to use in NewEncoder(...) which sets raw encoding mode to true
// points will not be compressed using this option
func RawEncoding() func(*Encoder)  {
	return func(enc *Encoder)  {
		enc.raw = true
	}
}

{{template "encode" dict "Raw" ""}}
{{template "encode" dict "Raw" "Raw"}}



{{ define "encode"}}

func (enc *Encoder) encode{{- $.Raw}}(v interface{}) (err error) {

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap
	// TODO double check memory usage and factorize this

	var written int
	switch t := v.(type) {
	case uint64:
		err = binary.Write(enc.w, binary.BigEndian, t)
		enc.n += 8
		return
	case *fr.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return 
	case *fp.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return 
	case *G1Affine:
		buf := t.{{- $.Raw}}Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return  
	case *G2Affine:
		buf := t.{{- $.Raw}}Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return 
	case []G1Affine:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint64(len(t)))
		if err != nil {
			return
		}
		enc.n += 8

		for i := 0; i < len(t); i++ {
			buf := t[i].{{- $.Raw}}Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	case []G2Affine:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint64(len(t)))
		if err != nil {
			return
		}
		enc.n += 8

		for i := 0; i < len(t); i++ {
			buf := t[i].{{- $.Raw}}Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	default:
		return errors.New("{{.CurveName}} encoder: unsupported type")
	}
}
{{end}}

`

// MarshalTests ...
const MarshalTests = `
import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestEncoder(t *testing.T) {

	// TODO need proper fuzz testing here

	var inA uint64
	var inB fr.Element 
	var inC fp.Element 
	var inD G1Affine
	var inE G1Affine
	var inF G2Affine
	var inG []G1Affine
	var inH []G2Affine

	// set values of inputs
	inA = rand.Uint64()
	inB.SetRandom()
	inC.SetRandom()
	inD.ScalarMultiplication(&g1GenAff, new(big.Int).SetUint64(rand.Uint64()))
	// inE --> infinity
	inF.ScalarMultiplication(&g2GenAff, new(big.Int).SetUint64(rand.Uint64()))
	inG = make([]G1Affine, 2)
	inH = make([]G2Affine, 0)
	inG[1] = inD 

	// encode them, compressed and raw
	var buf, bufRaw bytes.Buffer
	enc := NewEncoder(&buf)
	encRaw := NewEncoder(&bufRaw, RawEncoding())
	toEncode := []interface{}{inA, &inB, &inC, &inD, &inE, &inF, inG, inH}
	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		if err := encRaw.Encode(v); err != nil {
			t.Fatal(err)
		}
	}

	
	testDecode := func(t *testing.T, r io.Reader) {
		dec := NewDecoder(r)
		var outA uint64
		var outB fr.Element 
		var outC fp.Element 
		var outD G1Affine
		var outE G1Affine
		outE.X.SetOne()
		outE.Y.SetUint64(42)
		var outF G2Affine
		var outG []G1Affine
		var outH []G2Affine

		toDecode := []interface{}{&outA, &outB, &outC, &outD, &outE, &outF, &outG, &outH}
		for _, v := range toDecode {
			if err := dec.Decode(v); err != nil {
				t.Fatal(err)
			}
		}

		// compare values 
		if inA != outA {
			t.Fatal("didn't encode/decode uint64 value properly")
		}

		if !inB.Equal(&outB) || !inC.Equal(&outC) {
			t.Fatal("decode(encode(Element) failed")
		}
		if !inD.Equal(&outD) || !inE.Equal(&outE) {
			t.Fatal("decode(encode(G1) failed")
		}
		if !inF.Equal(&outF) {
			t.Fatal("decode(encode(G2) failed")
		}
		if (len(inG) != len(outG)) || (len(inH) != len(outH)) {
			t.Fatal("decode(encode(slice(points))) failed")
		}
		for i:=0; i<len(inG);i++ {
			if !inG[i].Equal(&outG[i]) {
				t.Fatal("decode(encode(slice(points))) failed")	
			}
		}
	}

	// decode them 
	testDecode(t, &buf)
	testDecode(t, &bufRaw)


}



func TestIsCompressed(t *testing.T) {
	var g1Inf, g1 G1Affine
	var g2Inf, g2 G2Affine

	g1 = g1GenAff
	g2 = g2GenAff

	{
		b := g1Inf.Bytes() 
		if !isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g1Inf.Bytes() should be compressed")
		}
	}

	{
		b := g1Inf.RawBytes() 
		if isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g1Inf.RawBytes() should be uncompressed")
		}
	}

	{
		b := g1.Bytes() 
		if !isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g1.Bytes() should be compressed")
		}
	}

	{
		b := g1.RawBytes() 
		if isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g1.RawBytes() should be uncompressed")
		}
	}

	

	{
		b := g2Inf.Bytes() 
		if !isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g2Inf.Bytes() should be compressed")
		}
	}

	{
		b := g2Inf.RawBytes() 
		if isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g2Inf.RawBytes() should be uncompressed")
		}
	}

	{
		b := g2.Bytes() 
		if !isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g2.Bytes() should be compressed")
		}
	}

	{
		b := g2.RawBytes() 
		if isCompressed(binary.BigEndian.Uint64(b[:8])) {
			t.Fatal("g2.RawBytes() should be uncompressed")
		}
	}

}

`
