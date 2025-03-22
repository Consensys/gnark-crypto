package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

func transversalHash(codewords []koalabear.Element, s *sis.RSis, sizeCodeWord int) []koalabear.Element {
	N := s.Degree

	nbCols := sizeCodeWord
	nbRows := len(codewords) / sizeCodeWord

	res := make([]koalabear.Element, nbCols*N)

	parallel.Execute(nbCols, func(start, end int) {
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
				s.Hash(transposed[j], res[(col+j)*N:(col+j)*N+N])
			}
		}
	})

	return res
}
