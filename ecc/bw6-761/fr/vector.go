// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Vector represents a slice of Element.
//
// It implements the following interfaces:
//   - Stringer
//   - io.WriterTo
//   - io.ReaderFrom
//   - encoding.BinaryMarshaler
//   - encoding.BinaryUnmarshaler
//   - sort.Interface
type Vector []Element

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

// WriteTo implements io.WriterTo and writes a vector of big endian encoded Element.
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

// AsyncReadFrom reads a vector of big endian encoded Element.
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
		execute(int(sliceLen), func(start, end int) {

			var z Element
			for i := start; i < end; i++ {
				// we have to set vector[i]
				bstart := i * Bytes
				bend := bstart + Bytes
				b := bSlice[bstart:bend]
				z[0] = binary.BigEndian.Uint64(b[40:48])
				z[1] = binary.BigEndian.Uint64(b[32:40])
				z[2] = binary.BigEndian.Uint64(b[24:32])
				z[3] = binary.BigEndian.Uint64(b[16:24])
				z[4] = binary.BigEndian.Uint64(b[8:16])
				z[5] = binary.BigEndian.Uint64(b[0:8])

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

// ReadFrom implements io.ReaderFrom and reads a vector of big endian encoded Element.
// Length of the vector must be encoded as a uint32 on the first 4 bytes.
func (vector *Vector) ReadFrom(r io.Reader) (int64, error) {

	var buf [Bytes]byte
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
		return int64(read), err
	}
	sliceLen := binary.BigEndian.Uint32(buf[:4])

	n := int64(4)
	(*vector) = make(Vector, sliceLen)

	for i := 0; i < int(sliceLen); i++ {
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
	for i := 0; i < len(vector); i++ {
		sbb.WriteString(vector[i].String())
		if i != len(vector)-1 {
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

// Add adds two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Add(a, b Vector) {
	addVecGeneric(*vector, a, b)
}

// Sub subtracts two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Sub(a, b Vector) {
	subVecGeneric(*vector, a, b)
}

// ScalarMul multiplies a vector by a scalar element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) ScalarMul(a Vector, b *Element) {
	scalarMulVecGeneric(*vector, a, b)
}

func addVecGeneric(res, a, b Vector) {
	if len(a) != len(b) || len(a) != len(res) {
		panic("vector.Add: vectors don't have the same length")
	}
	for i := 0; i < len(a); i++ {
		res[i].Add(&a[i], &b[i])
	}
}

func subVecGeneric(res, a, b Vector) {
	if len(a) != len(b) || len(a) != len(res) {
		panic("vector.Sub: vectors don't have the same length")
	}
	for i := 0; i < len(a); i++ {
		res[i].Sub(&a[i], &b[i])
	}
}

func scalarMulVecGeneric(res, a Vector, b *Element) {
	if len(a) != len(res) {
		panic("vector.ScalarMul: vectors don't have the same length")
	}
	for i := 0; i < len(a); i++ {
		res[i].Mul(&a[i], b)
	}
}

// TODO @gbotrel make a public package out of that.
// execute executes the work function in parallel.
// this is copy paste from internal/parallel/parallel.go
// as we don't want to generate code importing internal/
func execute(nbIterations int, work func(int, int), maxCpus ...int) {

	nbTasks := runtime.NumCPU()
	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
		if nbTasks < 1 {
			nbTasks = 1
		} else if nbTasks > 512 {
			nbTasks = 512
		}
	}

	if nbTasks == 1 {
		// no go routines
		work(0, nbIterations)
		return
	}

	nbIterationsPerCpus := nbIterations / nbTasks

	// more CPUs than tasks: a CPU will work on exactly one iteration
	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
		nbTasks = nbIterations
	}

	var wg sync.WaitGroup

	extraTasks := nbIterations - (nbTasks * nbIterationsPerCpus)
	extraTasksOffset := 0

	for i := 0; i < nbTasks; i++ {
		wg.Add(1)
		_start := i*nbIterationsPerCpus + extraTasksOffset
		_end := _start + nbIterationsPerCpus
		if extraTasks > 0 {
			_end++
			extraTasks--
			extraTasksOffset++
		}
		go func() {
			work(_start, _end)
			wg.Done()
		}()
	}

	wg.Wait()
}
