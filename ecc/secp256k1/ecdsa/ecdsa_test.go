package ecdsa

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestECDSA(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256K1] test the signing and verification", prop.ForAll(
		func() bool {

			_, g := secp256k1.Generators()
			dsa := NewParams(g)
			privKey, _ := dsa.GenerateKey(rand.Reader)

			hashval := big.NewInt(int64(40))
			sig, _ := dsa.Sign(hashval, &privKey.D, rand.Reader)

			return dsa.Verify(hashval, sig, privKey.PublicKey.A)
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
