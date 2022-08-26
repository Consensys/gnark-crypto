package sis

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"golang.org/x/crypto/blake2b"
)

/*
	Add the optimizations to ring-SIS from
	https://hal.archives-ouvertes.fr/hal-01242273/document
*/

// Largest supported logNormBound
const MAX_LOG_NORM_BOUND = 63

// Bit entropy generable in our SIS instantiation
const MAX_ENTROPY_PER_POSITION = 256

// Considered as the minimal compression ratio
// to use for our instances. Used for sanity-checks
const minimalCompressionRatio int = 2

// Gathers all precomputations for a specific SIS instance
type RingSis struct {
	// ring-SIS commitment key
	// Indexed by n° polynomial || n° evaluation point
	// The polynomials are stored in bit-reversed low-degree evaluation form
	key [][]fr.Element
	// Logarithm in base 2 of the norm bound of the SIS instance
	log2NormBound int
	// FFT domain, for the entry polynomials evaluations
	// domainSize == len(key[0])
	fftDomain *fft.Domain
}

// Generate a new random SIS instance from a seed
// - seed : randomness seed
// - deg : degree of the modulus polynomial of the ring. Must be a power of two.
// - log2NormBound : logarithm in base 2 of the norm bound of the SIS instance
// - numPoly : number of polynomials to generate in the key
func NewRingSis(seed int64, deg, log2NormBound, numPoly int) RingSis {

	// Test if a power of two
	// if !utils.IsPowerOfTwoInt(deg) {
	// 	utils.Panic("deg %v should be a power of two", deg)
	// }

	// Make sure the log2NormBound does not overtake the fr size
	// if log2NormBound >= fr.Bits {
	// 	utils.Panic("log2NormBound %v >= %v", log2NormBound, fr.Bits)
	// }

	// Assert the length can't be negative
	// if !utils.AllStrictPositive(numPoly, deg, log2NormBound) {
	// 	utils.Panic("numPoly %v, deg %v, log2NormBound %v can't be negative", numPoly, deg, log2NormBound)
	// }

	// Not an error itself, but otherwise the function is non-compressive
	// and it is probably a programming mistake
	if numPoly*log2NormBound <= fr.Bits*minimalCompressionRatio {
		panic(
			fmt.Sprintf(
				"Uncompressing SIS instance, certainly a user mistake. The compression ratio is %v / %v and the minimum acceptable is %v",
				numPoly*log2NormBound, fr.Bits, minimalCompressionRatio,
			),
		)
	}

	// Sanity-check : we use Blake2b to generate the the random hashing key
	// But our implementation can only generate to `32 bytes` of entropy
	// Thus if we pick a fr with a larger order, we would not generate as
	// much entropy as we should.
	if fr.Bits > MAX_ENTROPY_PER_POSITION {
		panic("Cannot generate that much entropy")
	}

	// The closure capute the seed and the arguments passed to the function
	// This is to ensure, we don't get the same parameters when rerunning the
	// function.
	genRandom := func(i, j int64) fr.Element {
		var buf bytes.Buffer
		buf.WriteString("SIS")
		binary.Write(&buf, binary.BigEndian, seed)
		binary.Write(&buf, binary.BigEndian, deg)
		binary.Write(&buf, binary.BigEndian, log2NormBound)
		binary.Write(&buf, binary.BigEndian, numPoly)
		binary.Write(&buf, binary.BigEndian, i)
		binary.Write(&buf, binary.BigEndian, j)

		slice := buf.Bytes()
		digest := blake2b.Sum256(slice)

		var res fr.Element
		res.SetBytes(digest[:])
		return res
	}

	fftDomain := fft.NewDomain(uint64(deg) * 2)

	key := make([][]fr.Element, numPoly)
	for poly := range key {
		key[poly] = make([]fr.Element, deg<<1)
		for i := 0; i < deg; i++ {
			key[poly][i] = genRandom(int64(poly), int64(i))
		}
		// Convert in bit-reversed evaluation form
		// as precomputations
		fftDomain.FFT(key[poly], fft.DIF)
	}

	return RingSis{
		key:           key,
		fftDomain:     fftDomain,
		log2NormBound: log2NormBound,
	}
}

