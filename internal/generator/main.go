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
	"github.com/consensys/gurvy/internal/generator/curve"
	"github.com/consensys/gurvy/internal/generator/edwards"
	"github.com/consensys/gurvy/internal/generator/fft"
	"github.com/consensys/gurvy/internal/generator/pairing"
	"github.com/consensys/gurvy/internal/generator/tower"
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, "gurvy")

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
			curveDir := filepath.Join(baseDir, conf.Name)

			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(curveDir, "fr")))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(curveDir, "fp")))

			assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), bgen))
			assertNoError(curve.Generate(conf, curveDir, bgen))
			assertNoError(pairing.Generate(conf, curveDir, bgen))
			assertNoError(edwards.Generate(conf, filepath.Join(curveDir, "twistededwards"), bgen))
			assertNoError(fft.Generate(conf, filepath.Join(curveDir, "fr", "fft"), bgen))

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

const (
	fpTower         = "fptower"
	copyrightHolder = "ConsenSys Software Inc."
	baseDir         = "../../"
)
