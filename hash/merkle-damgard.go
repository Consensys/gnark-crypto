package hash

import (
	"bytes"
	"fmt"
	"slices"
)

type merkleDamgardHasher struct {
	state []byte
	iv    []byte
	f     Compressor
}

// Write implements hash.Write
func (h *merkleDamgardHasher) Write(p []byte) (n int, err error) {
	blockSize := h.f.BlockSize()
	for len(p) != 0 {
		if len(p) < blockSize {
			if p, err = cloneLeftPadded(p, blockSize); err != nil {
				panic(err) // this should not be possible
			}
		}
		if h.state, err = h.f.Compress(h.state, p[:blockSize]); err != nil {
			return
		}
		n += blockSize
		p = p[blockSize:]
	}
	return
}

func (h *merkleDamgardHasher) Sum(b []byte) []byte {
	return append(b, h.state...)
}

func (h *merkleDamgardHasher) Reset() {
	h.state = bytes.Clone(h.iv)
}

func (h *merkleDamgardHasher) Size() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) BlockSize() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) State() []byte {
	return h.state
}

// SetState sets h's state to state. If len(state) > BlockSize, an error is thrown.
// If len(state) < BlockSize, h's state is set to state, and left padded with zeroes.
// len(state) > BlockSize will result in an error.
func (h *merkleDamgardHasher) SetState(state []byte) error {
	var err error
	h.state, err = cloneLeftPadded(state, h.BlockSize())
	return err
}

// NewMerkleDamgardHasher transforms a 2-1 one-way compression function into a
// hash function using a Merkle-Damgard construction. The resulting hash
// function has a block size equal to the block size of compression function.
//
// NB! The construction does not perform explicit padding on the input data. The
// last block of input data is zero-padded to full block size. This means that
// the construction is not collision resistant for generic data as the digest of
// input and input concatenated with zeros (up to the same number of total
// blocks) is same. For collision resistance the caller should perform explicit
// padding on the input data.
//
// - initialState is provided as initial input to the compression
// function. Its preimage should not be known and thus it should be generated
// using a deterministic method.
// If the given initialState is shorter than the hash block size, it will be zero-padded
// on the left. An oversized initialState will cause a panic.
func NewMerkleDamgardHasher(f Compressor, initialState []byte) StateStorer {
	iv, err := cloneLeftPadded(initialState, f.BlockSize())
	if err != nil {
		panic(err)
	}
	return &merkleDamgardHasher{
		f:     f,
		iv:    iv,
		state: bytes.Clone(iv),
	}
}

// cloneLeftPadded copies b into a new byte slice of size n.
// If len(b) < n, it will be padded on the left.
// len(b) > n will result in an error.
func cloneLeftPadded(b []byte, n int) ([]byte, error) {
	if len(b) > n {
		return nil, fmt.Errorf("state/iv must not exceed the hash block size: %d > %d", len(b), n)
	}
	res := make([]byte, n)
	copy(res[n-len(b):], b)
	return res, nil
}

// SmallFieldMerkleDamgardHasher is a Merkle-Damgard hash function over blocks of
// elements of a small field F. It is both a [StateStorer], whose bytes are the
// marshalled elements, and a [SmallFieldHasher]. See
// [NewSmallFieldMerkleDamgardHasher].
type SmallFieldMerkleDamgardHasher[F any, PF interface {
	*F
	SetBytesCanonical([]byte) error
	Marshal() []byte
}] struct {
	f        SmallFieldCompressor[F]
	iv       []F
	state    []F
	elemSize int // size of a marshaled element, in bytes
}

