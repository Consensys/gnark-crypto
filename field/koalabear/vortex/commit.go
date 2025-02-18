package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
)

// CommitSis returns the commitment to the input matrix. The
// matrix is provided row-by-row in the input.
func (p *Params) CommitSis(input [][]koalabear.Element) ([][]koalabear.Element, error) {

	var (
		codewords = make([][]koalabear.Element, len(input))
		err       error
		colBuffer = make([]koalabear.Element, len(input))
		colHashes = make([][]koalabear.Element, len(input))
	)

	for i := range input {
		if codewords[i], err = p.EncodeReedSolomon(input[i]); err != nil {
			return nil, fmt.Errorf("error in reed-solomon encode: %w", err)
		}
	}

	for col := 0; col < len(codewords[0]); col++ {

		// Transpose the values of the
		for row := range colBuffer {
			colBuffer[row] = codewords[row][col]
		}

		colHashes[col] = make([]koalabear.Element, p.Key.Degree)
		if err := p.Key.Hash(colBuffer, colHashes[col]); err != nil {
			return nil, fmt.Errorf("error in sis hash: %w", err)
		}
	}

	return colHashes, nil
}

func getTransposedColumn(codewords [][]koalabear.Element, col int) []koalabear.Element {

	var (
		colBuffer = make([]koalabear.Element, len(codewords))
	)

	for row := range colBuffer {
		colBuffer[row] = codewords[row][col]
	}

	return colBuffer
}
