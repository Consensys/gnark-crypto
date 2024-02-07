package kzg

import (
	"fmt"
	"hash"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/internal/parallel"
	iciclegnark "github.com/ingonyama-zk/iciclegnark/curves/bn254"
)

// Commit commits to a polynomial using a multi exponentiation with the SRS.
// It is assumed that the polynomial is in canonical form, in Montgomery form.
func OnDeviceCommit(p []fr.Element, G1 unsafe.Pointer, nbTasks ...int) (Digest, error) {
	// Size of the polynomial
	np := len(p)

	// Size of the polynomial in bytes
	sizeBytesScalars := np * fr.Bytes

	// Initialize Scalar channels
	copyCpDone := make(chan unsafe.Pointer, 1)
	cpDeviceData := make(chan iciclegnark.OnDeviceData, 1)

	// Copy Scalar to device
	go func() {
		// Perform copy operation
		iciclegnark.CopyToDevice(p, sizeBytesScalars, copyCpDone)

		// Receive result once copy operation is done
		cpDevice := <-copyCpDone

		// Create OnDeviceData
		cpDeviceValue := iciclegnark.OnDeviceData{
			P:    cpDevice,
			Size: sizeBytesScalars,
		}

		// Send OnDeviceData to respective channel
		cpDeviceData <- cpDeviceValue

		// Close channels
		close(copyCpDone)
		close(cpDeviceData)
	}()

	// Wait for copy operation to finish
	cpDeviceValue := <-cpDeviceData

	// KZG Committment on device
	var wg sync.WaitGroup

	// Perform multi exponentiation on device
	wg.Add(1)
	tmpChan := make(chan bn254.G1Affine, 1)
	go func() {
		defer wg.Done()
		tmp, _, err := iciclegnark.MsmOnDevice(cpDeviceValue.P, G1, np, true)
		//fmt.Println("tmp", tmp)
		if err != nil {
			fmt.Print("error", err)
		}
		var res bn254.G1Affine
		res.FromJacobian(&tmp)
		tmpChan <- res
	}()
	wg.Wait()

	// Receive result once copy operation is done
	res := <-tmpChan

	// Free device memory
	go func() {
		iciclegnark.FreeDevicePointer(unsafe.Pointer(&cpDeviceValue))
	}()

	return res, nil
}

// Open computes an opening proof of polynomial p at given point.
// fft.Domain Cardinality must be larger than p.Degree()
func OnDeviceOpen(p []fr.Element, point fr.Element, G1 unsafe.Pointer) (OpeningProof, error) {
	// build the proof
	res := OpeningProof{
		ClaimedValue: eval(p, point),
	}

	// compute H
	// h reuses memory from _p
	_p := make([]fr.Element, len(p))
	copy(_p, p)
	h := dividePolyByXminusA(_p, res.ClaimedValue, point)

	// commit to H
	hCommit, err := OnDeviceCommit(h, G1)
	if err != nil {
		return OpeningProof{}, err
	}
	res.H.Set(&hCommit)

	return res, nil
}

// BatchOpenSinglePoint creates a batch opening proof at point of a list of polynomials.
// It's an interactive protocol, made non-interactive using Fiat Shamir.
//
// * point is the point at which the polynomials are opened.
// * digests is the list of committed polynomials to open, need to derive the challenge using Fiat Shamir.
// * polynomials is the list of polynomials to open, they are supposed to be of the same size.
// * dataTranscript extra data that might be needed to derive the challenge used for folding
func OnDeviceBatchOpenSinglePoint(polynomials [][]fr.Element, digests []Digest, point fr.Element, hf hash.Hash, pk ProvingKey, G1 unsafe.Pointer, dataTranscript ...[]byte) (BatchOpeningProof, error) {

	// check for invalid sizes
	nbDigests := len(digests)
	if nbDigests != len(polynomials) {
		return BatchOpeningProof{}, ErrInvalidNbDigests
	}

	// TODO ensure the polynomials are of the same size
	largestPoly := -1
	for _, p := range polynomials {
		if len(p) == 0 || len(p) > len(pk.G1) {
			return BatchOpeningProof{}, ErrInvalidPolynomialSize
		}
		if len(p) > largestPoly {
			largestPoly = len(p)
		}
	}

	var res BatchOpeningProof

	// compute the purported values
	res.ClaimedValues = make([]fr.Element, len(polynomials))
	var wg sync.WaitGroup
	wg.Add(len(polynomials))
	for i := 0; i < len(polynomials); i++ {
		go func(_i int) {
			res.ClaimedValues[_i] = eval(polynomials[_i], point)
			wg.Done()
		}(i)
	}

	// wait for polynomial evaluations to be completed (res.ClaimedValues)
	wg.Wait()

	// derive the challenge γ, binded to the point and the commitments
	gamma, err := deriveGamma(point, digests, res.ClaimedValues, hf, dataTranscript...)
	if err != nil {
		return BatchOpeningProof{}, err
	}

	// ∑ᵢγⁱf(a)
	var foldedEvaluations fr.Element
	chSumGammai := make(chan struct{}, 1)
	go func() {
		foldedEvaluations = res.ClaimedValues[nbDigests-1]
		for i := nbDigests - 2; i >= 0; i-- {
			foldedEvaluations.Mul(&foldedEvaluations, &gamma).
				Add(&foldedEvaluations, &res.ClaimedValues[i])
		}
		close(chSumGammai)
	}()

	// compute ∑ᵢγⁱfᵢ
	// note: if we are willing to parallelize that, we could clone the poly and scale them by
	// gamma n in parallel, before reducing into foldedPolynomials
	foldedPolynomials := make([]fr.Element, largestPoly)
	copy(foldedPolynomials, polynomials[0])
	gammas := make([]fr.Element, len(polynomials))
	gammas[0] = gamma
	for i := 1; i < len(polynomials); i++ {
		gammas[i].Mul(&gammas[i-1], &gamma)
	}

	for i := 1; i < len(polynomials); i++ {
		i := i
		parallel.Execute(len(polynomials[i]), func(start, end int) {
			var pj fr.Element
			for j := start; j < end; j++ {
				pj.Mul(&polynomials[i][j], &gammas[i-1])
				foldedPolynomials[j].Add(&foldedPolynomials[j], &pj)
			}
		})
	}

	// compute H
	<-chSumGammai
	h := dividePolyByXminusA(foldedPolynomials, foldedEvaluations, point)
	foldedPolynomials = nil // same memory as h

	res.H, err = OnDeviceCommit(h, G1)
	if err != nil {
		return BatchOpeningProof{}, err
	}

	return res, nil
}
