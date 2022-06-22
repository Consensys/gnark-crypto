package polynomial

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// MultiLin tracks the values of a (dense i.e. not sparse) multilinear polynomial
// The variables are X₁ through Xₙ where n = log(len(.))
// .[∑ᵢ 2ⁱ⁻¹ bₙ₋ᵢ] = the polynomial evaluated at (b₁, b₂, ..., bₙ)
// It is understood that any hypercube evaluation can be extrapolated to a multilinear polynomial
type MultiLin []fr.Element

// Fold is partial evaluation function k[X₁, X₂, ..., Xₙ] → k[X₂, ..., Xₙ] by setting X₁=r
func (m *MultiLin) Fold(r fr.Element) {
	mid := len(*m) / 2

	bottom, top := (*m)[:mid], (*m)[mid:]

	// updating bookkeeping table
	// knowing that the polynomial f ∈ (k[X₂, ..., Xₙ])[X₁] is linear, we would get f(r) = f(0) + r(f(1) - f(0))
	// the following loop computes the evaluations of f(r) accordingly:
	//		f(r, b₂, ..., bₙ) = f(0, b₂, ..., bₙ) + r(f(1, b₂, ..., bₙ) - f(0, b₂, ..., bₙ))
	for i := 0; i < mid; i++ {
		// table[i] ← table[i] + r (table[i + mid] - table[i])
		top[i].Sub(&top[i], &bottom[i])
		top[i].Mul(&top[i], &r)
		bottom[i].Add(&bottom[i], &top[i])
	}

	*m = (*m)[:mid]
}

//TODO: See if the general version is needed anywhere
/*
// Folds one part of the table
func (bkt *MultiLin) FoldChunk(r fr.Element, start, stop int) {
	mid := len(*bkt) / 2
	bottom, top := (*bkt)[:mid], (*bkt)[mid:]
	for i := start; i < stop; i++ {
		// updating bookkeeping table
		// table[i] <- table[i] + r (table[i + mid] - table[i])
		top[i].Sub(&top[i], &bottom[i])
		top[i].Mul(&top[i], &r)
		bottom[i].Add(&bottom[i], &top[i])
	}
}*/

// Evaluate extrapolate the value of the multilinear polynomial corresponding to m
// on the given coordinates
func (m MultiLin) Evaluate(coordinates []fr.Element) fr.Element {
	// Folding is a mutating operation
	bkCopy := m.DeepCopy()

	// Evaluate step by step through repeated folding (i.e. evaluation at the first remaining variable)
	for _, r := range coordinates {
		bkCopy.Fold(r)
	}

	return bkCopy[0]
}

// DeepCopy creates a deep copy of a book-keeping table.
// Both multilinear interpolation and sumcheck require folding an underlying
// array, but folding changes the array. To do both one requires a deep copy
// of the book-keeping table.
func (m MultiLin) DeepCopy() MultiLin {
	tableDeepCopy := make([]fr.Element, len(m))
	copy(tableDeepCopy, m)
	return tableDeepCopy
}

// DeepCopyLarge creates a deep copy of a multi-linear table.
func (m MultiLin) DeepCopyLarge() MultiLin {
	tableDeepCopy := MakeLarge(len(m))
	copy(tableDeepCopy, m)
	return tableDeepCopy
}

// Add two bookKeepingTables
func (m *MultiLin) Add(left, right MultiLin) {
	size := len(left)
	// Check that left and right have the same size
	if len(right) != size {
		panic("Left and right do not have the right size")
	}
	// Reallocate the table if necessary
	if cap(*m) < size {
		*m = make([]fr.Element, size)
	}

	// Resize the destination table
	*m = (*m)[:size]

	// Add elementwise
	for i := 0; i < size; i++ {
		(*m)[i].Add(&left[i], &right[i])
	}
}

// RandomMultiLin returns a random array
// TODO: Ask alex if true randomness is required here
/*func RandMultiLin(size int) MultiLin {
	return common.RandomFrArray(size)
}
*/