// Commits to vs using the ringSIS instance
func (sis RingSis) Commit(vs []fr.Element) []fr.Element {

	// if len(vs) > sis.MaxInputSize() {
	// 	utils.Panic("`len(v)` %v > `maxInputsize` %v", len(vs), sis.MaxInputSize())
	// }

	res := make([]fr.Element, sis.DegPoly()*2)

	// useful to rezeroize tmp, when its reused
	zeroes := make([]fr.Element, sis.DegPoly())

	// First we expand the fr elements in vector of limbs
	bigMask := big.NewInt(int64(1<<sis.log2NormBound) - 1)
	tmpBuff := make([]fr.Element, 0, sis.DegPoly()+sis.NumLimbs())
	tmpPoly := make([]fr.Element, sis.DegPoly()*2)
	curPoly := 0

	for i := range vs {

		var bigint, tmpBig big.Int
		var tmpField fr.Element
		vs[i].ToBigIntRegular(&bigint)

		// Convert the current fr element in a sequence of limbs
		for limb := 0; limb < sis.NumLimbs(); limb++ {
			tmpBig.And(&bigint, bigMask)
			bigint.Rsh(&bigint, uint(sis.log2NormBound))
			tmpField.SetBigInt(&tmpBig)
			tmpBuff = append(tmpBuff, tmpField)
		}

		// If we have enough limbs in the buffer, accumulate it
		// Until we have exhausted all possibilities
		for len(tmpBuff) >= sis.DegPoly() {

			copy(tmpPoly, tmpBuff[:sis.DegPoly()])

			// Get evaluation form
			sis.fftDomain.FFT(tmpPoly, fft.DIF)

			// Multiply by the key
			// vector.MulElementWise(tmpPoly, tmpPoly, sis.key[curPoly])
			for i := 0; i < len(tmpPoly); i++ {
				tmpPoly[i].Mul(&tmpPoly[i], &sis.key[curPoly][i])
			}

			// Accumulate in the result
			// vector.Add(res, res, tmpPoly)
			for i := 0; i < len(res); i++ {
				res[i].Add(&res[i], &tmpPoly[i])
			}

			// Then clean the second part of `tmpPoly`
			copy(tmpPoly[sis.DegPoly():], zeroes)

			// And `clean` tmpBuff out of the values we
			// just accumulated
			remaining := len(tmpBuff) - sis.DegPoly()
			copy(tmpBuff, tmpBuff[sis.DegPoly():])
			tmpBuff = tmpBuff[:remaining]
			curPoly++
		}
	}

	// Accumulate the last chunk
	if len(tmpBuff) > 0 {

		toPad := make([]fr.Element, 2*sis.DegPoly()-len(tmpBuff))
		tmpBuff = append(tmpBuff, toPad...)

		// Get evaluation form
		sis.fftDomain.FFT(tmpBuff, fft.DIF)

		// Multiply by the key
		// vector.MulElementWise(tmpBuff, tmpBuff, sis.key[curPoly])
		for i := 0; i < len(tmpBuff); i++ {
			tmpBuff[i].Mul(&tmpBuff[i], &sis.key[curPoly][i])
		}

		// Accumulate in the result
		// vector.Add(res, res, tmpPoly)
		for i := 0; i < len(res); i++ {
			res[i].Add(&res[i], &tmpPoly[i])
		}

		// No need to clean after ourselve now
	}

	// back in coefficient form
	sis.fftDomain.FFTInverse(res, fft.DIT)

	// carry out the modular reduction by X^deg + 1
	for i := 0; i < sis.DegPoly(); i++ {
		res[i].Sub(&res[i], &res[i+sis.DegPoly()])
	}

	return res[:sis.DegPoly()]
}

// Returns the maximal input size that we can handle
// as number of fr elements that can be committed at once
func (sis RingSis) MaxInputSize() int {
	return len(sis.key) * sis.DegPoly() / sis.NumLimbs()
}

// Returns the degree of the ring modulus polynomial
func (sis RingSis) DegPoly() int {
	return len(sis.key[0]) / 2
}

// Returns the number of limbs per entries with the parameters of the instance
func (sis RingSis) NumLimbs() int {
	return GetNumLimbs(sis.log2NormBound)
}

// check the size of the vector to be committed
// - Its size should be divisible by the polynomial degree of the ring-SIS instance
// - Its size should not overtake the size of the commitment key
func (sis RingSis) checkVectorLength(vs []fr.Element) {

}

// Get the key of the SIS instance
func (sis RingSis) Key() [][]fr.Element {
	return sis.key
}

// Return true iff vs hashes into the allegedResult
func (sis RingSis) Check(vs, allegedResult []fr.Element) bool {
	hash := sis.Commit(vs)
	if len(allegedResult) != len(hash) {
		return false
	}
	res := true
	for i := 0; i < len(vs); i++ {
		res = res && (vs[i] == allegedResult[i])
	}
	return res
}

// Get the number of limbs from the log2NormBound
func GetNumLimbs(logBound int) int {
	// return utils.DivCeil(fr.Bits, logBound)
	return 0
}
