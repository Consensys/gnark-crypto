package field

import (
	"crypto/rand"
	"fmt"
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
	gen := genFull(t)

	properties.Property("must recover initial non-montgomery value by repeated halving", prop.ForAll(
		func(c fieldInfo) (bool, error) {

			// turn into mont
			mont := c.i
			c.field.IntToMont(&mont)

			// recover initial value by unorthodox means
			// halve nbWords * 64 times
			for c.bitLen = c.nbWords * 64; c.bitLen > 0; c.bitLen-- {
				if mont.Bit(0) != 0 {
					mont.Add(&mont, &c.modulus)
				}
				mont.Rsh(&mont, 1)
			}

			return mont.Cmp(&c.i) == 0, nil
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

func TestBigIntMatchUint64Slice(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1
	properties := gopter.NewProperties(parameters)
	gen := genFull(t)

	properties.Property("random big.int must match uint64 slice made out of .Bytes()", prop.ForAll(
		func(c fieldInfo) (bool, error) {
			bytes := c.i.Bytes()
			ints := make([]uint64, (len(bytes)-1)/8+1)

			fmt.Print("Bytes in hex: [")
			for j := 0; j < len(bytes); j++ {
				fmt.Printf("%x,", bytes[j])
				ints[j/8] ^= uint64(bytes[len(bytes)-1-j]) << (8 * (j % 8))
			}

			fmt.Print("]\nints in hex: [")
			for j := 0; j < len(ints); j++ {
				fmt.Printf("%x,", ints[j])
			}
			fmt.Print("]\nOriginal int in hex: [")
			for j := 0; j < len(c.i.Bits()); j++ {
				fmt.Printf("%x,", c.i.Bits()[j])
			}
			fmt.Println("]")

			err := BigIntMatchUint64Slice(&c.i, ints)
			return err == nil, err
		}, gen))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

type fieldInfo struct {
	nbWords int
	bitLen  int
	modulus big.Int
	field   *Field
	i       big.Int
}

func genFull(t *testing.T) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {

		genFieldInfo := func() fieldInfo {

			var c fieldInfo
			var err error

			c.nbWords = 5 + mrand.Intn(32)
			c.bitLen = c.nbWords*64 - 1 - mrand.Intn(64)

			temp, err := rand.Prime(rand.Reader, c.bitLen)
			if err != nil {
				t.Fatal(err)
			}
			c.modulus.Set(temp)

			c.field, err = NewField("dummy", "DummyElement", c.modulus.Text(10), false)
			if err != nil {
				t.Fatal(err)
			}
			if c.modulus.Bit(0) == 0 {
				panic("Not a prime")
			}

			temp, err = rand.Int(rand.Reader, &c.modulus)
			if err != nil {
				t.Fatal(err)
			}
			c.i.Set(temp)

			return c
		}
		a := genFieldInfo()

		genResult := gopter.NewGenResult(a, gopter.NoShrinker)
		return genResult
	}
}
