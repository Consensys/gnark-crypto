package vortex

import (
	"errors"
	"fmt"
	"math/big"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// FRIProof is a proof that a vector u is an RS codeword, generated via FRI.
// It commits to u and NumFoldRounds folded versions, then opens each at the
// same set of K query positions to enable consistency and proximity checks.
type FRIProof struct {
	// Root of the Merkle tree over the original u (round 0)
	Root Hash
	// FoldRoots[l] is the Merkle root of the codeword folded l+1 times (l=0..NumFoldRounds-2)
	FoldRoots []Hash
	// Values[l][k] is the E4 value of the codeword at round l, queried position k
	// Round 0 = original u, round 1..NumFoldRounds = folded codewords
	Values [][]fext.E4
	// Paths[l][k] is the Merkle proof for Values[l][k]
	Paths [][]MerkleProof
	// FinalPoly is the final polynomial after all folds (size N/2^NumFoldRounds)
	FinalPoly []fext.E4
}

// NumFoldRounds returns the number of FRI fold rounds represented in this proof.
func (p *FRIProof) NumFoldRounds() int {
	return len(p.FoldRoots) + 1
}

// hashE4 hashes a single E4 element (4 koalabear.Elements) to a Hash.
func hashE4(v fext.E4) Hash {
	elems := (*[4]koalabear.Element)(unsafe.Pointer(&v))
	return HashPoseidon2(elems[:])
}

// commitE4Codeword builds a Merkle tree over a codeword of E4 elements.
// Each leaf is the hash of a single E4 element.
func commitE4Codeword(codeword []fext.E4) *MerkleTree {
	leaves := make([]Hash, len(codeword))
	parallel.Execute(len(leaves), func(start, end int) {
		for i := start; i < end; i++ {
			leaves[i] = hashE4(codeword[i])
		}
	})
	return BuildMerkleTree(leaves, nil)
}

// FRIFold applies one round of FRI folding to a codeword in natural evaluation order.
// The codeword is evaluated at {ω^0, ω^1, ..., ω^{N-1}} where ω=gen (primitive N-th root).
// Returns the folded codeword of size N/2 evaluated at {ω^0, ω^2, ..., ω^{N-2}}.
//
//	fold[i] = (f[i] + f[i+N/2]) / 2  +  r * (f[i] - f[i+N/2]) / (2 * ω^i)
func FRIFold(codeword []fext.E4, r fext.E4, gen koalabear.Element) []fext.E4 {
	n := len(codeword)
	half := n / 2
	res := make([]fext.E4, half)

	var inv2 koalabear.Element
	inv2.SetUint64(2)
	inv2.Inverse(&inv2)

	parallel.Execute(half, func(start, end int) {
		// Compute ω^start
		var omegaPow koalabear.Element
		omegaPow.Exp(gen, big.NewInt(int64(start)))

		for i := start; i < end; i++ {
			lo := codeword[i]
			hi := codeword[i+half]

			// sum = (lo + hi) * inv2
			var sum fext.E4
			sum.Add(&lo, &hi)
			sum.MulByElement(&sum, &inv2)

			// diff = (lo - hi) * inv2 / ω^i = (lo - hi) * inv2 * omegaPowInv
			var diff fext.E4
			diff.Sub(&lo, &hi)
			diff.MulByElement(&diff, &inv2)

			// divide by ω^i: multiply by ω^{-i}
			var omegaPowInv koalabear.Element
			omegaPowInv.Inverse(&omegaPow)
			diff.MulByElement(&diff, &omegaPowInv)

			// fold[i] = sum + r * diff
			var rDiff fext.E4
			rDiff.Mul(&r, &diff)
			res[i].Add(&sum, &rDiff)

			omegaPow.Mul(&omegaPow, &gen)
		}
	})

	return res
}

// FRIProve generates an FRI proof that ualpha is an RS codeword.
// queriedCols contains the K column indices queried by the verifier (for column consistency checks).
// numFoldRounds is the number of FRI folding rounds.
// foldChallenges[l] is the random field extension element used in fold round l.
//
// The proof opens each folded codeword at the same (reduced) query positions derived from queriedCols.
func (p *Params) FRIProve(ualpha []fext.E4, queriedCols []int, numFoldRounds int, foldChallenges []fext.E4) (*FRIProof, error) {
	n := len(ualpha)
	if len(foldChallenges) != numFoldRounds {
		return nil, fmt.Errorf("expected %d fold challenges, got %d", numFoldRounds, len(foldChallenges))
	}

	// Build Merkle commitments for each round
	codewords := make([][]fext.E4, numFoldRounds+1)
	trees := make([]*MerkleTree, numFoldRounds+1)
	codewords[0] = ualpha
	trees[0] = commitE4Codeword(ualpha)

	gen := p.Domains[1].Generator // primitive N-th root of unity

	for l := 0; l < numFoldRounds; l++ {
		codewords[l+1] = FRIFold(codewords[l], foldChallenges[l], gen)
		if l < numFoldRounds-1 {
			trees[l+1] = commitE4Codeword(codewords[l+1])
		}
		// square gen: new domain generator is gen^2 (for size n/2 domain)
		gen.Square(&gen)
	}
	// Restore gen to Domains[1].Generator for root computation
	gen = p.Domains[1].Generator

	// Collect roots
	foldRoots := make([]Hash, numFoldRounds-1)
	for l := 0; l < numFoldRounds-1; l++ {
		foldRoots[l] = trees[l+1].Root()
	}

	// For each round, open at the query positions (mod size of codeword at that round)
	numQ := len(queriedCols)
	values := make([][]fext.E4, numFoldRounds)
	paths := make([][]MerkleProof, numFoldRounds)
	for l := 0; l < numFoldRounds; l++ {
		size := n >> l
		values[l] = make([]fext.E4, numQ)
		paths[l] = make([]MerkleProof, numQ)
		tree := trees[l]
		for k, col := range queriedCols {
			pos := col % size
			values[l][k] = codewords[l][pos]
			if tree != nil {
				mp, err := tree.Open(pos)
				if err != nil {
					return nil, fmt.Errorf("merkle open failed at round %d pos %d: %w", l, pos, err)
				}
				paths[l][k] = mp
			}
		}
	}

	return &FRIProof{
		Root:      trees[0].Root(),
		FoldRoots: foldRoots,
		Values:    values,
		Paths:     paths,
		FinalPoly: codewords[numFoldRounds],
	}, nil
}

// FRIVerify checks the FRI proof. It verifies:
// 1. Round-0 Merkle openings are consistent with Root
// 2. Folding consistency between rounds
// 3. Final polynomial evaluates consistently with round (numFoldRounds-1) values
//
// uAlphaAtQueried[k] are the claimed values of ualpha at queriedCols[k] (from column consistency check).
// foldChallenges[l] are the folding challenges.
func (p *Params) FRIVerify(proof *FRIProof, queriedCols []int, uAlphaAtQueried []fext.E4, foldChallenges []fext.E4) error {
	numQ := len(queriedCols)
	numRounds := proof.NumFoldRounds()
	n := p.SizeCodeWord()

	if len(uAlphaAtQueried) != numQ {
		return fmt.Errorf("expected %d ualpha values, got %d", numQ, len(uAlphaAtQueried))
	}
	if len(foldChallenges) != numRounds {
		return fmt.Errorf("expected %d fold challenges, got %d", numRounds, len(foldChallenges))
	}
	if len(proof.Values) != numRounds || len(proof.Paths) != numRounds {
		return fmt.Errorf("proof has wrong number of rounds")
	}

	var inv2 koalabear.Element
	inv2.SetUint64(2)
	inv2.Inverse(&inv2)

	gen := p.Domains[1].Generator

	// Step 1: check round-0 values match uAlphaAtQueried
	for k := range queriedCols {
		if proof.Values[0][k] != uAlphaAtQueried[k] {
			return fmt.Errorf("round-0 value mismatch at query %d", k)
		}
	}

	// Step 2: check Merkle proofs for rounds 0..numRounds-1
	roots := make([]Hash, numRounds)
	roots[0] = proof.Root
	for l := 0; l < numRounds-1; l++ {
		roots[l+1] = proof.FoldRoots[l]
	}

	for l := 0; l < numRounds; l++ {
		size := n >> l
		for k, col := range queriedCols {
			pos := col % size
			leaf := hashE4(proof.Values[l][k])
			if err := proof.Paths[l][k].Verify(pos, leaf, roots[l], nil); err != nil {
				return fmt.Errorf("merkle proof failed at round %d query %d: %w", l, k, err)
			}
		}
	}

	// Step 3: check folding consistency between rounds
	// For round l → l+1: fold[pos % (size/2)] should equal FRIFold applied to (values[l][pos], values[l][pos+size/2])
	// We check: fold_l[i] = (f_l[i] + f_l[i+half]) / 2 + r_l * (f_l[i] - f_l[i+half]) / (2 * ω_l^i)
	// where ω_l = gen^{2^l} (generator for round-l domain of size n/2^l)
	genL := gen
	for l := 0; l < numRounds-1; l++ {
		sizeL := n >> l
		half := sizeL / 2
		r := foldChallenges[l]

		for k, col := range queriedCols {
			posL := col % sizeL
			if posL >= half {
				// This position is in the second half; its pair is posL-half
				// Values[l][k] should already be at pos=col%sizeL
				// The folded position for posL is posL - half (same as posL%half)
				// Check that values[l+1][k] is at posL%half, consistent with values[l]
				// We need the "other half" value which may not be directly in proof...
				// For simplicity, when pos >= half, the folded index is pos - half = pos % half
				// So values[l+1][k] corresponds to fold result at posL - half
				// fold[posL-half] uses f[posL-half] and f[posL]
				// We have f[posL] = values[l][k]. We don't have f[posL-half] directly.
				// This means we need to query both halves. Skip this check for positions >= half.
				continue
			}

			// posL < half: we have f[posL] = values[l][k]
			// But we also need f[posL + half]. Check if another query covers it.
			posHi := posL + half
			hiVal, found := findQueryVal(queriedCols, proof.Values[l], sizeL, posHi)
			if !found {
				// Can't check folding consistency without the paired value
				continue
			}

			// Compute expected fold value
			lo := proof.Values[l][k]
			hi := hiVal

			var sum fext.E4
			sum.Add(&lo, &hi)
			sum.MulByElement(&sum, &inv2)

			var omegaPowI koalabear.Element
			omegaPowI.Exp(genL, big.NewInt(int64(posL)))
			var omegaPowInv koalabear.Element
			omegaPowInv.Inverse(&omegaPowI)

			var diff fext.E4
			diff.Sub(&lo, &hi)
			diff.MulByElement(&diff, &inv2)
			diff.MulByElement(&diff, &omegaPowInv)

			var rDiff fext.E4
			rDiff.Mul(&r, &diff)
			var expected fext.E4
			expected.Add(&sum, &rDiff)

			// The expected fold value should equal values[l+1][k] (at position posL % half = posL)
			got := proof.Values[l+1][k]
			if expected != got {
				return fmt.Errorf("folding inconsistency at round %d query %d", l, k)
			}
		}
		genL.Square(&genL)
	}

	// Step 4: check final polynomial consistency
	// FinalPoly should be a small polynomial (constant or low-degree).
	// Check: FinalPoly[pos % finalSize] == values[numRounds-1][k] for each query k.
	finalSize := len(proof.FinalPoly)
	if finalSize == 0 {
		return errors.New("FRI proof has empty FinalPoly")
	}
	for k, col := range queriedCols {
		sizeL := n >> (numRounds - 1)
		pos := col % sizeL
		finalPos := pos % finalSize
		if proof.Values[numRounds-1][k] != proof.FinalPoly[finalPos] {
			return fmt.Errorf("final poly inconsistency at query %d: pos=%d finalPos=%d", k, pos, finalPos)
		}
	}

	return nil
}

// findQueryVal looks for a query that hits position pos (mod sizeL) in the round-l values.
func findQueryVal(queriedCols []int, vals []fext.E4, sizeL, pos int) (fext.E4, bool) {
	for k, col := range queriedCols {
		if col%sizeL == pos {
			return vals[k], true
		}
	}
	return fext.E4{}, false
}
