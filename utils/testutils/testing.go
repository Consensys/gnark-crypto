package testutils

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Serializable interface {
	io.ReaderFrom
	io.WriterTo
}

type RawSerializable interface {
	WriteRawTo(io.Writer) (int64, error)
}

type BinaryDumper interface {
	WriteDump(w io.Writer, maxPkPoints ...int) error
	ReadDump(r io.Reader, maxPkPoints ...int) error
}

func SerializationRoundTrip(o Serializable) func(*testing.T) {
	return func(t *testing.T) {
		// serialize it...
		var buf bytes.Buffer
		_, err := o.WriteTo(&buf)
		assert.NoError(t, err)

		// reconstruct the object
		_o := reflect.New(reflect.TypeOf(o).Elem()).Interface().(Serializable)
		_, err = _o.ReadFrom(&buf)
		assert.NoError(t, err)

		// compare
		assert.Equal(t, o, _o)
	}
}

func SerializationRoundTripRaw(o RawSerializable) func(*testing.T) {
	return func(t *testing.T) {
		// serialize it...
		var buf bytes.Buffer
		_, err := o.WriteRawTo(&buf)
		assert.NoError(t, err)

		// reconstruct the object
		_o := reflect.New(reflect.TypeOf(o).Elem()).Interface().(Serializable)
		_, err = _o.ReadFrom(&buf)
		assert.NoError(t, err)

		// compare
		assert.Equal(t, o, _o)
	}
}

func UnsafeBinaryMarshalerRoundTrip(o BinaryDumper) func(*testing.T) {
	return func(t *testing.T) {
		// serialize it...
		var buf bytes.Buffer
		err := o.WriteDump(&buf)
		assert.NoError(t, err)

		// reconstruct the object
		_o := reflect.New(reflect.TypeOf(o).Elem()).Interface().(BinaryDumper)
		err = _o.ReadDump(&buf)
		assert.NoError(t, err)

		// compare
		assert.Equal(t, o, _o)
	}
}
