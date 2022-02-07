package field

import (
	"crypto/rand"
	"fmt"
	"github.com/leanovate/gopter/gen"
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
	genF := genField(t)

	properties.Property("must recover initial non-montgomery value by repeated halving", prop.ForAll(
		func(f *Field, i *big.Int) (bool, error) {

			// turn into mont
			var mont big.Int
			f.ToMont(&mont, i)
			f.FromMont(&mont, &mont)

			return mont.Cmp(i) == 0, nil
		}, genF, genF.Map(randomElement)),
	)

	properties.Property("turning R into montgomery form must match the R value from field", prop.ForAll(
		func(f *Field) (bool, error) {
			// test if using the same R
			i := big.NewInt(1)
			i.Lsh(i, 64*uint(f.NbWords))
			f.ToMont(i, i)

			err := BigIntMatchUint64Slice(i, f.RSquare)
			return err == nil, err
		}, genF),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestBigIntMatchUint64Slice(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	genF := genField(t)

	properties.Property("random big.int must match uint64 slice made out of .Bytes()", prop.ForAll(
		func(f *Field, i *big.Int) (bool, error) {
			bytes := i.Bytes()
			ints := make([]uint64, (len(bytes)-1)/8+1)

			for j := 0; j < len(bytes); j++ {
				ints[j/8] ^= uint64(bytes[len(bytes)-1-j]) << (8 * (j % 8))
			}

			err := BigIntMatchUint64Slice(i, ints)
			return err == nil, err
		}, genF, genF.Map(randomElement2)))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestQuadExtensionSqrt(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	genF := genField(t)

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

			f := NewTower(base, 2, base.NonResidue.Int64()) //TODO: Derive extension generator from field generator
			i := []big.Int{*i0, *i1}

			if z := f.Sqrt(i); z != nil {
				qrDetected++
				z = f.Mul(z, z)
				return f.Equal(i, z), nil
			}
			return true, nil
		}, genF, genF.Map(randomElement), genF.Map(randomElement)))

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

func TestQuadExtensionMul(t *testing.T) {

	verifyMul := func(base *Field, x8Slice [][]uint8, y8Slice [][]uint8) (bool, error) {
		var nonRes big.Int
		base.FromMont(&nonRes, &base.NonResidue)
		if !nonRes.IsInt64() {
			return false, fmt.Errorf("non-residue too large: %v", nonRes)
		}

		f := NewTower(base, 2, base.NonResidue.Int64())
		x := uint8SliceSliceToBigIntSlice(&f, x8Slice)
		y := uint8SliceSliceToBigIntSlice(&f, y8Slice)

		z := f.Mul(x, y)

		var z0, z1, u big.Int

		base.
			Mul(&z0, &x[0], &y[0]).
			Mul(&u, &x[1], &y[1]).
			Mul(&u, &u, big.NewInt(base.NonResidue.Int64())).
			Add(&z0, &z0, &u)

		base.
			Mul(&z1, &x[0], &y[1]).
			Mul(&u, &x[1], &y[0]).
			Add(&z1, &z1, &u)

		return z0.Cmp(&z[0]) == 0 && z1.Cmp(&z[1]) == 0, nil
	}
	genF := genField(t)
	parameters := gopter.DefaultTestParameters()

	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	properties.Property("multiplication should yield the correct value", prop.ForAll(verifyMul, genF, genUint8SliceSlice(2), genUint8SliceSlice(2)))
	properties.TestingRun(t, gopter.ConsoleReporter(false))

	parameters.MinSuccessfulTests = 4
	properties = gopter.NewProperties(parameters)
	properties.Property("multiplication should yield the correct value (small cases)", prop.ForAll(
		verifyMul,
		genF,
		genSmallUint8SliceSlice(2, 3),
		genSmallUint8SliceSlice(2, 3),
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

const minNbWords = 5
const maxNbWords = 37

func genSmallUint8SliceSlice(outerSize int, max uint8) gopter.Gen {
	return gen.SliceOfN(
		outerSize,
		gen.SliceOfN(1, gen.UInt8Range(0, max)),
	)
}

func genUint8SliceSlice(outerSize int) gopter.Gen {
	return gen.SliceOfN(
		outerSize,
		gen.SliceOfN(maxNbWords*8, gen.UInt8()),
	)
}

func uint8SliceSliceToBigIntSlice(f *Extension, in [][]uint8) []big.Int {
	res := make([]big.Int, f.Degree)
	bytes := make([]byte, f.Base.NbWords*8)

	for i := 0; i < len(res); i++ {

		j := 0
		for ; j < len(bytes) && j < len(in[i]); j++ {
			bytes[j] = in[i][len(in[i])-j-1]
		}

		res[i].SetBytes(bytes[:j]).Mod(&res[i], f.Base.ModulusBig)
	}

	return res
}

func randomElement3(p *gopter.GenParameters, f *Extension) []big.Int {
	bytes := make([]byte, f.Base.NbWords*8)
	res := make([]big.Int, f.Degree)

	for i := 0; i < len(res); i++ {

		for j := 0; j < f.Base.NbWords; j++ {
			w := p.NextUint64()

			for k := 0; k < 8; k++ {
				bytes[8*j+k] = byte(w)
				w = w >> k
			}
		}

		res[i].SetBytes(bytes).Mod(&res[i], f.Base.ModulusBig)
	}

	return res
}

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

		field := genField()
		genResult := gopter.NewGenResult(field, gopter.NoShrinker)
		return genResult
	}
}
