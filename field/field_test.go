package field

import (
	"crypto/rand"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"math/big"
	mrand "math/rand"
	"testing"
)

func TestIntToMont(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)
	gen := genFull(t)

	properties.Property("must recover initial non-montgomery value by repeated halving", prop.ForAll(
		func(c fieldInfo) (bool, error) {

			i, err := rand.Int(rand.Reader, c.modulus)
			if err != nil {
				return false, err
			}

			// turn into mont
			mont := *i
			c.field.IntToMont(&mont)

			// recover initial value by unorthodox means
			// halve nbWords * 64 times
			for c.bitLen = c.nbWords * 64; c.bitLen > 0; c.bitLen-- {
				if mont.Bit(0) != 0 {
					mont.Add(&mont, c.modulus)
				}
				mont.Rsh(&mont, 1)
			}

			return mont.Cmp(i) == 0, nil
		}, gen),
	)

	properties.Property("turning R into montgomery form must match the R value from field", prop.ForAll(
		func(c fieldInfo) (bool, error) {
			// test if using the same R
			i := big.NewInt(1)
			i.Lsh(i, 64*uint(c.nbWords))
			c.field.IntToMont(i)

			err := BigIntMatchUint64Slice(i, c.field.RSquare)
			return err == nil, err
		}, gen),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

type fieldInfo struct {
	nbWords int
	bitLen  int
	modulus *big.Int
	field   *Field
}

func genFull(t *testing.T) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {

		genFieldInfo := func() fieldInfo {

			var c fieldInfo
			var err error

			c.nbWords = 5 + mrand.Intn(32)
			c.bitLen = c.nbWords*64 - 1 - mrand.Intn(64)

			c.modulus, err = rand.Prime(rand.Reader, c.bitLen)
			if err != nil {
				t.Fatal(err)
			}

			c.field, err = NewField("dummy", "DummyElement", c.modulus.Text(10), false)
			if err != nil {
				t.Fatal(err)
			}
			if c.modulus.Bit(0) == 0 {
				panic("Not a prime")
			}

			return c
		}
		a := genFieldInfo()

		genResult := gopter.NewGenResult(a, gopter.NoShrinker)
		return genResult
	}
}
