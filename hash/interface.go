package hash

import "hash"

// StateStorer allows to store and retrieve the state of a hash function.
type StateStorer interface {
	hash.Hash

	// State retrieves the current state of the hash function. Calling this
	// method should not destroy the current state and allow continue the use of
	// the current hasher.
	State() []byte
	// SetState sets the state of the hash function from a previously stored
	// state retrieved using [StateStorer.State] method.
	SetState(state []byte) error
}

// Compressor is a 2-1 one-way function. It takes two inputs and compresses
// them into one output. The inputs and outputs are all of the same size, which
// is the block size. See [BlockSize].
//
// NB! This is lossy compression, meaning that the output is not guaranteed to
// be unique for different inputs. The function must be stateless, meaning that
// the output is guaranteed to be the same for the same inputs.
// It must furthermore not modify its input, even temporarily.
//
// The Compressor is used in the Merkle-Damgard construction to build a hash
// function.
type Compressor interface {
	// Compress compresses the two inputs into one output. All the inputs and
	// outputs are of the same size, which is the block size. See [BlockSize].
	Compress(left []byte, right []byte) (compressed []byte, err error)
	// BlockSize returns the blocks size.
	BlockSize() int
}

// FieldHasher hashes elements of a field F into a single element, which is
// enough for a digest.
type FieldHasher[F any] interface {
	// WriteField absorbs the elements into the hash function.
	WriteField(...F)
	// SumField returns the digest of the elements written so far.
	SumField() F
}

// SmallFieldHasher hashes elements of a small field F into several elements,
// as a single element is too small for a digest. It is also a [StateStorer], whose
// bytes are the marshalled elements.
type SmallFieldHasher[F any] interface {
	StateStorer
	// WriteField absorbs the elements into the hash function.
	WriteField(...F)
	// SumSmallField returns the digest of the elements written so far.
	SumSmallField() []F
	// BlockLen returns the number of elements in a block of the hasher. If the
	// hasher pads the last block of each write, the digest depends on how the
	// input is split among the writes unless they are multiples of BlockLen. A
	// hasher with no such requirement returns 1.
	BlockLen() int
}

// SmallFieldCompressor is a 2-1 one-way function over blocks of elements of a
// small field F. It is the analogue of [Compressor] on elements.
type SmallFieldCompressor[F any] interface {
	// CompressSmallField compresses left and right into dst. All three have
	// [SmallFieldCompressor.BlockLen] elements, and dst may alias left or right.
	// It must not modify left or right unless they alias dst.
	CompressSmallField(dst, left, right []F)
	// BlockLen returns the number of elements in a block.
	BlockLen() int
}
