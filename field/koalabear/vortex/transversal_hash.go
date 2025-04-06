package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// transversalHash hashes the columns of the codewords in parallel
// using the SIS hash function.
func transversalHash(codewords []koalabear.Element, s *sis.RSis, sizeCodeWord int) []koalabear.Element {
	nbCols := sizeCodeWord
	nbRows := len(codewords) / sizeCodeWord
	sisKeySize := s.Degree

	res := make([]koalabear.Element, nbCols*sisKeySize)

	parallel.Execute(nbCols, func(start, end int) {
		// we transpose the columns using a windowed approach
		// this is done to improve memory accesses when transposing the matrix

		// perf note; we could allocate only blocks of 256 elements here and do the SIS hash
		// block by block, but surprisingly it is slower than the current implementation
		// it would however save some memory allocation.
		windowSize := 4
		n := end - start
		for n%windowSize != 0 {
			windowSize /= 2
		}
		transposed := make([][]koalabear.Element, windowSize)
		for i := range transposed {
			transposed[i] = make([]koalabear.Element, nbRows)
		}
		for col := start; col < end; col += windowSize {
			for i := 0; i < nbRows; i++ {
				for j := range transposed {
					transposed[j][i] = codewords[i*sizeCodeWord+col+j]
				}
			}
			for j := range transposed {
				s.Hash(transposed[j], res[(col+j)*sisKeySize:(col+j)*sisKeySize+sisKeySize])
			}
		}
	})

	return res
}
