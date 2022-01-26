package field

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"
)

func TestIntToMont(t *testing.T) {

	nbWords := 5 + mrand.Intn(32)
	bitLen := nbWords*64 - 1 - mrand.Intn(64)

	modulus, err := rand.Prime(rand.Reader, bitLen)
	if err != nil {
		t.Fatal(err)
	}
	if modulus.Bit(0) == 0 {
		panic("Not a prime")
	}

	i, err := rand.Int(rand.Reader, modulus)
	if err != nil {
		t.Fatal(err)
	}

	// turn into mont
	mont := *i
	IntToMont(&mont, modulus)

	// recover initial value by unorthodox means
	// halve nbWords * 64 times
	for bitLen = nbWords * 64; bitLen > 0; bitLen-- {
		if mont.Bit(0) != 0 {
			mont.Add(&mont, modulus)
		}
		mont.Rsh(&mont, 1)
	}

	if mont.Cmp(i) != 0 {
		t.Fail()
	}
}
