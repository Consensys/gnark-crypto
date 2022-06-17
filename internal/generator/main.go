package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/field"
	"github.com/consensys/gnark-crypto/internal/field/generator"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/mimc"
	"github.com/consensys/gnark-crypto/internal/generator/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/edwards"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa"
	"github.com/consensys/gnark-crypto/internal/generator/fft"
	fri "github.com/consensys/gnark-crypto/internal/generator/fri/template"
	"github.com/consensys/gnark-crypto/internal/generator/kzg"
	"github.com/consensys/gnark-crypto/internal/generator/pairing"
	"github.com/consensys/gnark-crypto/internal/generator/permutation"
	"github.com/consensys/gnark-crypto/internal/generator/plookup"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/tower"
)

const (
	copyrightHolder = "ConsenSys Software Inc."
	copyrightYear   = 2020
	baseDir         = "../../"
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, copyrightYear, "consensys/gnark-crypto")

//go:generate go run main.go
func main() {
	var wg sync.WaitGroup
	for _, conf := range config.Curves {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf config.Curve) {
			defer wg.Done()
			var err error

			curveDir := filepath.Join(baseDir, "ecc", conf.Name)
			// generate base field
			conf.Fp, err = field.NewFieldConfig("fp", "Element", conf.FpModulus, true)
			assertNoError(err)

			conf.Fr, err = field.NewFieldConfig("fr", "Element", conf.FrModulus, true)
			assertNoError(err)

			conf.FpUnusedBits = 64 - (conf.Fp.NbBits % 64)

			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(curveDir, "fr")))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(curveDir, "fp")))

			// generate tower of extension
			assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), bgen))

			// generate fft on fr
			assertNoError(fft.Generate(conf, filepath.Join(curveDir, "fr", "fft"), bgen))

			// generate polynomial on fr
			assertNoError(polynomial.Generate(conf, filepath.Join(curveDir, "fr", "polynomial"), bgen))

			// generate kzg on fr
			assertNoError(kzg.Generate(conf, filepath.Join(curveDir, "fr", "kzg"), bgen))

			// generate plookup on fr
			assertNoError(plookup.Generate(conf, filepath.Join(curveDir, "fr", "plookup"), bgen))

			// generate permutation on fr
			assertNoError(permutation.Generate(conf, filepath.Join(curveDir, "fr", "permutation"), bgen))

			// generate mimc on fr
			assertNoError(mimc.Generate(conf, filepath.Join(curveDir, "fr", "mimc"), bgen))

			// generate eddsa on companion curves
			assertNoError(fri.Generate(conf, filepath.Join(curveDir, "fr", "fri"), bgen))

			// generate G1, G2, multiExp, ...
			assertNoError(ecc.Generate(conf, curveDir, bgen))

			// generate pairing tests
			assertNoError(pairing.Generate(conf, curveDir, bgen))

		}(conf)

	}

	wg.Wait()

	for _, conf := range config.TwistedEdwardsCurves {
		wg.Add(1)

		go func(conf config.TwistedEdwardsCurve) {
			defer wg.Done()

			curveDir := filepath.Join(baseDir, "ecc", conf.Name, conf.Package)
			// generate twisted edwards companion curves
			assertNoError(edwards.Generate(conf, curveDir, bgen))

			// generate eddsa on companion curves
			assertNoError(eddsa.Generate(conf, curveDir, bgen))
		}(conf)

	}

	wg.Wait()

	// format the whole directory

	cmd := exec.Command("gofmt", "-s", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	cmd = exec.Command("asmfmt", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	//mathfmt doesn't accept directories. TODO: PR?
	/*cmd = exec.Command("mathfmt", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())*/
}

func assertNoError(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}
}
