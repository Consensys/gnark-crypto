package point

// Marshal ...
const Marshal = `
import (
	"io"
	"math/big"
	"runtime"
	"encoding/binary"

	"github.com/consensys/gurvy/{{ toLower .Name}}/fp"
	"github.com/consensys/gurvy/{{ toLower .Name}}/fr"
	"github.com/consensys/gurvy/utils/debug"
)


// To encode G1 and G2 points, we mask the most significant bits with these bits to specify without ambiguity
// metadata needed for point (de)compression
{{- if ge .FpUnusedBits 3}}
// we follow the BLS381 style encoding as specified in ZCash and now IETF
// The most significant bit, when set, indicates that the point is in compressed form. Otherwise, the point is in uncompressed form.
// The second-most significant bit indicates that the point is at infinity. If this bit is set, the remaining bits of the group element's encoding should be set to zero.
// The third-most significant bit is set if (and only if) this point is in compressed form and it is not the point at infinity and its y-coordinate is the lexicographically largest of the two associated with the encoded x-coordinate.
const (
	mMask                 byte = 0b111 << 5
	mUncompressed         byte = 0b000 << 5
	mUncompressedInfinity byte = 0b010 << 5
	mCompressedSmallest   byte = 0b100 << 5
	mCompressedLargest    byte = 0b101 << 5
	mCompressedInfinity   byte = 0b110 << 5
)
{{- else}}
// we have less than 3 bits available on the msw, so we can't follow BLS381 style encoding.
// the difference is the case where a point is infinity and uncompressed is not flagged
const (
	mMask               byte = 0b11 << 6
	mUncompressed       byte = 0b00 << 6
	mCompressedSmallest byte = 0b10 << 6
	mCompressedLargest  byte = 0b11 << 6
	mCompressedInfinity byte = 0b01 << 6
)
{{- end}}



// Encoder writes {{.Name}} object values to an output stream
type Encoder struct {
	w io.Writer
	n int64 		// written bytes
	raw bool 		// raw vs compressed encoding 
}

// Decoder reads {{.Name}} object values from an inbound stream
type Decoder struct {
	r io.Reader
	n int64 // read bytes
}

// NewDecoder returns a binary decoder supporting curve {{.Name}} objects in both 
// compressed and uncompressed (raw) forms
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}


// Decode reads the binary encoding of v from the stream
// type must be *uint64, *fr.Element, *fp.Element, *G1, *G2, *[]G1 or *[]G2
func (dec *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() || !rv.Elem().CanSet() {
		return errors.New("{{.Name}} decoder: unsupported type, need pointer")
	}

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap
	// TODO double check memory usage and factorize this

	var buf [SizeOfG2Uncompressed]byte
	var read int

	switch t := v.(type) {
	case *uint64:
		var r uint64
		r, err = dec.readUint64()
		if err != nil {
			return
		}
		*t = r
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
	case *G1:
		// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
		read, err = io.ReadFull(dec.r, buf[:SizeOfG1Compressed])
		dec.n += int64(read)
		if err != nil {
			return
		}
		nbBytes := SizeOfG1Compressed
		// most significant byte contains metadata 
		if !isCompressed(buf[0]) {
			nbBytes = SizeOfG1Uncompressed
			// we read more. 
			read, err = io.ReadFull(dec.r, buf[SizeOfG1Compressed:SizeOfG1Uncompressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
		}
		_, err = t.SetBytes(buf[:nbBytes])
		return 
	case *G2:
		// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
		read, err = io.ReadFull(dec.r, buf[:SizeOfG2Compressed])
		dec.n += int64(read)
		if err != nil {
			return
		}
		nbBytes := SizeOfG2Compressed
		// most significant byte contains metadata 
		if !isCompressed(buf[0]) {
			nbBytes = SizeOfG2Uncompressed
			// we read more. 
			read, err = io.ReadFull(dec.r, buf[SizeOfG2Compressed:SizeOfG2Uncompressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
		}
		_, err = t.SetBytes(buf[:nbBytes])
		return 
	case *[]G1:
		var sliceLen uint32
		sliceLen, err = dec.readUint32()
		if err != nil {
			return
		}
		if len(*t) != int(sliceLen) {
			*t = make([]G1, sliceLen)
		}
		compressed := make([]bool, sliceLen)
		for i := 0; i < len(*t); i++ {

			// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
			read, err = io.ReadFull(dec.r, buf[:SizeOfG1Compressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
			nbBytes := SizeOfG1Compressed
			// most significant byte contains metadata 
			if !isCompressed(buf[0]) {
				nbBytes = SizeOfG1Uncompressed
				// we read more. 
				read, err = io.ReadFull(dec.r, buf[SizeOfG1Compressed:SizeOfG1Uncompressed])
				dec.n += int64(read)
				if err != nil {
					return
				}
				_, err = (*t)[i].SetBytes(buf[:nbBytes])
				if err != nil {
					return
				}
			} else {
				compressed[i] = !((*t)[i].unsafeSetCompressedBytes(buf[:nbBytes]))
			}
		}
		var nbErrs uint64
		parallel.Execute(len(compressed), func(start, end int){
			for i := start; i < end; i++ {
				if compressed[i] {
					if err := (*t)[i].unsafeComputeY(); err != nil {
						atomic.AddUint64(&nbErrs, 1)
					}
				}
			}
		})
		if nbErrs != 0 {
			return errors.New("point decompression failed")
		}
		
		return nil
	case *[]G2:
		var sliceLen uint32
		sliceLen, err = dec.readUint32()
		if err != nil {
			return
		}
		if len(*t) != int(sliceLen) {
			*t = make([]G2, sliceLen)
		}
		compressed := make([]bool, sliceLen)
		for i := 0; i < len(*t); i++ {

			// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
			read, err = io.ReadFull(dec.r, buf[:SizeOfG2Compressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
			nbBytes := SizeOfG2Compressed
			// most significant byte contains metadata 
			if !isCompressed(buf[0]) {
				nbBytes = SizeOfG2Uncompressed
				// we read more. 
				read, err = io.ReadFull(dec.r, buf[SizeOfG2Compressed:SizeOfG2Uncompressed])
				dec.n += int64(read)
				if err != nil {
					return
				}
				_, err = (*t)[i].SetBytes(buf[:nbBytes])
				if err != nil {
					return
				}
			} else {
				compressed[i] = !((*t)[i].unsafeSetCompressedBytes(buf[:nbBytes]))
			}
		}
		var nbErrs uint64
		parallel.Execute(len(compressed), func(start, end int){
			for i := start; i < end; i++ {
				if compressed[i] {
					if err := (*t)[i].unsafeComputeY(); err != nil {
						atomic.AddUint64(&nbErrs, 1)
					}
				}
			}
		})
		if nbErrs != 0 {
			return errors.New("point decompression failed")
		}
		
		return nil
	default:
		return errors.New("{{.Name}} encoder: unsupported type")
	}
}

// BytesRead return total bytes read from reader
func (dec *Decoder) BytesRead() int64 {
	return dec.n
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

func (dec *Decoder) readUint32() (r uint32, err error) {
	var read int
	var buf [4]byte
	read, err = io.ReadFull(dec.r, buf[:4])
	dec.n += int64(read)
	if err != nil {
		return
	}
	r = binary.BigEndian.Uint32(buf[:4])
	return 
}


func isCompressed(msb byte) bool {
	mData := msb & mMask
	return !((mData == mUncompressed){{- if ge .FpUnusedBits 3}}||(mData == mUncompressedInfinity) {{- end}})
}


// NewEncoder returns a binary encoder supporting curve {{.Name}} objects
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
// type must be uint64, *fr.Element, *fp.Element, *G1, *G2, []G1 or []G2
func (enc *Encoder) Encode(v interface{}) (err error) {
	if enc.raw {
		return enc.encodeRaw(v)
	}
	return enc.encode(v)
}

// BytesWritten return total bytes written on writer
func (enc *Encoder) BytesWritten() int64 {
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
	case *G1:
		buf := t.{{- $.Raw}}Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return  
	case *G2:
		buf := t.{{- $.Raw}}Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return 
	case []G1:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4

		var buf [SizeOfG1{{- if $.Raw}}Uncompressed{{- else}}Compressed{{- end}}]byte

		for i := 0; i < len(t); i++ {
			buf = t[i].{{- $.Raw}}Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	case []G2:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4

		var buf [SizeOfG2{{- if $.Raw}}Uncompressed{{- else}}Compressed{{- end}}]byte

		for i := 0; i < len(t); i++ {
			buf = t[i].{{- $.Raw}}Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	default:
		return errors.New("{{.Name}} encoder: unsupported type")
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
	var inD G1
	var inE G1
	var inF G2
	var inG []G1
	var inH []G2

	// set values of inputs
	inA = rand.Uint64()
	inB.SetRandom()
	inC.SetRandom()
	inD.ScalarMultiplication(&g1GenAff, new(big.Int).SetUint64(rand.Uint64()))
	// inE --> infinity
	inF.ScalarMultiplication(&g2GenAff, new(big.Int).SetUint64(rand.Uint64()))
	inG = make([]G1, 2)
	inH = make([]G2, 0)
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

	
	testDecode := func(t *testing.T, r io.Reader, n int64) {
		dec := NewDecoder(r)
		var outA uint64
		var outB fr.Element 
		var outC fp.Element 
		var outD G1
		var outE G1
		outE.X.SetOne()
		outE.Y.SetUint64(42)
		var outF G2
		var outG []G1
		var outH []G2

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
		if n != dec.BytesRead() {
			t.Fatal("bytes read don't match bytes written")
		}
	}

	// decode them 
	testDecode(t, &buf, enc.BytesWritten())
	testDecode(t, &bufRaw, encRaw.BytesWritten())


}



func TestIsCompressed(t *testing.T) {
	var g1Inf, g1 G1
	var g2Inf, g2 G2

	g1 = g1GenAff
	g2 = g2GenAff

	{
		b := g1Inf.Bytes() 
		if !isCompressed(b[0]) {
			t.Fatal("g1Inf.Bytes() should be compressed")
		}
	}

	{
		b := g1Inf.RawBytes() 
		if isCompressed(b[0]) {
			t.Fatal("g1Inf.RawBytes() should be uncompressed")
		}
	}

	{
		b := g1.Bytes() 
		if !isCompressed(b[0]) {
			t.Fatal("g1.Bytes() should be compressed")
		}
	}

	{
		b := g1.RawBytes() 
		if isCompressed(b[0]) {
			t.Fatal("g1.RawBytes() should be uncompressed")
		}
	}

	

	{
		b := g2Inf.Bytes() 
		if !isCompressed(b[0]) {
			t.Fatal("g2Inf.Bytes() should be compressed")
		}
	}

	{
		b := g2Inf.RawBytes() 
		if isCompressed(b[0]) {
			t.Fatal("g2Inf.RawBytes() should be uncompressed")
		}
	}

	{
		b := g2.Bytes() 
		if !isCompressed(b[0]) {
			t.Fatal("g2.Bytes() should be compressed")
		}
	}

	{
		b := g2.RawBytes() 
		if isCompressed(b[0]) {
			t.Fatal("g2.RawBytes() should be uncompressed")
		}
	}

}

`
