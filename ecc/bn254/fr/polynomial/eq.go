package polynomial

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// EvalEq computes Eq(q₁, ... , qₙ, h₁, ... , hₙ) = Π₁ⁿ Eq(qᵢ, hᵢ)
// where Eq(x,y) = xy + (1-x)(1-y) = 1 - x - y + xy + xy interpolates
//      _________________
//      |       |       |
//      |   0   |   1   |
//      |_______|_______|
//  y   |       |       |
//      |   1   |   0   |
//      |_______|_______|
//
//              x
// In other words the polynomial evaluated here is the multilinear extrapolation of
// one that evaluates to q' == h' for vectors q', h' of binary values
func EvalEq(q, h []fr.Element) fr.Element {
	var res, nxt, one, sum fr.Element
	one.SetOne()
	for i := 0; i < len(q); i++ {
		nxt.Mul(&q[i], &h[i]) // nxt <- qᵢ * hᵢ
		nxt.Double(&nxt)      // nxt <- 2 * qᵢ * hᵢ
		nxt.Add(&nxt, &one)   // nxt <- 1 + 2 * qᵢ * hᵢ
		sum.Add(&q[i], &h[i]) // sum <- qᵢ + hᵢ	TODO: Why not subtract one by one from nxt? More parallel?

		if i == 0 {
			res.Sub(&nxt, &sum) // nxt <- 1 + 2 * qᵢ * hᵢ - qᵢ - hᵢ
		} else {
			nxt.Sub(&nxt, &sum) // nxt <- 1 + 2 * qᵢ * hᵢ - qᵢ - hᵢ
			res.Mul(&res, &nxt) // res <- res * nxt
		}
	}
	fmt.Println(res.Text(10))
	return res
}

// Eq sets m to the representation of the polynomial Eq(q₁, ..., qₙ, *, ..., *) × m[0]
func (m *MultiLin) Eq(q []fr.Element) {
	n := len(q)

	if len(*m) != 1<<n {
		n := Make(1 << n)
		n[0].Set(&(*m)[0])
		//TODO: Dump m?
		*m = n
	}

	//At the end of each iteration, m(h₁, ..., hₙ) = Eq(q₁, ..., qᵢ₊₁, h₁, ..., hᵢ₊₁)
	for i, qI := range q { // In the comments we use a 1-based index so qI = qᵢ₊₁
		// go through all assignments of (b₁, ..., bᵢ) ∈ {0,1}ⁱ
		for j := 0; j < (1 << i); j++ {
			j0 := j << (n - i)                 // bᵢ₊₁ = 0
			j1 := j0 + 1<<(n-1-i)              // bᵢ₊₁ = 1
			(*m)[j1].Mul(&qI, &(*m)[j0])       // Eq(q₁, ..., qᵢ₊₁, b₁, ..., bᵢ, 1) = Eq(q₁, ..., qᵢ, b₁, ..., bᵢ) Eq(qᵢ₊₁, 1) = Eq(q₁, ..., qᵢ, b₁, ..., bᵢ) qᵢ₊₁
			(*m)[j0].Sub(&(*m)[j0], &(*m)[j1]) // Eq(q₁, ..., qᵢ₊₁, b₁, ..., bᵢ, 0) = Eq(q₁, ..., qᵢ, b₁, ..., bᵢ) Eq(qᵢ₊₁, 0) = Eq(q₁, ..., qᵢ, b₁, ..., bᵢ) (1-qᵢ₊₁)
		}
	}

}
