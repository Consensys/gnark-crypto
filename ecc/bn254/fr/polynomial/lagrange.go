package polynomial

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func init() {
	//TODO: Check for whether already computed in the Getter or this?
	lagrangeBasis = make([][]Polynomial, maxLagrangeDomainSize+1)

	//size = 0: Cannot extrapolate with no data points

	//size = 1: Constant polynomial
	lagrangeBasis[1] = []Polynomial{make(Polynomial, 1)}
	lagrangeBasis[1][0][0].SetOne()

	//for size ≥ 2, the function works
	for size := uint8(2); size <= maxLagrangeDomainSize; size++ {
		lagrangeBasis[size] = computeLagrangeBasis(size)
	}
}

func getLagrangeBasis(domainSize int) []Polynomial {
	//TODO: Precompute everything at init or this?
	/*if lagrangeBasis[domainSize] == nil {
		lagrangeBasis[domainSize] = computeLagrangeBasis(domainSize)
	}*/
	return lagrangeBasis[domainSize]
}

const maxLagrangeDomainSize uint8 = 12

var lagrangeBasis [][]Polynomial

// computeLagrangeBasis precomputes in explicit coefficient form for each 0 ≤ l < domainSize the polynomial
// pₗ := X (X-1) ... (X-l-1) (X-l+1) ... (X - domainSize + 1) / ( l (l-1) ... 2 (-1) ... (l - domainSize +1) )
// Note that pₗ(l) = 1 and pₗ(n) = 0 if 0 ≤ l < domainSize, n ≠ l
func computeLagrangeBasis(domainSize uint8) []Polynomial {

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
// TODO: Am I crazy or is this EXTRApolation and not INTERpolation
func InterpolateOnRange(values []fr.Element) Polynomial {
	nEvals := len(values)
	lagrange := getLagrangeBasis(nEvals)

	var res Polynomial
	res.Scale(&values[0], lagrange[0])

	temp := make(Polynomial, nEvals)

	for i := 1; i < nEvals; i++ {
		temp.Scale(&values[i], lagrange[i])
		res.Add(res, temp)
	}

	return res
}
