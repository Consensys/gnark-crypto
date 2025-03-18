package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

func transversalHash(codewords [][]koalabear.Element, s *sis.RSis) [][sisKeySize]koalabear.Element {
	nbCols := len(codewords[0])

	// N := s.Degree
	const N = 512
	if N != s.Degree {
		panic("sis key size must be 512")
	}

	res := make([][N]koalabear.Element, nbCols)

	parallel.Execute(nbCols, func(start, end int) {
		column := make([]koalabear.Element, len(codewords))
		for col := start; col < end; col++ {
			for r := 0; r < len(codewords); r++ {
				column[r] = codewords[r][col]
			}
			s.Hash(column[:], res[col][:])
		}
	})

	// parallel.Execute(nbCols, func(start, end int) {
	// 	const blockSize = 2
	// 	var column [blockSize][]koalabear.Element
	// 	for i := range column {
	// 		column[i] = make([]koalabear.Element, len(codewords))
	// 	}

	// 	for col := start; col < end-blockSize; col += blockSize {
	// 		for r := 0; r < len(codewords); r++ {
	// 			for i := 0; i < blockSize; i++ {
	// 				column[i][r] = codewords[r][col+i]
	// 			}
	// 		}
	// 		for i := 0; i < blockSize; i++ {
	// 			s.Hash(column[i], res[col+i][:])
	// 		}
	// 	}
	// 	for col := end - 4; col < end && col >= 0; col++ {
	// 		for r := 0; r < len(codewords); r++ {
	// 			column[0][r] = codewords[r][col]
	// 		}
	// 		s.Hash(column[0], res[col][:])
	// 	}
	// })

	return res
}
