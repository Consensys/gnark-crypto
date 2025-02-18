package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

// Hash represents a hash as they occur in Merkle trees
type Hash [8]koalabear.Element

// MerkleTree represents a Merkle tree.
type MerkleTree struct {
	// Levels collects the nodes of the tree in descending order:
	// The first level has a size of 1 and stores the root of the tree
	// The last level has a size of 1 << Depth and stores the leaves
	Levels [][]Hash

	Hasher poseidon2.Permutation
}

// MerkleProof is a Merkle proof that can be used to verify the membership
// of an element in the tree. The proof is a list of hashes in ascending order.
// i.e. the first hash is the immediate neighbor of the opened leaf and the
// last one is the one just under the root. So it has a length of depth.
type MerkleProof []Hash

// MerkleCompute builds a Merkle tree from a list of hashes. If the provided
// number of leaves is not a power of two, the leaves are padded with zero
// hashes.
func MerkleCompute(hashes []Hash) *MerkleTree {

	var (
		numLeaves    = len(hashes)
		newPow2      = NextPowerOfTwo(numLeaves)
		depth        = Log2Ceil(numLeaves)
		paddedHashes = hashes
	)

	if len(hashes) != newPow2 {
		paddedHashes = make([]Hash, newPow2)
		copy(paddedHashes, hashes)
	}

	levels := make([][]Hash, depth+1)
	for i := depth; i >= 0; i-- {

		if i == depth {
			levels[i] = paddedHashes
			continue
		}

		levels[i] = make([]Hash, newPow2>>(depth-i))
		for k := range levels[i] {
			left, right := levels[i+1][2*k], levels[i+1][2*k+1]
			levels[i][k] = CompressPoseidon2(left, right)
		}
	}

	return &MerkleTree{
		Levels: levels,
	}
}

// Open returns the Merkle proof for the element at index i, returns an error
// of the index is out of range.
func (mt *MerkleTree) Open(i int) (MerkleProof, error) {

	var (
		res       = make(MerkleProof, 0, mt.Depth())
		parentPos = i
		posBound  = 1 << mt.Depth()
	)

	if i >= posBound {
		return nil, errors.New("error: index out of range")
	}

	for level := len(mt.Levels) - 1; level > 0; level-- {
		fmt.Printf("level %v\n", level)
		neighborPos := parentPos ^ 1
		res = append(res, mt.Levels[level][neighborPos])
		parentPos = parentPos >> 1
	}

	// sanity-checking that we have the expected number of elements
	if len(res) != mt.Depth() {
		panic("error: incorrect number of hashes")
	}

	return res, nil
}

// Verify checks the validity of a merkle membership proof. Returns nil
// if it passes and an error indicating the failed check.
func (proof MerkleProof) Verify(i int, leaf, root Hash) error {

	var (
		parentPos = i
		curNode   = leaf
	)

	for _, h := range proof {

		a, b := leaf, h
		if parentPos&1 == 1 {
			a, b = b, a
		}

		curNode = CompressPoseidon2(a, b)
		parentPos = parentPos >> 1
	}

	if curNode != root {
		return errors.New("error: invalid proof")
	}

	return nil
}

// Depth returns the depth of the tree. A tree of depth n has 2^n leaves.
func (mt *MerkleTree) Depth() int {
	return len(mt.Levels) - 1
}

// Root returns the root of the tree
func (mt *MerkleTree) Root() Hash {
	return mt.Levels[0][0]
}

// Return true if n is a power of two
func IsPowerOfTwo[T ~int](n T) bool {
	return n&(n-1) == 0 && n > 0
}

/*
NextPowerOfTwo returns the next power of two for the given number.
It returns the number itself if it's a power of two. As an edge case,
zero returns zero.

Taken from :
https://github.com/protolambda/zrnt/blob/v0.13.2/eth2/util/math/math_util.go#L58
The function panics if the input is more than  2**62 as this causes overflow
*/
func NextPowerOfTwo[T ~int64 | ~uint64 | ~uintptr | ~int | ~uint](in T) T {
	if in < 0 || uint64(in) > 1<<62 {
		panic("input out of range")
	}
	v := in
	v--
	v |= v >> (1 << 0)
	v |= v >> (1 << 1)
	v |= v >> (1 << 2)
	v |= v >> (1 << 3)
	v |= v >> (1 << 4)
	v |= v >> (1 << 5)
	v++
	return v
}

// Log2Floor computes the floored value of Log2
func Log2Floor(a int) int {
	res := 0
	for i := a; i > 1; i = i >> 1 {
		res++
	}
	return res
}

// Log2Ceil computes the ceiled value of Log2
func Log2Ceil(a int) int {
	floor := Log2Floor(a)
	if a != 1<<floor {
		floor++
	}
	return floor
}
