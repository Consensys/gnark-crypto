package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// transversalHash hashes the columns of codewords, using SIS by default, unless ots (="other than sis") is not nil.
func transversalHash(codewords []koalabear.Element, s *sis.RSis, sizeCodeWord int, ots NewHash) []koalabear.Element {
	if ots != nil {
		return transveralHashGeneric(codewords, ots, sizeCodeWord)
	} else {
		return transversalHashSIS(codewords, s, sizeCodeWord)
	}
}

// transveralHashGeneric hashes the columns of the codewords in parallel
// using the provided hash function, whose sum is on 32bytes. The result is a slice that should be read
// 8 elements at a time, which makes 32 bytes, the i-th batch of 8 koalbear elements is the hash of the i-th column.
func transveralHashGeneric(codewords []koalabear.Element, newHash NewHash, sizeCodeWord int) []koalabear.Element {

	const nbKoalbearElementsPerHash = 8

	nbCols := sizeCodeWord
	nbRows := len(codewords) / sizeCodeWord

	// the result in that case consists of concatenated blocks of 32 bytes, interpreted as 8 consecutive koalabear elements.
	res := make([]koalabear.Element, nbCols*nbKoalbearElementsPerHash)

	parallel.Execute(nbCols, func(start, end int) {
		h := newHash()
		for i := start; i < end; i++ {
			for j := 0; j < nbRows; j++ {
				curElmt := codewords[j*nbCols+i]
				h.Write(curElmt.Marshal())
			}
			curHash := h.Sum(nil)
			s := i * nbKoalbearElementsPerHash
			byteSize := koalabear.Bytes
			for j := 0; j < nbKoalbearElementsPerHash; j++ {
				res[s+j].SetBytes(curHash[j*byteSize : (j+1)*byteSize])
			}
		}
	})
	return res
}

// transversalHashSIS hashes the columns of the codewords in parallel
// using the SIS hash function.
func transversalHashSIS(codewords []koalabear.Element, s *sis.RSis, sizeCodeWord int) []koalabear.Element {

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
