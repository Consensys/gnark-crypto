package bls12381

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

const maxMultiExpSmallD = 30

// MultiExpSmallD implements the paper's "multidimensional Shamir's trick" (Method 1).
// It is optimized for small d and uses precomputed subset sums.
func (p *G1Affine) MultiExpSmallD(points []G1Affine, scalars []fr.Element) (*G1Affine, error) {
	var jac G1Jac
	if _, err := jac.MultiExpSmallD(points, scalars); err != nil {
		return nil, err
	}
	p.FromJacobian(&jac)
	return p, nil
}

// MultiExpSmallD implements the paper's "multidimensional Shamir's trick" (Method 1).
// It is optimized for small d and uses precomputed subset sums.
func (p *G1Jac) MultiExpSmallD(points []G1Affine, scalars []fr.Element) (*G1Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if nbPoints == 0 {
		p.Set(&g1Infinity)
		return p, nil
	}
	if nbPoints > maxMultiExpSmallD {
		return nil, errors.New("multi-exp small-d: d is too large for precomputation")
	}

	scalarBits := make([][fr.Limbs]uint64, nbPoints)
	for i := 0; i < nbPoints; i++ {
		scalarBits[i] = scalars[i].Bits()
	}

	tableSize := 1 << nbPoints
	table := make([]G1Jac, tableSize)
	table[0].Set(&g1Infinity)

	for i := 0; i < nbPoints; i++ {
		step := 1 << i
		for mask := 0; mask < step; mask++ {
			table[mask+step].Set(&table[mask])
			table[mask+step].AddMixed(&points[i])
		}
	}

	var acc G1Jac
	acc.Set(&g1Infinity)

	for j := fr.Bits - 1; j >= 0; j-- {
		acc.DoubleAssign()

		word := j / 64
		mask := uint64(1) << uint(j%64)
		xj := 0
		for i := 0; i < nbPoints; i++ {
			if scalarBits[i][word]&mask != 0 {
				xj |= 1 << i
			}
		}
		if xj != 0 {
			acc.AddAssign(&table[xj])
		}

		if j == 0 {
			break
		}
	}

	p.Set(&acc)
	return p, nil
}

// MultiExpSmallD implements the paper's "multidimensional Shamir's trick" (Method 1).
// It is optimized for small d and uses precomputed subset sums.
func (p *G2Affine) MultiExpSmallD(points []G2Affine, scalars []fr.Element) (*G2Affine, error) {
	var jac G2Jac
	if _, err := jac.MultiExpSmallD(points, scalars); err != nil {
		return nil, err
	}
	p.FromJacobian(&jac)
	return p, nil
}

// MultiExpSmallD implements the paper's "multidimensional Shamir's trick" (Method 1).
// It is optimized for small d and uses precomputed subset sums.
func (p *G2Jac) MultiExpSmallD(points []G2Affine, scalars []fr.Element) (*G2Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if nbPoints == 0 {
		p.Set(&g2Infinity)
		return p, nil
	}
	if nbPoints > maxMultiExpSmallD {
		return nil, errors.New("multi-exp small-d: d is too large for precomputation")
	}

	scalarBits := make([][fr.Limbs]uint64, nbPoints)
	for i := 0; i < nbPoints; i++ {
		scalarBits[i] = scalars[i].Bits()
	}

	tableSize := 1 << nbPoints
	table := make([]G2Jac, tableSize)
	table[0].Set(&g2Infinity)

	for i := 0; i < nbPoints; i++ {
		step := 1 << i
		for mask := 0; mask < step; mask++ {
			table[mask+step].Set(&table[mask])
			table[mask+step].AddMixed(&points[i])
		}
	}

	var acc G2Jac
	acc.Set(&g2Infinity)

	for j := fr.Bits - 1; j >= 0; j-- {
		acc.DoubleAssign()

		word := j / 64
		mask := uint64(1) << uint(j%64)
		xj := 0
		for i := 0; i < nbPoints; i++ {
			if scalarBits[i][word]&mask != 0 {
				xj |= 1 << i
			}
		}
		if xj != 0 {
			acc.AddAssign(&table[xj])
		}

		if j == 0 {
			break
		}
	}

	p.Set(&acc)
	return p, nil
}
