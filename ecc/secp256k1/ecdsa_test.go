package secp256k1

import (
	"crypto/rand"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestECDSA(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256K1] check that the generated keys are valid", prop.ForAll(
		func() bool {
			priv, _ := GenerateKey(rand.Reader)

			return priv.PublicKey.A.IsInSubGroup()
		},
	))

	properties.Property("[SECP256K1] test the sign and verify protocol", prop.ForAll(
		func() bool {
			priv, _ := GenerateKey(rand.Reader)

			hashed := []byte("testing")
			r, s, err := Sign(rand.Reader, priv, hashed)
			if err != nil {
				t.Errorf("error signing: %s", err)
				return false
			}

			return Verify(&priv.PublicKey, hashed, r, s)
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
