package plookup

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
)

var (
	ErrIncompatibleSize = errors.New("the tables in f and t are not of the same size")
	ErrFoldedCommitment = errors.New("the folded commitment is malformed")
)

// ProofLookupTables proofs that a list of tables
type ProofLookupTables struct {

	// commitments to the rows f and t
	fs []kzg.Digest
	ts []kzg.Digest

	// lookup proof for the f and t folded
	foldedProof ProofLookupVector
}

// ProveLookupTables generates a proof that f, seen as a multi dimensional table,
// consists of vectors that are in t. In other words for each i, f[:][i] must be one
// of the t[:][j].
//
// For instance, if t is the truth table of the XOR function, t will be populated such
// that t[:][i] contains the i-th entry of the truth table, so t[0][i] XOR t[1][i] = t[2][i].
//
// The Table in f and t are supposed to be of the same size constant size.
func ProveLookupTables(srs *kzg.SRS, f, t []Table) (ProofLookupTables, error) {

	// res
	proof := ProofLookupTables{}
	var err error

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// check the sizes
	if len(f) != len(t) {
		return proof, ErrIncompatibleSize
	}
	s := len(f[0])
	for i := 1; i < len(f); i++ {
		if len(f[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}
	s = len(t[0])
	for i := 1; i < len(t); i++ {
		if len(t[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}

	// commit to the tables in f and t
	sizeTable := len(t)
	proof.fs = make([]kzg.Digest, sizeTable)
	proof.ts = make([]kzg.Digest, sizeTable)
	m := len(f[0]) + 1
	if m < len(t[0]) {
		m = len(t[0])
	}
	d := fft.NewDomain(uint64(m), 0, false)
	lfs := make([][]fr.Element, sizeTable)
	lts := make([][]fr.Element, sizeTable)
	cfs := make([][]fr.Element, sizeTable)
	cts := make([][]fr.Element, sizeTable)

	for i := 0; i < sizeTable; i++ {

		cfs[i] = make([]fr.Element, d.Cardinality)
		lfs[i] = make([]fr.Element, d.Cardinality)
		copy(cfs[i], f[i])
		copy(lfs[i], f[i])
		for j := len(f[i]); j < int(d.Cardinality); j++ {
			cfs[i][j] = f[i][len(f[i])-1]
			lfs[i][j] = f[i][len(f[i])-1]
		}
		d.FFTInverse(cfs[i], fft.DIF, 0)
		fft.BitReverse(cfs[i])
		proof.fs[i], err = kzg.Commit(cfs[i], srs)
		if err != nil {
			return proof, err
		}

		cts[i] = make([]fr.Element, d.Cardinality)
		lts[i] = make([]fr.Element, d.Cardinality)
		copy(cts[i], t[i])
		copy(lts[i], t[i])
		for j := len(t[i]); j < int(d.Cardinality); j++ {
			cts[i][j] = t[i][len(t[i])-1]
			lts[i][j] = t[i][len(t[i])-1]
		}
		d.FFTInverse(cts[i], fft.DIF, 0)
		fft.BitReverse(cts[i])
		proof.ts[i], err = kzg.Commit(cts[i], srs)
		if err != nil {
			return proof, err
		}
	}

	// fold f and t
	comms := make([]*kzg.Digest, 2*sizeTable)
	for i := 0; i < sizeTable; i++ {
		comms[i] = new(kzg.Digest)
		comms[sizeTable+i] = new(kzg.Digest)
		comms[i].Set(&proof.fs[i])
		comms[sizeTable+i].Set(&proof.ts[i])
	}
	lambda, err := deriveRandomness(&fs, "lambda", comms...)
	if err != nil {
		return proof, err
	}
	// lambda.SetUint64(238293208029)
	lambda.SetString("1535610991669198651944444444444444444444")
	fmt.Printf("lambda (prover):   %s\n", lambda.String())
	foldedf := make(Table, d.Cardinality)
	foldedt := make(Table, d.Cardinality)
	for i := 0; i < len(cfs[0]); i++ {
		for j := sizeTable - 1; j >= 0; j-- {
			foldedf[i].Mul(&foldedf[i], &lambda).
				Add(&foldedf[i], &lfs[j][i])
			foldedt[i].Mul(&foldedt[i], &lambda).
				Add(&foldedt[i], &lts[j][i])
		}
	}

	// call plookupVector, on foldedf[:len(foldedf)-1] to ensure that the domain size
	// in ProveLookupVector is the same as d's
	fmt.Println("folded f")
	for i := 0; i < len(foldedf)-1; i++ {
		fmt.Printf("fvector[%d].SetString(\"%s\")\n", i, foldedf[i].String())
	}
	fmt.Println("")
	fmt.Println("folded t")
	for i := 0; i < len(foldedt); i++ {
		fmt.Printf("lookupVector[%d].SetString(\"%s\")\n", i, foldedt[i].String())
	}
	proof.foldedProof, err = ProveLookupVector(srs, foldedf[:len(foldedf)-1], foldedt)

	return proof, err
}

// VerifyLookupTables verifies that a ProofLookupTables proof is correct.
func VerifyLookupTables(srs *kzg.SRS, proof ProofLookupTables) error {

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// fold the commitments
	sizeTable := len(proof.fs)
	comms := make([]*kzg.Digest, 2*sizeTable)
	for i := 0; i < sizeTable; i++ {
		comms[i] = &proof.fs[i]
		comms[sizeTable+i] = &proof.ts[i]
	}
	lambda, err := deriveRandomness(&fs, "lambda", comms...)
	if err != nil {
		return err
	}
	// lambda.SetUint64(238293208029)
	lambda.SetString("1535610991669198651944444444444444444444")
	fmt.Printf("lambda (verifier): %s\n", lambda.String())

	// verify that the commitments in the inner proof are consistant
	// with the folded commitments.
	var comt, comf kzg.Digest
	comf.Set(&proof.fs[sizeTable-1])
	comt.Set(&proof.ts[sizeTable-1])
	var blambda big.Int
	lambda.ToBigIntRegular(&blambda)
	for i := sizeTable - 2; i >= 0; i-- {
		comf.ScalarMultiplication(&comf, &blambda).
			Add(&comf, &proof.fs[i])
		comt.ScalarMultiplication(&comt, &blambda).
			Add(&comt, &proof.ts[i])
	}

	if !comf.Equal(&proof.foldedProof.f) {
		return ErrFoldedCommitment
	}
	if !comt.Equal(&proof.foldedProof.t) {
		return ErrFoldedCommitment
	}

	// verify the inner proof
	return VerifyLookupVector(srs, proof.foldedProof)
}

// TODO put that in fiat-shamir package
func deriveRandomness(fs *fiatshamir.Transcript, challenge string, points ...*bn254.G1Affine) (fr.Element, error) {

	var buf [bn254.SizeOfG1AffineUncompressed]byte
	var r fr.Element

	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}

	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}
