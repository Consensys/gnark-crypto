// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bw6761

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
)

// doubleAndAddStepRef is the reference (pre-optimization) implementation
// of the doubleAndAddStep function. It computes 2P+Q using two field inversions.
//
// This version uses the standard chord-tangent method:
//   - λ1 = (y2-y1)/(x2-x1) for P + Q
//   - λ2 = -λ1 - 2y1/(x3-x1) for doubling and adding back
//
// The optimized version uses the Eisenträger-Lauter-Montgomery formula
// (https://eprint.iacr.org/2003/257) which computes both slopes with a single
// field inversion via Montgomery's batch inversion trick.
func doubleAndAddStepRef(p *G2Affine, evaluations1, evaluations2 *LineEvaluationAff, a *G2Affine) {
	var n, d, l1, x3, l2, x4, y4 fp.Element

	// compute λ1 = (y2-y1)/(x2-x1)
	n.Sub(&p.Y, &a.Y)
	d.Sub(&p.X, &a.X)
	l1.Div(&n, &d)

	// compute x3 =λ1²-x1-x2
	x3.Square(&l1)
	x3.Sub(&x3, &p.X)
	x3.Sub(&x3, &a.X)

	// omit y3 computation

	// compute line1
	evaluations1.R0.Set(&l1)
	evaluations1.R1.Mul(&l1, &p.X)
	evaluations1.R1.Sub(&evaluations1.R1, &p.Y)

	// compute λ2 = -λ1-2y1/(x3-x1)
	n.Double(&p.Y)
	d.Sub(&x3, &p.X)
	l2.Div(&n, &d)
	l2.Add(&l2, &l1)
	l2.Neg(&l2)

	// compute x4 = λ2²-x1-x3
	x4.Square(&l2)
	x4.Sub(&x4, &p.X)
	x4.Sub(&x4, &x3)

	// compute y4 = λ2(x1 - x4)-y1
	y4.Sub(&p.X, &x4)
	y4.Mul(&l2, &y4)
	y4.Sub(&y4, &p.Y)

	// compute line2
	evaluations2.R0.Set(&l2)
	evaluations2.R1.Mul(&l2, &p.X)
	evaluations2.R1.Sub(&evaluations2.R1, &p.Y)

	p.X.Set(&x4)
	p.Y.Set(&y4)
}
