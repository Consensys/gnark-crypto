package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/field"
	"github.com/consensys/gurvy/field/generator"
	"github.com/consensys/gurvy/internal/generator/config"
	"github.com/consensys/gurvy/internal/generator/crypto/hash/mimc"
	"github.com/consensys/gurvy/internal/generator/crypto/signature/eddsa"
	"github.com/consensys/gurvy/internal/generator/curve"
	"github.com/consensys/gurvy/internal/generator/edwards"
	"github.com/consensys/gurvy/internal/generator/fft"
	"github.com/consensys/gurvy/internal/generator/pairing"
	"github.com/consensys/gurvy/internal/generator/polynomial"
	"github.com/consensys/gurvy/internal/generator/tower"
)

const (
	fpTower         = "fptower"
	copyrightHolder = "ConsenSys Software Inc."
	copyrightYear   = 2020
	baseDir         = "../../"
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, copyrightYear, "gurvy")

//go:generate go run main.go
func main() {
	var wg sync.WaitGroup
	for _, conf := range config.Curves {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf config.Curve) {
			defer wg.Done()
			conf.Fp, _ = field.NewField("fp", "Element", conf.FpModulus)
			conf.Fr, _ = field.NewField("fr", "Element", conf.FrModulus)
			conf.FpUnusedBits = 64 - (conf.Fp.NbBits % 64)
			curveDir := filepath.Join(baseDir, "curve", conf.Name)

			// generate base field
			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(curveDir, "fr")))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(curveDir, "fp")))

			// generate tower of extension
			assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), bgen))

			// generate G1, G2, multiExp, ...
			assertNoError(curve.Generate(conf, curveDir, bgen))

			// generate pairing tests
			assertNoError(pairing.Generate(conf, curveDir, bgen))

			// generate twisted edwards companion curves
			assertNoError(edwards.Generate(conf, filepath.Join(curveDir, "twistededwards"), bgen))

			// generate fft on fr
			assertNoError(fft.Generate(conf, filepath.Join(curveDir, "fr", "fft"), bgen))

			// generate polynomial on fr
			assertNoError(polynomial.Generate(conf, filepath.Join(curveDir, "fr", "polynomial"), bgen))

			// generate mimc on fr
			assertNoError(mimc.Generate(conf, filepath.Join(curveDir, "fr", "mimc"), bgen))

			// generate eddsa on companion curves
			assertNoError(eddsa.Generate(conf, filepath.Join(curveDir, "twistededwards", "eddsa"), bgen))

		}(conf)

	}
	wg.Wait()

	// run go fmt on whole directory
	cmd := exec.Command("gofmt", "-s", "-w", "../../")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())
}

func assertNoError(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}
}
