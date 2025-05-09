package generator

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
	"github.com/consensys/gnark-crypto/field/generator/config"
	eccconfig "github.com/consensys/gnark-crypto/internal/generator/config"
)

func generateFFT(F *config.Field, fft *config.FFT, outputDir string) error {

	if fft.GeneratorFullMultiplicativeGroup == 0 || fft.GeneratorMaxTwoAdicSubgroup == "" {
		// try to populate ourselves
		// TODO @gbotrel right now hardcoded lookup tables, we should do better.
		data, ok := fftConfigs[F.Modulus]
		if !ok {
			return fmt.Errorf("no fft config for modulus %s", F.Modulus)
		}
		fft = &data
	}

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}
	data := &fftTemplateData{
		FFT:              *fft,
		FieldPackagePath: fieldImportPath,
		FF:               F.PackageName,
		HasASMKernel:     F.F31,
		Kernels:          []int{5, 8},
		Package:          "fft",
		F31:              F.F31,
	}
	outputDir = filepath.Join(outputDir, "fft")

	pureGoBuildTag := ""
	if data.HasASMKernel {
		pureGoBuildTag = "purego || (!amd64)"
		data.Kernels = []int{8}
	}

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain_test.go"), Templates: []string{"tests/domain.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain.go"), Templates: []string{"domain.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft_test.go"), Templates: []string{"tests/fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "bitreverse_test.go"), Templates: []string{"tests/bitreverse.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft.go"), Templates: []string{"fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "kernel_purego.go"), Templates: []string{"kernel.purego.go.tmpl"}, BuildTag: pureGoBuildTag},
		{File: filepath.Join(outputDir, "bitreverse.go"), Templates: []string{"bitreverse.go.tmpl"}},
		{File: filepath.Join(outputDir, "options.go"), Templates: []string{"options.go.tmpl"}},
	}
	if F.F31 {
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext_test.go"), Templates: []string{"tests/fftext.go.tmpl"}})
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext.go"), Templates: []string{"fftext.go.tmpl"}})
	}
	if data.HasASMKernel {
		data.Q = F.Q[0]
		data.QInvNeg = F.QInverse[0]
		entries = append(entries,
			bavard.Entry{
				File:      filepath.Join(outputDir, "kernel_amd64.go"),
				Templates: []string{"kernel.amd64.go.tmpl"},
				BuildTag:  "!purego"})

		// generate the assembly file;
		fftKernels, err := os.Create(filepath.Join(outputDir, "kernel_amd64.s"))
		if err != nil {
			return err
		}

		fftKernels.WriteString("//go:build !purego\n")

		if err := amd64.GenerateF31FFTKernels(fftKernels, F.NbBits, data.Kernels); err != nil {
			fftKernels.Close()
			return err
		}
		fftKernels.Close()
	}

	funcs := make(map[string]interface{})
	funcs["bitReverse"] = bitReverse
	funcs["reverseBits"] = func(x, n any) uint64 {
		return bits.Reverse64(anyToUint64(x)) >> anyToUint64(n)
	}
	funcs["shl"] = func(x, n any) uint64 {
		return anyToUint64(x) << anyToUint64(n)
	}
	funcs["logicalOr"] = func(x, y any) uint64 {
		return anyToUint64(x) | anyToUint64(y)
	}

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	fftTemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	fftTemplatesRootDir = filepath.Join(fftTemplatesRootDir, "fft")

	if err := bgen.GenerateWithOptions(data, "fft", fftTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	// put the generator in the parent dir (fr)
	// TODO this should be in goff
	entries = []bavard.Entry{
		{File: filepath.Join(outputDir, "../generator.go"), Templates: []string{"fr.generator.go.tmpl"}},
	}
	fieldNameSplitted := strings.Split(data.FieldPackagePath, "/")
	fieldName := fieldNameSplitted[len(fieldNameSplitted)-1]
	if err := bgen.GenerateWithOptions(data, fieldName, fftTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}

type fftTemplateData struct {
	config.FFT

	FieldPackagePath string // path to the finite field package
	FF               string // name of the package corresponding to the finite field
	HasASMKernel     bool   // indicates if the kernels have an assembly impl
	Kernels          []int  // indicates which kernels to generate
	Package          string // package name
	Q, QInvNeg       uint64
	F31              bool
}

func findTemplatesRootDir() (string, error) {
	// walks through the directory tree to find the templates root dir;
	// we find the dir with go.mod file, then go up to field/generator/internal/templates
	// this is a bit hacky, but it works

	// find the dir with go.mod file
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		dir = filepath.Dir(dir)
		if dir == "/" || dir == "." {
			return "", errors.New("could not find templates root dir")
		}
	}

	return filepath.Join(dir, "field/generator/internal/templates"), nil

}

func anyToUint64(x any) uint64 {
	switch v := x.(type) {
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	default:
		panic("unknown type")
	}
}

var fftConfigs map[string]config.FFT

func init() {
	fftConfigs = make(map[string]config.FFT)

	// TODO @gbotrel temporary
	// bls12-377
	fftConfigs[eccconfig.BLS12_377.FrModulus] = config.NewConfig(
		22,
		"8065159656716812877374967518403273466521432693661810619979959746626482506078",
		"47",
	)

	// bls12-381
	fftConfigs[eccconfig.BLS12_381.FrModulus] = (config.NewConfig(
		7,
		"10238227357739495823651030575849232062558860180284477541189508159991286009131",
		"32",
	))

	// bn254
	fftConfigs[eccconfig.BN254.FrModulus] = (config.NewConfig(
		5,
		"19103219067921713944291392827692070036145651957329286315305642004821462161904",
		"28",
	))

	// bw6-761
	fftConfigs[eccconfig.BW6_761.FrModulus] = (config.NewConfig(
		15,
		"32863578547254505029601261939868325669770508939375122462904745766352256812585773382134936404344547323199885654433",
		"46",
	))

	// bw6-633
	fftConfigs[eccconfig.BW6_633.FrModulus] = (config.NewConfig(
		13,
		"4991787701895089137426454739366935169846548798279261157172811661565882460884369603588700158257",
		"20",
	))

	// bls24-315
	fftConfigs[eccconfig.BLS24_315.FrModulus] = (config.NewConfig(
		7,
		"1792993287828780812362846131493071959406149719416102105453370749552622525216",
		"22",
	))

	// bls24-317
	fftConfigs[eccconfig.BLS24_317.FrModulus] = (config.NewConfig(
		7,
		"16532287748948254263922689505213135976137839535221842169193829039521719560631",
		"60",
	))

	// goldilocks
	fftConfigs["18446744069414584321"] = (config.NewConfig(
		7,
		"1753635133440165772",
		"32",
	))

	// koala bear
	fftConfigs["2130706433"] = (config.NewConfig(
		3,
		"1791270792",
		"24",
	))

	// baby bear
	fftConfigs["2013265921"] = (config.NewConfig(
		31,
		"440564289",
		"27",
	))

}

func bitReverse(n, i int64) uint64 {
	nn := uint64(64 - bits.TrailingZeros64(uint64(n)))
	r := make([]uint64, n)
	for i := 0; i < len(r); i++ {
		r[i] = uint64(i)
	}
	for i := 0; i < len(r); i++ {
		irev := bits.Reverse64(r[i]) >> nn
		if irev > uint64(i) {
			r[i], r[irev] = r[irev], r[i]
		}
	}
	return r[i]
}
