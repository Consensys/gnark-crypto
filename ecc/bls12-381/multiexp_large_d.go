package bls12381

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// MultiExpLargeD implements the paper's "multidimensional double-and-add" (Method 2).
// It is optimized for large d, avoiding precomputation and buckets.
func (p *G1Affine) MultiExpLargeD(points []G1Affine, scalars []fr.Element) (*G1Affine, error) {
	var jac G1Jac
	if _, err := jac.MultiExpLargeD(points, scalars); err != nil {
		return nil, err
	}
	p.FromJacobian(&jac)
	return p, nil
}

// MultiExpLargeD implements the paper's "multidimensional double-and-add" (Method 2).
// It is optimized for large d, avoiding precomputation and buckets.
func (p *G1Jac) MultiExpLargeD(points []G1Affine, scalars []fr.Element) (*G1Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if nbPoints == 0 {
		p.Set(&g1Infinity)
		return p, nil
	}

	scalarBits := make([][fr.Limbs]uint64, nbPoints)
	for i := 0; i < nbPoints; i++ {
		scalarBits[i] = scalars[i].Bits()
	}

	var acc G1Jac
	acc.Set(&g1Infinity)

	for j := fr.Bits - 1; j >= 0; j-- {
		acc.DoubleAssign()

		var t G1Jac
		t.Set(&g1Infinity)

		word := j / 64
		mask := uint64(1) << uint(j%64)

		for i := 0; i < nbPoints; i++ {
			if scalarBits[i][word]&mask != 0 {
				t.AddMixed(&points[i])
			}
		}

		acc.AddAssign(&t)

		if j == 0 {
			break
		}
	}

	p.Set(&acc)
	return p, nil
}

// MultiExpLargeD implements the paper's "multidimensional double-and-add" (Method 2).
// It is optimized for large d, avoiding precomputation and buckets.
func (p *G2Affine) MultiExpLargeD(points []G2Affine, scalars []fr.Element) (*G2Affine, error) {
	var jac G2Jac
	if _, err := jac.MultiExpLargeD(points, scalars); err != nil {
		return nil, err
	}
	p.FromJacobian(&jac)
	return p, nil
}

// MultiExpLargeD implements the paper's "multidimensional double-and-add" (Method 2).
// It is optimized for large d, avoiding precomputation and buckets.
func (p *G2Jac) MultiExpLargeD(points []G2Affine, scalars []fr.Element) (*G2Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if nbPoints == 0 {
		p.Set(&g2Infinity)
		return p, nil
	}

	scalarBits := make([][fr.Limbs]uint64, nbPoints)
	for i := 0; i < nbPoints; i++ {
		scalarBits[i] = scalars[i].Bits()
	}

	var acc G2Jac
	acc.Set(&g2Infinity)

	for j := fr.Bits - 1; j >= 0; j-- {
		acc.DoubleAssign()

		var t G2Jac
		t.Set(&g2Infinity)

		word := j / 64
		mask := uint64(1) << uint(j%64)

		for i := 0; i < nbPoints; i++ {
			if scalarBits[i][word]&mask != 0 {
				t.AddMixed(&points[i])
			}
		}

		acc.AddAssign(&t)

		if j == 0 {
			break
		}
	}

	p.Set(&acc)
	return p, nil
}
