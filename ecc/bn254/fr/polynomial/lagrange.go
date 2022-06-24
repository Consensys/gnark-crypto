package polynomial

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func GetLagrangeBasis(domainSize int) []Polynomial {
	/*if lagrangeBasis[domainSize] == nil {
		lagrangeBasis[domainSize] = precomputeLagrangeCoefficients(domainSize)
	}*/
	return lagrangeBasis[domainSize]
}

const maxLagrangeDomainSize uint8 = 12

var lagrangeBasis [][]Polynomial

// precomputeLagrangeCoefficients precomputes in explicit coefficient form for each 0 ≤ l < domainSize the polynomial
// pₗ := X (X-1) ... (X-l-1) (X-l+1) ... (X - domainSize + 1) / ( l (l-1) ... 2 (-1) ... (l - domainSize +1) )
// Note that pₗ(l) = 1 and pₗ(n) = 0 if 0 ≤ l < domainSize, n ≠ l
func precomputeLagrangeCoefficients(domainSize uint8) []Polynomial {

	constTerms := make([]fr.Element, domainSize)
	for i := uint8(0); i < domainSize; i++ {
		constTerms[i].SetInt64(-int64(i))
	}

	res := make([]Polynomial, domainSize)
	multScratch := make(Polynomial, domainSize-1)

	// compute pₗ
	for l := uint8(0); l < domainSize; l++ {

		// TODO: Optimize this with some trees? O(log(domainSize)) polynomial mults instead of O(domainSize)? Then again it would be fewer big poly mults vs many small poly mults
		d := uint8(0) //n is the current degree of res
		for i := uint8(0); i < domainSize; i++ {
			if i == l {
				continue
			}
			if d == 0 {
				res[l] = make(Polynomial, domainSize)
				res[l][domainSize-2] = constTerms[i]
				res[l][domainSize-1].SetOne()
			} else {
				current := res[l][domainSize-d-2:]
				timesConst := multScratch[domainSize-d-2:]

				timesConst.Scale(&constTerms[i], current[1:]) //TODO: Directly double and add since constTerms are tiny? (even less than 4 bits)
				nonLeading := current[0 : d+1]

				nonLeading.Add(nonLeading, timesConst)

			}
			d++
		}

	}

	// We have pₗ(i≠l)=0. Now scale so that pₗ(l)=1
	// Replace the constTerms with norms
	for l := uint8(0); l < domainSize; l++ {
		constTerms[l].Neg(&constTerms[l])
		constTerms[l] = res[l].Eval(&constTerms[l])
	}
	constTerms = fr.BatchInvert(constTerms)
	for l := uint8(0); l < domainSize; l++ {
		res[l].ScaleInPlace(&constTerms[l])
	}

	return res
}

// InterpolateOnRange performs the interpolation of the given list of elements
// On the range [0, 1,..., len(values) - 1]
func InterpolateOnRange(values []fr.Element) Polynomial {
	nEvals := len(values)
	lagrange := GetLagrangeBasis(nEvals)
	result := make([]fr.Element, nEvals)
	var tmp fr.Element

	for i, value := range values {
		for j, lagrangeCoeff := range lagrange[i] {
			tmp.Set(&lagrangeCoeff)
			tmp.Mul(&tmp, &value)
			result[j].Add(&result[j], &tmp)
		}
	}

	return result
}
