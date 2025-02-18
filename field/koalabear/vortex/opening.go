package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
)

// Proof is an opening proof
type Proof struct {
	// UAlpha is the random linear combination of the encoded rows
	// of the committed matrix. We use [4]koalabear.Element because
	// we don't have a field extension implementation.
	UAlpha [][4]koalabear.Element
	// OpenedColumns is the list of columns that have been opened
	OpenedColumns [][]koalabear.Element
}

// OpenLinComb performs the "UAlpha" part of the proof computation.
func (p *Params) OpenLinComb(encodedMatrix [][]koalabear.Element, alpha [4]koalabear.Element) (Proof, error) {

	for i := range encodedMatrix {
		if len(encodedMatrix[i]) != p.NbEncodedColumns() {
			return Proof{}, fmt.Errorf("expected %d encoded columns, got %d", p.NbEncodedColumns(), len(encodedMatrix[i]))
		}
	}

	ualpha := make([][4]koalabear.Element, p.NbEncodedColumns())
	for row := len(encodedMatrix) - 1; row >= 0; row-- {
		for col := 0; col < p.NbEncodedColumns(); col++ {

			// Note, we do 4 field multiplication but really we should be doing a field
			// extension operation. Here we are computing the linear combination of the
			// rows by successive powers of alpha "coordinate-by-coordinate" but we should
			// be doing the "extension" multiplication.
			ualpha[col][0].Mul(&alpha[0], &ualpha[col][0])
			ualpha[col][1].Mul(&alpha[1], &ualpha[col][1])
			ualpha[col][2].Mul(&alpha[2], &ualpha[col][2])
			ualpha[col][3].Mul(&alpha[3], &ualpha[col][3])

			ualpha[col][0].Add(&ualpha[col][0], &encodedMatrix[row][col])
			ualpha[col][1].Add(&ualpha[col][1], &encodedMatrix[row][col])
			ualpha[col][2].Add(&ualpha[col][2], &encodedMatrix[row][col])
			ualpha[col][3].Add(&ualpha[col][3], &encodedMatrix[row][col])
		}
	}

	return Proof{UAlpha: ualpha}, nil
}

// OpenColumn performs the "OpenedColumns" part of the proof computation.
func (p *Proof) OpenColumn(codewords [][]koalabear.Element, selectedColumns []int) error {

	for _, col := range selectedColumns {
		openedColumn := getTransposedColumn(codewords, col)
		p.OpenedColumns = append(p.OpenedColumns, openedColumn)
	}

	return nil
}
