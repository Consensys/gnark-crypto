package ecdsa

import (
	"crypto/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestECDSA(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256K1] test the signing and verification", prop.ForAll(
		func() bool {

			var pp params
			_, g := secp256k1.Generators()
			pp.Base.Set(&g)
			pp.Order = fr.Modulus()

			privKey, _ := pp.GenerateKey(rand.Reader)

			hash := []byte("testing ECDSA")
			signature, _ := pp.Sign(hash, *privKey, rand.Reader)

			return pp.Verify(hash, signature, privKey.PublicKey.Q)
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