// NewSmallFieldMerkleDamgardHasher transforms a 2-1 one-way compression
// function over elements of a small field into a hash function using a
// Merkle-Damgard construction. The block size of the hasher is the block size
// of the compression function, and its digest is a block.
//
// NB! The construction does not perform explicit padding on the input data,
// see [NewMerkleDamgardHasher]. The last block of input is padded with zeros on
// the left to a full block, so the digests of an input and of the input
// preceded by zeros in its last block are the same.
//
// Write compresses each block of its input, and pads its last block, which may
// be partial, with zeros on the left to a full block, so that the digest depends
// on how the input is split among the calls.
//
// initialState is provided as initial input to the compression function. Its
// preimage should not be known and thus it should be generated using a
// deterministic method. If it is shorter than a block, it is padded with zeros
// on the left. An oversized initialState causes a panic.
func NewSmallFieldMerkleDamgardHasher[F any, PF interface {
	*F
	SetBytesCanonical([]byte) error
	Marshal() []byte
}](f SmallFieldCompressor[F], initialState []F) *SmallFieldMerkleDamgardHasher[F, PF] {
	blockLen := f.BlockLen()
	if len(initialState) > blockLen {
		panic(fmt.Errorf("state/iv must not exceed the hash block size: %d > %d", len(initialState), blockLen))
	}
	iv := make([]F, blockLen)
	copy(iv[blockLen-len(initialState):], initialState)

	var e F
	return &SmallFieldMerkleDamgardHasher[F, PF]{
		f:        f,
		iv:       iv,
		state:    slices.Clone(iv),
		elemSize: len(PF(&e).Marshal()),
	}
}

// WriteField implements [SmallFieldHasher]. It pads a partial last block with
// zeros on the left to a full block.
func (h *SmallFieldMerkleDamgardHasher[F, PF]) WriteField(e ...F) {
	blockLen := h.f.BlockLen()
	for len(e) != 0 {
		block := e
		if len(e) >= blockLen {
			block, e = e[:blockLen], e[blockLen:]
		} else {
			block = make([]F, blockLen)
			copy(block[blockLen-len(e):], e)
			e = nil
		}
		h.f.CompressSmallField(h.state, h.state, block)
	}
}

// BlockLen implements [SmallFieldHasher].
func (h *SmallFieldMerkleDamgardHasher[F, PF]) BlockLen() int {
	return h.f.BlockLen()
}

// SumSmallField implements [SmallFieldHasher].
func (h *SmallFieldMerkleDamgardHasher[F, PF]) SumSmallField() []F {
	return slices.Clone(h.state)
}

// Write implements hash.Write
func (h *SmallFieldMerkleDamgardHasher[F, PF]) Write(p []byte) (n int, err error) {
	blockSize := h.BlockSize()
	block := make([]F, h.f.BlockLen())
	for len(p) != 0 {
		chunk := p[:min(len(p), blockSize)]
		if err = h.parseBlock(block, chunk); err != nil {
			return
		}
		h.WriteField(block...)
		n += blockSize
		p = p[len(chunk):]
	}
	return
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) Sum(b []byte) []byte {
	return h.appendMarshalled(b, h.SumSmallField())
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) Reset() {
	copy(h.state, h.iv)
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) Size() int {
	return h.BlockSize()
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) BlockSize() int {
	return h.f.BlockLen() * h.elemSize
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) State() []byte {
	return h.Sum(nil)
}

// SetState sets h's state to state. If len(state) < BlockSize, h's state is set
// to state, left padded with zeroes. len(state) > BlockSize will result in an
// error.
func (h *SmallFieldMerkleDamgardHasher[F, PF]) SetState(state []byte) error {
	newState := make([]F, len(h.state))
	if err := h.parseBlock(newState, state); err != nil {
		return err
	}
	h.state = newState
	return nil
}

// parseBlock parses b, padded with zeros on the left to a block, into the
// elements of dst. It is an error if b is longer than a block.
func (h *SmallFieldMerkleDamgardHasher[F, PF]) parseBlock(dst []F, b []byte) error {
	blockSize := h.BlockSize()
	if len(b) > blockSize {
		return fmt.Errorf("state/iv must not exceed the hash block size: %d > %d", len(b), blockSize)
	}
	if len(b) < blockSize {
		padded := make([]byte, blockSize)
		copy(padded[blockSize-len(b):], b)
		b = padded
	}
	for i := range dst {
		if err := PF(&dst[i]).SetBytesCanonical(b[i*h.elemSize : (i+1)*h.elemSize]); err != nil {
			return err
		}
	}
	return nil
}

func (h *SmallFieldMerkleDamgardHasher[F, PF]) appendMarshalled(b []byte, e []F) []byte {
	for i := range e {
		b = append(b, PF(&e[i]).Marshal()...)
	}
	return b
}
