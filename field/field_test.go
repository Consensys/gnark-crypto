package field

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	mrand "math/rand"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestIntToMont(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	gen := genField(t)

	properties.Property("must recover initial non-montgomery value by repeated halving", prop.ForAll(
		func(f *Field, i *big.Int) (bool, error) {

			// turn into mont
			var mont big.Int
			f.ToMont(&mont, i)
			f.FromMont(&mont, &mont)

			return mont.Cmp(i) == 0, nil
		}, gen, gen.Map(randomElement)),
	)

	properties.Property("turning R into montgomery form must match the R value from field", prop.ForAll(
		func(f *Field) (bool, error) {
			// test if using the same R
			i := big.NewInt(1)
			i.Lsh(i, 64*uint(f.NbWords))
			f.ToMont(i, i)

			err := BigIntMatchUint64Slice(i, f.RSquare)
			return err == nil, err
		}, gen),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestBigIntMatchUint64Slice(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	gen := genField(t)

	properties.Property("random big.int must match uint64 slice made out of .Bytes()", prop.ForAll(
		func(f *Field, i *big.Int) (bool, error) {
			bytes := i.Bytes()
			ints := make([]uint64, (len(bytes)-1)/8+1)

			for j := 0; j < len(bytes); j++ {
				ints[j/8] ^= uint64(bytes[len(bytes)-1-j]) << (8 * (j % 8))
			}

			err := BigIntMatchUint64Slice(i, ints)
			return err == nil, err
		}, gen, gen.Map(randomElement2)))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestQuadExtensionSqrt(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	gen := genField(t)

	runs := 0
	qrDetected := 0

	properties.Property("computed square roots must square back to original value", prop.ForAll(
		func(base *Field, i0, i1 *big.Int) (bool, error) {
			runs++

			var nonRes big.Int
			base.FromMont(&nonRes, &base.NonResidue)
			if !nonRes.IsInt64() {
				return false, fmt.Errorf("non-residue too large: %v", nonRes)
			}

			f := NewTower(base, 2, base.NonResidue.Int64())
			i := []big.Int{*i0, *i1}

			var z []big.Int
			if f.Sqrt(&z, i) {
				qrDetected++
				f.Mul(&z, z, z)
				return f.Equal(i, z), nil
			}
			return true, nil
		}, gen, gen.Map(randomElement), gen.Map(randomElement)))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

	// Hypothesis testing: too many or too few qr detected?
	// Y: ratio of qr observed; mean of independent 50/50 bernoulli trials

	yObservedDev := float64(qrDetected)/float64(runs) - 0.5
	yStdDevInv := math.Sqrt(float64(2 * runs))
	yDev := yObservedDev * yStdDevInv
	t.Logf("%f of observations decided as QR, off by %f standard deviatios from the expected 50%%.", yObservedDev, yDev)
	if yDev > 3.0 || yDev < -3.0 {
		t.Error("Hypothesis test failed. The probability of this happening to correct code is 0.27%, less than one in 370.")
	}
}

const minNbWords = 5
const maxNbWords = 37

type fieldWithElements struct {
	f        *Field
	elements []*big.Int
}

/*func withRandomElements(g gopter.Gen, elementsNum int) gopter.Gen {
	return g.FlatMap(func(f *Field) (gopter.Gen, error) {
		genFieldWithElements := func() (fieldWithElements, error) {
			ints, err := randomElements(f, elementsNum)
			if err != nil {
				return fieldWithElements{}, err
			}
			return fieldWithElements{f, ints}, nil
		}

		genResult := gopter.NewGenResult(genFieldWithElements, gopter.NoShrinker)
		return genResult, nil

	})
}*/

func randomElement2(f func() *Field) []*big.Int {
	length := 2
	res := make([]*big.Int, length)
	var err error
	for n := 0; n < length; n++ {
		res[n], err = rand.Int(rand.Reader, f().ModulusBig)
		if err != nil {
			return nil
		}
	}
	return res
}

func randomElement(f *Field, length int) ([]*big.Int, error) {
	res := make([]*big.Int, length)
	var err error
	for n := 0; n < length; n++ {
		res[n], err = rand.Int(rand.Reader, f.ModulusBig)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func genField(t *testing.T) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {

		genField := func() *Field {

			nbWords := minNbWords + mrand.Intn(maxNbWords-minNbWords)
			bitLen := nbWords*64 - 1 - mrand.Intn(64)

			modulus, err := rand.Prime(rand.Reader, bitLen)
			if err != nil {
				t.Error(err)
			}

			var field *Field
			field, err = NewField("dummy", "DummyElement", modulus.Text(10), false)

			if err == nil {
				if field.NbBits != bitLen || field.NbWords != nbWords {
					err = fmt.Errorf("mismatch: field.NbBits = %d, bitLen = %d, field.NbWords = %d, nbWords = %d", field.NbBits, bitLen, field.NbWords, nbWords)
				}
			}

			if err != nil {
				t.Error(err)
			}
			return field
		}
		genResult := gopter.NewGenResult(genField, gopter.NoShrinker)
		return genResult
	}
}
