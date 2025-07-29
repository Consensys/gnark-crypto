package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator"
	fieldConfig "github.com/consensys/gnark-crypto/field/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/mimc"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/poseidon2"
	"github.com/consensys/gnark-crypto/internal/generator/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/ecdsa"
	"github.com/consensys/gnark-crypto/internal/generator/edwards"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa"
	"github.com/consensys/gnark-crypto/internal/generator/fflonk"
	fri "github.com/consensys/gnark-crypto/internal/generator/fri/template"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_curve"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_field"
	"github.com/consensys/gnark-crypto/internal/generator/kzg"
	"github.com/consensys/gnark-crypto/internal/generator/mpcsetup"
	"github.com/consensys/gnark-crypto/internal/generator/pairing"
	"github.com/consensys/gnark-crypto/internal/generator/pedersen"
	"github.com/consensys/gnark-crypto/internal/generator/permutation"
	"github.com/consensys/gnark-crypto/internal/generator/plookup"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/shplonk"
	"github.com/consensys/gnark-crypto/internal/generator/tower"
)

const (
	copyrightHolder = "Consensys Software Inc."
	copyrightYear   = 2020
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, copyrightYear, "consensys/gnark-crypto")

//go:generate go run main.go
func main() {

	baseDir := filepath.Join("..", "..")
	// first we loop through the field arithmetic we must generate.
	// then, we create the common files (only once) for the assembly code.
	asmDirBuildPath := filepath.Join(baseDir, "field", "asm")
	asmDirIncludePath := filepath.Join(baseDir, "..", "field", "asm")

	asmConfig := &fieldConfig.Assembly{BuildDir: asmDirBuildPath, IncludeDir: asmDirIncludePath}
	// this enable the generation of fft functions;
	// the parameters are hard coded in a lookup table for now for the modulus we use.
	fftConfig := &fieldConfig.FFT{}

	var wg sync.WaitGroup
	for _, conf := range config.Curves {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf config.Curve) {
			defer wg.Done()

			var err error

			conf.Fp, err = fieldConfig.NewFieldConfig("fp", "Element", conf.FpModulus, true)
			assertNoError(err)

			conf.Fr, err = fieldConfig.NewFieldConfig("fr", "Element", conf.FrModulus, !conf.Equal(config.STARK_CURVE))
			assertNoError(err)

			curveDir := filepath.Join(baseDir, "ecc", conf.Name)

			conf.FpUnusedBits = 64 - (conf.Fp.NbBits % 64)

			frOpts := []generator.Option{generator.WithASM(asmConfig)}
			if !(conf.Equal(config.STARK_CURVE) || conf.Equal(config.SECP256K1) || conf.Equal(config.GRUMPKIN)) {
				frOpts = append(frOpts, generator.WithFFT(fftConfig), generator.WithIOP())
			}
			if conf.Equal(config.BLS12_377) {
				frOpts = append(frOpts, generator.WithSIS())
			}
			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(curveDir, "fr"), frOpts...))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(curveDir, "fp"), generator.WithASM(asmConfig)))

			// generate ecdsa
			assertNoError(ecdsa.Generate(conf, curveDir, bgen))

			if conf.Equal(config.STARK_CURVE) {
				return // TODO @yelhousni
			}

			// generate G1, G2, multiExp, ...
			assertNoError(ecc.Generate(conf, curveDir, bgen))

			if conf.Equal(config.SECP256K1) {
				return
			}

			// generate mimc on fr
			assertNoError(mimc.Generate(conf, filepath.Join(curveDir, "fr", "mimc"), bgen))

			// generate polynomial on fr
			frInfo := fieldConfig.FieldDependency{
				FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fr",
				FieldPackageName: "fr",
				ElementType:      "fr.Element",
			}
			assertNoError(polynomial.Generate(frInfo, filepath.Join(curveDir, "fr", "polynomial"), true, bgen))

			// generate poseidon2 on fr
			assertNoError(poseidon2.Generate(conf, filepath.Join(curveDir, "fr", "poseidon2"), bgen))

			fpInfo := fieldConfig.FieldDependency{
				FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fp",
				FieldPackageName: "fp",
				ElementType:      "fp.Element",
			}

			// generate wrapped hash-to-field for both fr and fp
			assertNoError(hash_to_field.Generate(frInfo, filepath.Join(curveDir, "fr", "hash_to_field"), bgen))
			assertNoError(hash_to_field.Generate(fpInfo, filepath.Join(curveDir, "fp", "hash_to_field"), bgen))

			// generate hash to curve for both G1 and G2
			assertNoError(hash_to_curve.Generate(conf, curveDir, bgen))

			if conf.Equal(config.GRUMPKIN) {
				return
			}

			// generate pedersen on fr
			assertNoError(pedersen.Generate(conf, filepath.Join(curveDir, "fr", "pedersen"), bgen))

			// generate tower of extension
			assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), bgen))

			// generate pairing tests
			assertNoError(pairing.Generate(conf, curveDir, bgen))

			// generate fri on fr
			assertNoError(fri.Generate(conf, filepath.Join(curveDir, "fr", "fri"), bgen))

			// generate mpc setup tools
			assertNoError(mpcsetup.Generate(conf, filepath.Join(curveDir, "mpcsetup"), bgen))

			// generate kzg on fr
			assertNoError(kzg.Generate(conf, filepath.Join(curveDir, "kzg"), bgen))

			// generate shplonk on fr
			assertNoError(shplonk.Generate(conf, filepath.Join(curveDir, "shplonk"), bgen))

			// generate fflonk on fr
			assertNoError(fflonk.Generate(conf, filepath.Join(curveDir, "fflonk"), bgen))

			// generate plookup on fr
			assertNoError(plookup.Generate(conf, filepath.Join(curveDir, "fr", "plookup"), bgen))

			// generate permutation on fr
			assertNoError(permutation.Generate(conf, filepath.Join(curveDir, "fr", "permutation"), bgen))

			// generate eddsa on companion curves
			assertNoError(fri.Generate(conf, filepath.Join(curveDir, "fr", "fri"), bgen))

		}(conf)

	}

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

	// generate the ecc_field.go file
	{
		entries := []bavard.Entry{
			{File: filepath.Join(baseDir, "ecc", "ecc_field.go"), Templates: []string{"ecc_field.go.tmpl"}},
		}

		assertNoError(bgen.Generate(config.Curves, "ecc", "./config/template", entries...))
	}

	// format the whole directory

	cmd := exec.Command("gofmt", "-s", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	cmd = exec.Command("asmfmt", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	//mathfmt doesn't accept directories. TODO: PR pending
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
