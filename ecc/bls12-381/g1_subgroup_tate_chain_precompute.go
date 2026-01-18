package bls12381

import (
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

type chainTable struct {
	q   G1Affine
	tab []fp.Element
}

var (
	precomputeChainTableOnce sync.Once
	precomputedChainTable    chainTable
)

// precomputeChainTableDefault returns a cached precomputation table for the
// chain-based Miller loop.
func precomputeChainTableDefault() chainTable {
	precomputeChainTableOnce.Do(func() {
		precomputedChainTable = generateChainTable(&torsionPoint)
	})
	return precomputedChainTable
}

// generateChainTable precomputes the lookup table for the chain-based Miller loop.
//
// The chain structure for BLS12-381 e2-1:
//  1. SQPL + SSUB (bits 63→62→61): 4 + 2 = 6 entries
//  2. SQPL + SADD (bits 61→60→59): 4 + 2 = 6 entries
//  3. SQPL (bits 59→58→57): 4 entries
//  4. SDADD (bit 57): 4 entries
//     5-8. 4× SQPL (bits 56→48): 16 entries
//  9. SDADD (bit 48): 4 entries
//     10-24. 15× SQPL (bits 47→17): 60 entries
//  25. SQPL + SADD (bits 17→16→15): 4 + 2 = 6 entries
//     26-33. 8× SQPL (bits 15→-1): 32 entries
//
// Total: 6+6+4+4+16+4+60+6+32 = 138 entries
func generateChainTable(q *G1Affine) chainTable {
	tab := make([]fp.Element, 0, 140)

	var t0, t1, qNeg G1Affine
	t0.Set(q)
	qNeg.Neg(q)

	var u0, u1 fp.Element

	// Helper to add SQPL entries (quadrupling)
	// Stores: λ_T, x_{2T}, y_{2T}, λ_{2T}
	// After call: t0 = 4T
	addSQPL := func() {
		// Compute λ_T = 3x²/(2y) for first doubling
		u0.Square(&t0.X)
		u1.Double(&u0)
		u0.Add(&u0, &u1) // 3x²
		u1.Double(&t0.Y)
		u1.Inverse(&u1)
		u1.Neg(&u1)
		u0.Mul(&u0, &u1) // λ_T
		tab = append(tab, u0)

		// Double to get 2T
		t0.Double(&t0)

		// Compute λ_{2T} for second doubling
		u0.Square(&t0.X)
		u1.Double(&u0)
		u0.Add(&u0, &u1) // 3x²
		u1.Double(&t0.Y)
		u1.Inverse(&u1)
		// Note: no negation for λ_{2T} in the table
		u0.Mul(&u0, &u1) // λ_{2T}

		// Store x_{2T}, y_{2T}, λ_{2T}
		tab = append(tab, t0.X, t0.Y, u0)

		// Double again to get 4T
		t0.Double(&t0)
	}

	// Helper to add SADD entries (addition after SQPL)
	// Stores: λ_{T,P}, x_T
	// After call: t0 = t0 + q
	addSADD := func() {
		// Compute λ_{T,P} = (y_T - y_P) / (x_T - x_P)
		u0.Sub(&t0.Y, &q.Y)
		u1.Sub(&t0.X, &q.X)
		u1.Inverse(&u1)
		u0.Mul(&u0, &u1)
		tab = append(tab, u0, t0.X)
		t0.Add(&t0, q)
	}

	// Helper to add SSUB entries (subtraction after SQPL)
	// Stores: λ_{P,-T}, x_{T-P}
	// After call: t0 = t0 - q
	addSSUB := func() {
		// λ_{P,-T} = (y_P + y_T) / (x_P - x_T)
		u0.Add(&q.Y, &t0.Y)
		u1.Sub(&q.X, &t0.X)
		u1.Inverse(&u1)
		u0.Mul(&u0, &u1)

		t0.Add(&t0, &qNeg) // t0 = t0 - q
		tab = append(tab, u0, t0.X)
	}

	// Helper to add SDADD entries (doubling + addition combined)
	// Stores: x_T, y_T, A, B
	// where A = x_T + x_{T+P} + λ_{T,P}·λ_{T+P,T}
	//       B = λ_{T,P} + λ_{T+P,T}
	// After call: t0 = 2T + P
	addSDADD := func() {
		tab = append(tab, t0.X, t0.Y)

		// λ_{T,P} = (y_T - y_P) / (x_T - x_P)
		var lambda1, lambda2 fp.Element
		u0.Sub(&t0.Y, &q.Y)
		u1.Sub(&t0.X, &q.X)
		u1.Inverse(&u1)
		lambda1.Mul(&u0, &u1)

		// Compute T + P
		t1.Add(&t0, q)

		// λ_{T+P,T} = (y_{T+P} - y_T) / (x_{T+P} - x_T)
		u0.Sub(&t1.Y, &t0.Y)
		u1.Sub(&t1.X, &t0.X)
		u1.Inverse(&u1)
		lambda2.Mul(&u0, &u1)

		// A = x_T + x_{T+P} + λ_{T,P}·λ_{T+P,T}
		u0.Mul(&lambda1, &lambda2)
		u0.Add(&u0, &t0.X)
		u0.Add(&u0, &t1.X)

		// B = λ_{T,P} + λ_{T+P,T}
		lambda2.Add(&lambda1, &lambda2)

		tab = append(tab, u0, lambda2)

		// Update t0 to 2T + P = (T + P) + T
		t0.Add(&t1, &t0)
	}

	// ========================================
	// 1. SQPL + SSUB (bits 63→62→61)
	// ========================================
	addSQPL() // 4 entries
	addSSUB() // 2 entries

	// ========================================
	// 2. SQPL + SADD (bits 61→60→59)
	// ========================================
	addSQPL() // 4 entries
	addSADD() // 2 entries

	// ========================================
	// 3. SQPL (bits 59→58→57), no add
	// ========================================
	addSQPL() // 4 entries

	// ========================================
	// 4. SDADD (bit 57, naf[57] = 1)
	// ========================================
	addSDADD() // 4 entries

	// ========================================
	// 5-8. 4× SQPL (bits 56→48)
	// ========================================
	for i := 0; i < 4; i++ {
		addSQPL() // 4 entries each
	}

	// ========================================
	// 9. SDADD (bit 48, naf[48] = 1)
	// ========================================
	addSDADD() // 4 entries

	// ========================================
	// 10-24. 15× SQPL (bits 47→17)
	// ========================================
	for i := 0; i < 15; i++ {
		addSQPL() // 4 entries each
	}

	// ========================================
	// 25. SQPL + SADD (bits 17→16→15)
	// ========================================
	addSQPL() // 4 entries
	addSADD() // 2 entries

	// ========================================
	// 26-33. 8× SQPL (bits 15→-1)
	// ========================================
	for i := 0; i < 8; i++ {
		addSQPL() // 4 entries each
	}

	return chainTable{
		q:   *q,
		tab: tab,
	}
}
