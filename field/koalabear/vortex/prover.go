package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// Proof is an opening proof
type Proof struct {
	// UAlpha is the random linear combination of the encoded rows
	// of the committed matrix.
	UAlpha []fext.E4
	// OpenedColumns is the list of columns that have been opened
	OpenedColumns [][]koalabear.Element
	// MerkleProof is the list of the Merkle-Proofs for the opened columns
	MerkleProofOpenedColumns []MerkleProof
}

const sisKeySize = 512

// ProverState stores the state of the prover in the Vortex protocol
// and tracks the internal values.
type ProverState struct {
	// Params are the parameters provided to the prover to commit.
	Params *Params
	// EncodedMatrix is computed by the prover during the commitment
	// time.
	EncodedMatrix [][]koalabear.Element
	// SisHashes are the SIS hashes of the encoded matrix
	SisHashes [][sisKeySize]koalabear.Element
	// MerkleTree is the Merkle tree of the SIS hashes
	MerkleTree *MerkleTree
	// Ualpha is the linear combination of the rows of the encoded matrix
	Ualpha []fext.E4
}

// GetCommitment returns the short commitment to the input matrix
func (ps *ProverState) GetCommitment() Hash {
	return ps.MerkleTree.Levels[0][0]
}

// CommitSis returns the commitment to the input matrix. The
// matrix is provided row-by-row in the input.
func Commit(p *Params, input [][]koalabear.Element) (*ProverState, error) {

	var (
		codewords = make([][]koalabear.Element, len(input))
		err       error
	)

	for i := range input {
		if codewords[i], err = p.EncodeReedSolomon(input[i]); err != nil {
			return nil, fmt.Errorf("error in reed-solomon encode: %w", err)
		}
	}

	const (
		blockSize = 16
	)
	if len(codewords[0])%blockSize != 0 {
		panic("len of codewords must be a multiple of 16")
	}
	if p.Key.Degree != sisKeySize {
		panic("sis key size must be 512")
	}

	var (
		sisHashes    = make([][sisKeySize]koalabear.Element, len(codewords[0]))
		merkleLeaves = make([]Hash, len(codewords[0]))
	)
	var colBuffer [blockSize][]koalabear.Element
	for i := range colBuffer {
		colBuffer[i] = make([]koalabear.Element, len(input))
	}

	for col := 0; col < len(codewords[0]); col += blockSize {

		// transpose blockSize columns at a time.
		transposeCodewords(codewords, col, blockSize, colBuffer)

		for i := range colBuffer {
			_ = p.Key.Hash(colBuffer[i], sisHashes[col+i][:])
		}
		HashPoseidon2x16(sisHashes[col:col+blockSize], merkleLeaves[col:col+blockSize])
	}

	return &ProverState{
		Params:        p,
		EncodedMatrix: codewords,
		SisHashes:     sisHashes,
		MerkleTree:    BuildMerkleTree(merkleLeaves),
	}, nil
}

func transposeCodewords(codewords [][]koalabear.Element, col, blockSize int, colBuffer [16][]koalabear.Element) {
	n := len(colBuffer[0])
	for i := 0; i < blockSize; i++ {
		for row := 0; row < n; row++ {
			colBuffer[i][row] = codewords[row][col+i]
		}
	}
}

// OpenLinComb performs the "UAlpha" part of the proof computation.
// UAlpha is computed as uAlpha := \sum_i row_i * alpha^i.
func (ps *ProverState) OpenLinComb(alpha fext.E4) {

	var (
		ualpha        = make([]fext.E4, ps.Params.NbEncodedColumns())
		tmp           = fext.E4{}
		alphaPow      = new(fext.E4).SetOne()
		encodedMatrix = ps.EncodedMatrix
	)

	// We don't use the Horner algorithm because we can save on fext
	// operations using the naive algorithm.
	for row := 0; row < len(encodedMatrix); row++ {
		for col := 0; col < ps.Params.NbEncodedColumns(); col++ {
			tmp.MulByElement(alphaPow, &encodedMatrix[row][col])
			ualpha[col].Add(&ualpha[col], &tmp)
		}

		alphaPow.Mul(alphaPow, &alpha)
	}

	ps.Ualpha = ualpha
}

// OpenColumns sets the OpenedColumns field of the proof using the provided
// codewords and selected columns.
func (ps *ProverState) OpenColumns(selectedColumns []int) (*Proof, error) {

	var (
		numSelectedColumns       = len(selectedColumns)
		openedColumns            = make([][]koalabear.Element, numSelectedColumns)
		merkleProofOpenedColumns = make([]MerkleProof, numSelectedColumns)
		encodedMatrix            = ps.EncodedMatrix
		err                      error
	)

	for i, col := range selectedColumns {

		// an error here indicates that the user samples integers that are
		// too large.
		if col >= ps.Params.NbEncodedColumns() {
			return nil, fmt.Errorf("column index out of range")
		}

		openedColumns[i] = getTransposedColumn(encodedMatrix, col)
		if merkleProofOpenedColumns[i], err = ps.MerkleTree.Open(col); err != nil {
			return nil, fmt.Errorf("error in merkle proof generation: %w", err)
		}
	}

	return &Proof{
		UAlpha:                   ps.Ualpha,
		OpenedColumns:            openedColumns,
		MerkleProofOpenedColumns: merkleProofOpenedColumns,
	}, nil
}

// getTransposedColumn returns the specified column from the codewords matrix.
// It extracts the column at index 'col' from a 2D slice of koalabear.Elements.
func getTransposedColumn(codewords [][]koalabear.Element, col int) []koalabear.Element {
	// Create a buffer to store the column elements
	colBuffer := make([]koalabear.Element, len(codewords))

	// Iterate over each row and extract the element from the specified column
	for row := range colBuffer {
		colBuffer[row] = codewords[row][col]
	}

	// Return the extracted column as a slice
	return colBuffer
}
