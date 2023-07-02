package element

const Vector = `
import (
	"io"
	"encoding/binary"
	"strings"
	"bytes"
	"unsafe"
	"sync/atomic"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"fmt"
)

// Vector represents a slice of {{.ElementName}}.
// 
// It implements the following interfaces:
//	- Stringer
//	- io.WriterTo
//	- io.ReaderFrom
//	- encoding.BinaryMarshaler
//	- encoding.BinaryUnmarshaler
//	- sort.Interface
type Vector []{{.ElementName}}

// MarshalBinary implements encoding.BinaryMarshaler
func (vector *Vector) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	if _, err = vector.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}


// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (vector *Vector) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	_, err := vector.ReadFrom(r)
	return err
}

// WriteTo implements io.WriterTo and writes a vector of big endian encoded {{.ElementName}}.
// Length of the vector is encoded as a uint32 on the first 4 bytes.
func (vector *Vector) WriteTo(w io.Writer) (int64, error) {
    // encode slice length
    if err := binary.Write(w, binary.BigEndian, uint32(len(*vector))); err != nil {
        return 0, err 
    }

	n := int64(4)

	var buf [Bytes]byte 
	for i := 0; i < len(*vector); i++ {
		BigEndian.PutElement(&buf, (*vector)[i])
		m, err := w.Write(buf[:])
		n += int64(m)
		if err != nil {
			return n, err 
		} 
	}
	return n, nil
}

// AsyncReadFrom reads a vector of big endian encoded {{.ElementName}}.
// Length of the vector must be encoded as a uint32 on the first 4 bytes.
// It consumes the needed bytes from the reader and returns the number of bytes read and an error if any.
// It also returns a channel that will be closed when the validation is done.
// The validation consist of checking that the elements are smaller than the modulus, and
// converting them to montgomery form.
func (vector *Vector) AsyncReadFrom(r io.Reader) (int64, error, chan error) {
	chErr := make(chan error, 1)
	var buf [Bytes]byte 
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
		close(chErr)
        return int64(read), err, chErr
    }
	sliceLen := binary.BigEndian.Uint32(buf[:4])

    n := int64(4)
	(*vector) = make(Vector, sliceLen)
	if sliceLen == 0 {
		close(chErr)
		return n, nil, chErr
	}

	bSlice := unsafe.Slice((*byte)(unsafe.Pointer(&(*vector)[0])), sliceLen*Bytes)
	read, err := io.ReadFull(r, bSlice)
	n += int64(read)
	if err != nil {
		close(chErr)
		return n, err, chErr
	}


	go func() {
		var cptErrors uint64
		// process the elements in parallel
		parallel.Execute(int(sliceLen), func(start, end int) {
			
			var z Element
			for i:=start; i < end; i++ {
				// we have to set vector[i]
				bstart := i*Bytes
				bend := bstart + Bytes
				b := bSlice[bstart:bend]
				{{- range $i := reverse .NbWordsIndexesFull}}
					{{- $j := mul $i 8}}
					{{- $k := sub $.NbWords 1}}
					{{- $k := sub $k $i}}
					{{- $jj := add $j 8}}
					z[{{$k}}] = binary.BigEndian.Uint64(b[{{$j}}:{{$jj}}])
				{{- end}}

				if !z.smallerThanModulus() {
					atomic.AddUint64(&cptErrors, 1)
					return
				}
				z.toMont()
				(*vector)[i] = z
			}
		})

		if cptErrors > 0 {
			chErr <- fmt.Errorf("async read: %d elements failed validation", cptErrors)
		}
		close(chErr)
	}()
	return n, nil, chErr
}

// ReadFrom implements io.ReaderFrom and reads a vector of big endian encoded {{.ElementName}}.
// Length of the vector must be encoded as a uint32 on the first 4 bytes.
func (vector *Vector) ReadFrom(r io.Reader) (int64, error) {

	var buf [Bytes]byte 
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
        return int64(read), err 
    }
	sliceLen := binary.BigEndian.Uint32(buf[:4])

    n := int64(4)
	(*vector) = make(Vector, sliceLen)

    for i:=0; i < int(sliceLen); i++ {
        read, err := io.ReadFull(r, buf[:])
        n += int64(read)
        if err != nil {
            return n, err
        }
		(*vector)[i], err = BigEndian.Element(&buf)
		if err != nil {
			return n, err
		}
    }
	

    return n, nil 
}

// String implements fmt.Stringer interface
func (vector Vector) String() string {
    var sbb strings.Builder
    sbb.WriteByte('[')
    for i:=0; i < len(vector); i++ {
        sbb.WriteString(vector[i].String())
		if i != len(vector) - 1 {
			sbb.WriteByte(',')
		}
    }
    sbb.WriteByte(']')
    return sbb.String()
}


// Len is the number of elements in the collection.
func (vector Vector) Len() int {
	return len(vector)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (vector Vector) Less(i, j int) bool {
	return vector[i].Cmp(&vector[j]) == -1
}

// Swap swaps the elements with indexes i and j.
func (vector Vector) Swap(i, j int) {
	vector[i], vector[j] = vector[j], vector[i]
}

`
