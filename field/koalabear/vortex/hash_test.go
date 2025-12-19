package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/hash"
)

// TODO @gbotrel keeping that here for now, but we should clean up the API and packages
// between this repo and linea-monorepo.

func TestComputeMTLeaves(t *testing.T) {
	sisHashes := make([]koalabear.Element, nbCols*sisKeySize)
	for i := range sisHashes {
		sisHashes[i].MustSetRandom()
	}
	copySisHashes := make([]koalabear.Element, len(sisHashes))
	copy(copySisHashes, sisHashes)

	leaves := computeMTLeaves(sisHashes, nbCols, sisKeySize)
	leaves2 := computeMTLeaves2x16(copySisHashes, nbCols, sisKeySize)

	for i := 0; i < nbCols; i++ {
		if leaves[i] != leaves2[i] {
			t.Fatalf("leaves do not match at index %d", i)
		}
	}
}

func BenchmarkComputeMTLeaves(b *testing.B) {
	sisHashes := make([]koalabear.Element, nbCols*sisKeySize)
	for i := range sisHashes {
		sisHashes[i].MustSetRandom()
	}

	b.Run("computeMDLeaves old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = computeMTLeaves(sisHashes, nbCols, sisKeySize)
		}
	})

	b.Run("computeMDLeaves new 2x16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = computeMTLeaves2x16(sisHashes, nbCols, sisKeySize)
		}
	})
}

const (
	sisKeySize = 512 // elements
	nbCols     = 1024 * 64
)

// reference implementation
func computeMTLeaves(sisHashes []koalabear.Element, nbCols, sisKeySize int) []Hash {
	leaves := make([]Hash, nbCols)

	hasher := newMDHasher()

	for colID := 0; colID < nbCols; colID++ {
		startChunk := colID * sisKeySize
		hasher.Reset()
		hasher.WriteElements(sisHashes[startChunk : startChunk+sisKeySize]...)
		leaves[colID] = hasher.SumElement()
	}
	return leaves
}

func computeMTLeaves2x16(sisHashes []koalabear.Element, nbCols, sisKeySize int) []Hash {
	leaves := make([]Hash, nbCols)

	if nbCols%16 != 0 {
		panic("nbCols must be multiple of 16")
	}
	chunkCol := nbCols / 16

	for chunkID := 0; chunkID < chunkCol; chunkID++ {
		CompressPoseidon2x16(sisHashes[chunkID*16*sisKeySize:(chunkID+1)*16*sisKeySize], sisKeySize, leaves[chunkID*16:(chunkID+1)*16])
	}

	return leaves
}

const BlockSize = 8

// mdHasher Merkle Damgard Hasher using Poseidon2 as compression function
type mdHasher struct {
	hash.StateStorer

	state Hash

	// data to hash
	buffer         [BlockSize]koalabear.Element
	bufferPosition int
}

// newMDHasher creates a new MDHasher with the given options.
func newMDHasher() *mdHasher {
	h := &mdHasher{
		StateStorer: poseidon2.NewMerkleDamgardHasher(),
	}

	return h
}

// Reset clears the buffer, and reset state to iv
func (d *mdHasher) Reset() {
	d.bufferPosition = 0
	d.state = Hash{}
}

// WriteElements adds a slice of field elements to the running hash.
func (d *mdHasher) WriteElements(elmts ...koalabear.Element) {
	// d.buffer has BlockSize slots. Some may already be filled (indicated by d.bufferPosition).
	// We fill up d.buffer, and whenever it gets full, we compress it and reset it.
	// We repeat this until all elmts are consumed.
	// At the end, d.buffer may be partially filled.
	for _, e := range elmts {
		d.buffer[d.bufferPosition] = e
		d.bufferPosition++
		if d.bufferPosition == BlockSize {
			// buffer full, compress
			d.state = CompressPoseidon2(d.state, d.buffer)
			d.bufferPosition = 0
		}
	}
}

func (d *mdHasher) SumElement() Hash {
	if d.bufferPosition == 0 {
		return d.state
	}
	// pad the buffer and compress
	// we need to pad on the left
	var buf [BlockSize]koalabear.Element
	copy(buf[BlockSize-d.bufferPosition:], d.buffer[:d.bufferPosition])
	d.state = CompressPoseidon2(d.state, buf)
	d.bufferPosition = 0
	return d.state
}
