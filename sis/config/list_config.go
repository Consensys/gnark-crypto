package config

var Configs []SIS

func init() {

	// bn254
	addConfig(NewConfig(
		"github.com/consensys/gnark-crypto/ecc/bn254/fr",
		"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft",
		"../../ecc/bn254/fr/sis",
	))

	// bls12-377
	addConfig(NewConfig(
		"github.com/consensys/gnark-crypto/ecc/bls12-377/fr",
		"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft",
		"../../ecc/bls12-377/fr/sis",
	))

	// goldilocks
	addConfig(NewConfig(
		"github.com/consensys/gnark-crypto/field/goldilocks",
		"github.com/consensys/gnark-crypto/field/goldilocks/fft",
		"../../field/goldilocks/sis",
	))

	// babybear
	addConfig(NewConfig(
		"github.com/consensys/gnark-crypto/field/babybear",
		"github.com/consensys/gnark-crypto/field/babybear/fft",
		"../../field/babybear/sis",
	))

	// koalabear
	addConfig(NewConfig(
		"github.com/consensys/gnark-crypto/field/koalabear",
		"github.com/consensys/gnark-crypto/field/koalabear/fft",
		"../../field/koalabear/sis",
	))

}

func addConfig(c SIS) {
	Configs = append(Configs, c)
}
