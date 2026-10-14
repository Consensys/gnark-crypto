package hash

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/hash"
)

// BaseFieldHashAdapterE6 hashes elements of E6 with a hasher over the base
// field. The six coordinates of each element are written in the order
// B0.A0, B0.A1, B1.A0, B1.A1, B2.A0, B2.A1. They are packed in blocks of
// blockSize base field elements, so a write of the base hasher does not
// correspond to a write of the adapter.
type BaseFieldHashAdapterE6 struct {
	hsh hash.SmallFieldHasher[koalabear.Element]
	// buffer holds the pending coordinates, fewer than blockSize between calls,
	// and room for the six coordinates of one more element.
	buffer     [blockSize - 1 + 6]koalabear.Element
	nbBuffered int
}

// blockSize is the number of base field elements per write to the base hasher.
const blockSize = 8

func NewBaseFieldHashAdapterE6(baseHasher hash.SmallFieldHasher[koalabear.Element]) *BaseFieldHashAdapterE6 {
	return &BaseFieldHashAdapterE6{hsh: baseHasher}
}

// WriteField buffers the coordinates of e, and writes each full block to the
// base hasher.
func (h *BaseFieldHashAdapterE6) WriteField(e ...extensions.E6) {
	for i := range e {
		h.buffer[h.nbBuffered] = e[i].B0.A0
		h.buffer[h.nbBuffered+1] = e[i].B0.A1
		h.buffer[h.nbBuffered+2] = e[i].B1.A0
		h.buffer[h.nbBuffered+3] = e[i].B1.A1
		h.buffer[h.nbBuffered+4] = e[i].B2.A0
		h.buffer[h.nbBuffered+5] = e[i].B2.A1
		h.nbBuffered += 6

		if h.nbBuffered >= blockSize {
			h.hsh.WriteSmallField(h.buffer[:blockSize]...)
			h.nbBuffered = copy(h.buffer[:], h.buffer[blockSize:h.nbBuffered])
		}
	}
}

// SumField writes the buffered coordinates to the base hasher, and returns the
// first six elements of its digest as an element of E6, in the order of the
// coordinates of WriteField.
func (h *BaseFieldHashAdapterE6) SumField() extensions.E6 {
	h.flush()
	d := h.hsh.SumSmallField()

	var res extensions.E6
	res.B0.A0, res.B0.A1, res.B1.A0, res.B1.A1, res.B2.A0, res.B2.A1 = d[0], d[1], d[2], d[3], d[4], d[5]
	return res
}

func (h *BaseFieldHashAdapterE6) flush() {
	if h.nbBuffered > 0 {
		h.hsh.WriteSmallField(h.buffer[:h.nbBuffered]...)
		h.nbBuffered = 0
	}
}
