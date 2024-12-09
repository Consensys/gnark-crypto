package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator"
	field "github.com/consensys/gnark-crypto/field/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/mimc"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/poseidon2"
	"github.com/consensys/gnark-crypto/internal/generator/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/ecdsa"
	"github.com/consensys/gnark-crypto/internal/generator/edwards"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa"
	"github.com/consensys/gnark-crypto/internal/generator/fflonk"
	"github.com/consensys/gnark-crypto/internal/generator/fft"
	fri "github.com/consensys/gnark-crypto/internal/generator/fri/template"
	"github.com/consensys/gnark-crypto/internal/generator/gkr"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_field"
	"github.com/consensys/gnark-crypto/internal/generator/iop"
	"github.com/consensys/gnark-crypto/internal/generator/kzg"
	"github.com/consensys/gnark-crypto/internal/generator/pairing"
	"github.com/consensys/gnark-crypto/internal/generator/pedersen"
	"github.com/consensys/gnark-crypto/internal/generator/permutation"
	"github.com/consensys/gnark-crypto/internal/generator/plookup"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/shplonk"
	"github.com/consensys/gnark-crypto/internal/generator/sis"
	"github.com/consensys/gnark-crypto/internal/generator/sumcheck"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils"
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

	// generate common assembly files depending on field number of words
	assertNoError(generator.GenerateAMD64(4, asmDirBuildPath, true))
	assertNoError(generator.GenerateAMD64(5, asmDirBuildPath, false))
	assertNoError(generator.GenerateAMD64(6, asmDirBuildPath, false))
	assertNoError(generator.GenerateAMD64(10, asmDirBuildPath, false))
	assertNoError(generator.GenerateAMD64(12, asmDirBuildPath, false))

	assertNoError(generator.GenerateARM64(4, asmDirBuildPath, false))
	assertNoError(generator.GenerateARM64(6, asmDirBuildPath, false))
	assertNoError(generator.GenerateARM64(10, asmDirBuildPath, false))
	assertNoError(generator.GenerateARM64(12, asmDirBuildPath, false))

	var wg sync.WaitGroup
	for _, conf := range config.Curves {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf config.Curve) {
			defer wg.Done()

			var err error

			conf.Fp, err = field.NewFieldConfig("fp", "Element", conf.FpModulus, fmt.Sprintf("fp_%s", conf.Name), true)
			assertNoError(err)

			conf.Fr, err = field.NewFieldConfig("fr", "Element", conf.FrModulus, fmt.Sprintf("fr_%s", conf.Name), !conf.Equal(config.STARK_CURVE))
			assertNoError(err)

			curveDir := filepath.Join(baseDir, "ecc", conf.Name)

			conf.FpUnusedBits = 64 - (conf.Fp.NbBits % 64)

			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(curveDir, "fr"), asmDirBuildPath, asmDirIncludePath))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(curveDir, "fp"), asmDirBuildPath, asmDirIncludePath))

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

			// generate tower of extension
			assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), bgen))

			// generate pairing tests
			assertNoError(pairing.Generate(conf, curveDir, bgen))

			// generate fri on fr
			assertNoError(fri.Generate(conf, filepath.Join(curveDir, "fr", "fri"), bgen))

			// generate fft on fr
			// TODO those ifs should disappear, the gen should be in fieldConfig, and an API in the ff pacakge should give the generator
			var fftConfig fft.FFTConfig
			if conf.Equal(config.BN254) {
				fftConfig = fft.NewFFTConfig("5", "19103219067921713944291392827692070036145651957329286315305642004821462161904", "28", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BLS12_377) {
				fftConfig = fft.NewFFTConfig("22", "8065159656716812877374967518403273466521432693661810619979959746626482506078", "47", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BLS12_381) {
				fftConfig = fft.NewFFTConfig("7", "10238227357739495823651030575849232062558860180284477541189508159991286009131", "32", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BLS24_315) {
				fftConfig = fft.NewFFTConfig("7", "1792993287828780812362846131493071959406149719416102105453370749552622525216", "22", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BLS24_317) {
				fftConfig = fft.NewFFTConfig("7", "16532287748948254263922689505213135976137839535221842169193829039521719560631", "60", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BW6_633) {
				fftConfig = fft.NewFFTConfig("13", "4991787701895089137426454739366935169846548798279261157172811661565882460884369603588700158257", "20", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			if conf.Equal(config.BW6_761) {
				fftConfig = fft.NewFFTConfig("15", "32863578547254505029601261939868325669770508939375122462904745766352256812585773382134936404344547323199885654433", "46", "github.com/consensys/gnark-crypto/ecc/"+conf.Name+"/fr")
			}
			assertNoError(fft.Generate(fftConfig, filepath.Join(curveDir, "fr", "fft"), bgen))

			if conf.Equal(config.BN254) || conf.Equal(config.BLS12_377) {
				assertNoError(sis.Generate(conf, filepath.Join(curveDir, "fr", "sis"), bgen))
			}

			// generate kzg on fr
			assertNoError(kzg.Generate(conf, filepath.Join(curveDir, "kzg"), bgen))

			// generate shplonk on fr
			assertNoError(shplonk.Generate(conf, filepath.Join(curveDir, "shplonk"), bgen))

			// generate fflonk on fr
			assertNoError(fflonk.Generate(conf, filepath.Join(curveDir, "fflonk"), bgen))

			// generate pedersen on fr
			assertNoError(pedersen.Generate(conf, filepath.Join(curveDir, "fr", "pedersen"), bgen))

			// generate plookup on fr
			assertNoError(plookup.Generate(conf, filepath.Join(curveDir, "fr", "plookup"), bgen))

			// generate permutation on fr
			assertNoError(permutation.Generate(conf, filepath.Join(curveDir, "fr", "permutation"), bgen))

			// generate mimc on fr
			assertNoError(mimc.Generate(conf, filepath.Join(curveDir, "fr", "mimc"), bgen))

			// generate poseidon2 on fr
			assertNoError(poseidon2.Generate(conf, filepath.Join(curveDir, "fr", "poseidon2"), bgen))

			frInfo := config.FieldDependency{
				FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fr",
				FieldPackageName: "fr",
				ElementType:      "fr.Element",
			}

			// generate polynomial on fr
			assertNoError(polynomial.Generate(frInfo, filepath.Join(curveDir, "fr", "polynomial"), true, bgen))

			// generate eddsa on companion curves
			assertNoError(fri.Generate(conf, filepath.Join(curveDir, "fr", "fri"), bgen))

			// generate sumcheck on fr
			assertNoError(sumcheck.Generate(frInfo, filepath.Join(curveDir, "fr", "sumcheck"), bgen))

			// generate gkr on fr
			assertNoError(gkr.Generate(gkr.Config{
				FieldDependency:         frInfo,
				GenerateTests:           true,
				TestVectorsRelativePath: "../../../../internal/generator/gkr/test_vectors",
			}, filepath.Join(curveDir, "fr", "gkr"), bgen))

			// generate test vector utils on fr
			assertNoError(test_vector_utils.Generate(test_vector_utils.Config{
				FieldDependency:             frInfo,
				RandomizeMissingHashEntries: false,
			}, filepath.Join(curveDir, "fr", "test_vector_utils"), bgen))

			// generate iop functions
			assertNoError(iop.Generate(conf, filepath.Join(curveDir, "fr", "iop"), bgen))

			fpInfo := config.FieldDependency{
				FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fp",
				FieldPackageName: "fp",
				ElementType:      "fp.Element",
			}

			// generate wrapped hash-to-field for both fr and fp
			assertNoError(hash_to_field.Generate(frInfo, filepath.Join(curveDir, "fr", "hash_to_field"), bgen))
			assertNoError(hash_to_field.Generate(fpInfo, filepath.Join(curveDir, "fp", "hash_to_field"), bgen))

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertNoError(test_vector_utils.GenerateRationals(bgen))
	}()
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

	//mathfmt doesn't accept directories. TODO: PR pending
	/*cmd = exec.Command("mathfmt", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())*/

	wg.Add(2)

	go func() {
		// generate test vectors for sumcheck
		cmd := exec.Command("go", "run", "./sumcheck/test_vectors")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		assertNoError(cmd.Run())
		wg.Done()
	}()

	go func() {
		// generate test vectors for gkr
		cmd := exec.Command("go", "run", "./gkr/test_vectors")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		assertNoError(cmd.Run())
		wg.Done()
	}()

	wg.Wait()
}

func assertNoError(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}
}
