package vortex

import (
	"fmt"
	"math/big"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/internal/parallel"
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
	EncodedMatrix []koalabear.Element
	// SisHashes are the SIS hashes of the encoded matrix
	SisHashes []koalabear.Element
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
	N := p.SizeCodeWord()
	const blockSize = 16
	if N%blockSize != 0 {
		panic("len of codewords must be a multiple of 16")
	}
	// if p.Key.Degree != sisKeySize {
	// 	panic("sis key size must be 512")
	// }

	codewords := make([]koalabear.Element, len(input)*N)
	merkleLeaves := make([]Hash, N)

	parallel.Execute(len(input), func(start, end int) {
		for i := start; i < end; i++ {
			p.EncodeReedSolomon(input[i], codewords[i*N:i*N+N])
		}
	})

	sisHashes := transversalHash(codewords, p.Key, p.SizeCodeWord())

	parallel.Execute(N/blockSize, func(start, end int) {
		sisKeySize := p.Key.Degree
		for block := start; block < end; block++ {
			b := block * blockSize
			_start := b * sisKeySize
			_end := _start + sisKeySize*blockSize
			HashPoseidon2x16(sisHashes[_start:_end], merkleLeaves[b:b+blockSize])
		}
	})

	return &ProverState{
		Params:        p,
		EncodedMatrix: codewords,
		SisHashes:     sisHashes,
		MerkleTree:    BuildMerkleTree(merkleLeaves),
	}, nil
}

// OpenLinComb performs the "UAlpha" part of the proof computation.
// UAlpha is computed as uAlpha := \sum_i row_i * alpha^i.
func (ps *ProverState) OpenLinComb(alpha fext.E4) {

	codewords := ps.EncodedMatrix

	// We don't use the Horner algorithm because we can save on fext
	// operations using the naive algorithm.
	N := ps.Params.SizeCodeWord()
	nbCodewords := len(codewords) / N
	_ualpha := make([]fext.E4, ps.Params.SizeCodeWord())
	var lock sync.Mutex
	parallel.Execute(nbCodewords, func(start, end int) {
		ualpha := make([]fext.E4, ps.Params.SizeCodeWord())
		alphaPow := new(fext.E4).SetOne()
		alphaPow.Exp(alpha, big.NewInt(int64(start)))
		for i := start; i < end; i++ {
			fext.MulAccE4(alphaPow, codewords[i*N:i*N+N], ualpha)
			alphaPow.Mul(alphaPow, &alpha)
		}

		// using unsafe, we take the address of _ualpha[0] and
		// create a vector of fr.Element of size M starting at _ualpha[0]
		M := len(ualpha) * 4
		vUalpha := koalabear.Vector(unsafe.Slice((*koalabear.Element)(unsafe.Pointer(&ualpha[0])), M))
		_vUalpha := koalabear.Vector(unsafe.Slice((*koalabear.Element)(unsafe.Pointer(&_ualpha[0])), M))

		lock.Lock()
		_vUalpha.Add(_vUalpha, vUalpha)
		lock.Unlock()
	})

	ps.Ualpha = _ualpha
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
		if col >= ps.Params.SizeCodeWord() {
			return nil, fmt.Errorf("column index out of range")
		}
		openedColumns[i] = getTransposedColumn(encodedMatrix, col, ps.Params.SizeCodeWord())
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
func getTransposedColumn(codewords []koalabear.Element, col int, sizeCodeWord int) []koalabear.Element {
	// Create a buffer to store the column elements
	colBuffer := make([]koalabear.Element, len(codewords)/sizeCodeWord)

	// Iterate over each row and extract the element from the specified column
	for row := range colBuffer {
		colBuffer[row] = codewords[row*sizeCodeWord+col]
	}

	// Return the extracted column as a slice
	return colBuffer
}
