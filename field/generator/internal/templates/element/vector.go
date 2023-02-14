package element

const Vector = `
import (
	"io"
	"encoding/binary"
	"strings"
	"bytes"
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
