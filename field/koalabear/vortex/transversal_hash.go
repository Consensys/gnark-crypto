package vortex

import (
	"runtime"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
)

func transversalHash(codewords [][]koalabear.Element, s *sis.RSis) [][sisKeySize]koalabear.Element {
	nbCols := len(codewords[0])

	/*
		v contains a list of rows. We want to hash the columns, in a cache-friendly
		manner.


		for example, if we consider the matrix
		v[0] -> [ 1  2  3  4  ]
		v[1] -> [ 5  6  7  8  ]
		v[2] -> [ 9  10 11 12 ]
		v[3] -> [ 13 14 15 16 ]

		we want to compute
		res = [ H(1,5,9,13) H(2,6,10,14) H(3,7,11,15) H(4,8,12,16) ]

		note that the output size of the hash is s.OutputSize() (i.e it's a slice)
		and that we will decompose the columns in "Limbs" of size s.LogTwoBound;
		this limbs are then interpreted as a slice of coefficients of
		a polynomial of size s.OutputSize()

		that is, we can decompose H(1,5,9,13) as;
		k0 := limbs(1,5) 	= [a b c d e f g h]
		k1 := limbs(9,13) 	= [i j k l m n o p]

		In practice, s.OutputSize() is a reasonable size (< 1024) so we can slide our tiles
		over the partial columns and compute the hash of the columns in parallel.

	*/

	nbBytePerLimb := s.LogTwoBound / 8
	nbLimbsPerField := koalabear.Bytes / nbBytePerLimb
	nbFieldPerPoly := s.Degree / nbLimbsPerField

	// N := s.Degree
	const N = 512
	if N != s.Degree {
		panic("sis key size must be 512")
	}

	nbPolys := divCeil(len(codewords), nbFieldPerPoly)
	res := make([][N]koalabear.Element, nbCols)

	nbRows := len(codewords)
	{
		column := make([]koalabear.Element, nbRows)
		for col := 0; col < nbCols; col++ {

			for r := 0; r < nbRows; r++ {
				column[r] = codewords[r][col]
			}

			s.Hash(column, res[col][:])
		}
	}
	return res

	// First we take care of the constant rows;
	// since they repeat the same value, we can compute them once for the matrix (instead of once per column)
	// and accumulate in res

	// indicates if a block of N rows is constant: in that case we can skip the computation
	// of all the columns sub-hashes in that block.
	// more over; we set the bit of a mask if the row is NOT constant, and exploit the mask
	// to minimize the number of operations we do (partial FFT)
	masks := make([]uint64, nbPolys)

	nbCpus := runtime.NumCPU()

	nbColPerTile := 16
	nbJobs := divCeil(nbCols, nbColPerTile)

	if nbCols < nbCpus {
		nbJobs = nbCols
		nbColPerTile = 1
	}

	for nbJobs < nbCpus && nbColPerTile > 1 {
		nbColPerTile--
		nbJobs = divCeil(nbCols, nbColPerTile)
	}

	executePoolChunky(nbJobs, func(jobID int) {
		startCol := jobID * nbColPerTile
		stopCol := startCol + nbColPerTile
		stopCol = min(stopCol, nbCols)

		// each go routine will iterate over a range of columns; we will hash the columns in parallel
		// and accumulate the result in res (no conflict since each go routine writes to a different range of res)

		itM := newMatrixIterator(codewords, s.LogTwoBound)
		k := make([]koalabear.Element, N)
		kz := make([]koalabear.Element, N)

		for startRow := 0; startRow < len(codewords); startRow += nbFieldPerPoly {
			polID := startRow / nbFieldPerPoly

			// // if it's a constant block, we can skip.
			// if masks[polID] == 0 {
			// 	continue
			// }

			stopRow := startRow + nbFieldPerPoly
			stopRow = min(stopRow, len(codewords))

			// hash the subcolumns.
			for colID := startCol; colID < stopCol; colID++ {
				itM.reset(startRow, stopRow, colID)
				s.InnerHash(itM.lit, res[colID][:], k, kz, polID, masks[polID])
			}

		}

		// mod X^n - 1
		for colID := startCol; colID < stopCol; colID++ {
			s.Domain.FFTInverse(res[colID][:], fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))
		}

	})

	return res
}

// matrixIterator helps allocate resources per go routine
// and iterate over the columns of a matrix (defined by a list of rows: smart-vectors)
type matrixIterator struct {
	it  columnIterator
	lit *sis.LimbIterator
}

func newMatrixIterator(v [][]koalabear.Element, log2bound int) matrixIterator {
	w := matrixIterator{
		it: columnIterator{
			v: v,
		},
	}
	w.lit = sis.NewLimbIterator(&w.it, log2bound/8)
	return w
}

func (w *matrixIterator) reset(startRow, stopRow, colIndex int) {
	w.it.startRow = startRow
	w.it.endRow = stopRow
	w.it.colIndex = colIndex
	w.lit.Reset(&w.it)
}

// columnIterator is a helper struct to iterate over the columns of a matrix
// it implements the sis.ElementIterator interface
type columnIterator struct {
	v                [][]koalabear.Element
	startRow, endRow int
	colIndex         int
}

func (it *columnIterator) Next() (koalabear.Element, bool) {
	if it.endRow == it.startRow {
		return koalabear.Element{}, false
	}
	row := it.v[it.startRow]
	it.startRow++

	return row[it.colIndex], true

}

func divCeil(a, b int) int {
	res := a / b
	if b*res < a {
		return res + 1
	}
	return res
}

var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{}, runtime.GOMAXPROCS(0))
var once sync.Once

func executePoolChunky(nbIterations int, work func(k int)) {
	once.Do(initialize)

	wg := sync.WaitGroup{}
	wg.Add(nbIterations)

	for i := 0; i < nbIterations; i++ {
		k := i
		queue <- func() {
			work(k)
			wg.Done()
			available <- struct{}{}
		}
	}

	wg.Wait()
}

func initialize() {
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		available <- struct{}{}
	}

	go scheduler()
}

func scheduler() {
	for {
		<-available
		task := <-queue
		go task()
	}
}
