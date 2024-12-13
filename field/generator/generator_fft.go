package generator

import (
	"errors"
	"math/bits"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
	eccconfig "github.com/consensys/gnark-crypto/internal/generator/config"
)

func generateFFT(F *config.Field, fft *config.FFT, outputDir string) error {
	outputDir = filepath.Join(outputDir, "fft")

	*fft = fftConfigs[F.Modulus]
	fft.Package = "fft"

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain_test.go"), Templates: []string{"tests/domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain.go"), Templates: []string{"domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft_test.go"), Templates: []string{"tests/fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "bitreverse_test.go"), Templates: []string{"tests/bitreverse.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft.go"), Templates: []string{"fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "bitreverse.go"), Templates: []string{"bitreverse.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(outputDir, "options.go"), Templates: []string{"options.go.tmpl", "imports.go.tmpl"}},
	}

	funcs := make(map[string]interface{})
	funcs["bitReverse"] = func(n, i int64) uint64 {
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

	if err := bgen.GenerateWithOptions(fft, fft.Package, fftTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	// put the generator in the parent dir (fr)
	// TODO this should be in goff
	entries = []bavard.Entry{
		{File: filepath.Join(outputDir, "../generator.go"), Templates: []string{"fr.generator.go.tmpl"}},
	}
	fieldNameSplitted := strings.Split(fft.FieldPackagePath, "/")
	fieldName := fieldNameSplitted[len(fieldNameSplitted)-1]
	if err := bgen.GenerateWithOptions(fft, fieldName, fftTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
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

var Configs []config.FFT

var fftConfigs map[string]config.FFT

func init() {
	fftConfigs = make(map[string]config.FFT)

	// TODO @gbotrel temporary
	// bls12-377
	fftConfigs[eccconfig.BLS12_377.FrModulus] = (config.NewConfig(

		"22",
		"8065159656716812877374967518403273466521432693661810619979959746626482506078",
		"47",
		"github.com/consensys/gnark-crypto/ecc/bls12-377/fr",
	))

	// bls12-381
	fftConfigs[eccconfig.BLS12_381.FrModulus] = (config.NewConfig(

		"7",
		"10238227357739495823651030575849232062558860180284477541189508159991286009131",
		"32",
		"github.com/consensys/gnark-crypto/ecc/bls12-381/fr",
	))

	// bn254
	fftConfigs[eccconfig.BN254.FrModulus] = (config.NewConfig(

		"5",
		"19103219067921713944291392827692070036145651957329286315305642004821462161904",
		"28",
		"github.com/consensys/gnark-crypto/ecc/bn254/fr",
	))

	// bw6-761
	fftConfigs[eccconfig.BW6_761.FrModulus] = (config.NewConfig(

		"15",
		"32863578547254505029601261939868325669770508939375122462904745766352256812585773382134936404344547323199885654433",
		"46",
		"github.com/consensys/gnark-crypto/ecc/bw6-761/fr",
	))

	// bw6-633
	fftConfigs[eccconfig.BW6_633.FrModulus] = (config.NewConfig(

		"13",
		"4991787701895089137426454739366935169846548798279261157172811661565882460884369603588700158257",
		"20",
		"github.com/consensys/gnark-crypto/ecc/bw6-633/fr",
	))

	// bls24-315
	fftConfigs[eccconfig.BLS24_315.FrModulus] = (config.NewConfig(

		"7",
		"1792993287828780812362846131493071959406149719416102105453370749552622525216",
		"22",
		"github.com/consensys/gnark-crypto/ecc/bls24-315/fr",
	))

	// bls24-317
	fftConfigs[eccconfig.BLS24_317.FrModulus] = (config.NewConfig(

		"7",
		"16532287748948254263922689505213135976137839535221842169193829039521719560631",
		"60",
		"github.com/consensys/gnark-crypto/ecc/bls24-317/fr",
	))

	// goldilocks
	fftConfigs["18446744069414584321"] = (config.NewConfig(

		"7",
		"1753635133440165772",
		"32",
		"github.com/consensys/gnark-crypto/field/goldilocks",
	))

	// koala bear
	fftConfigs["2130706433"] = (config.NewConfig(

		"3",
		"1791270792",
		"24",
		"github.com/consensys/gnark-crypto/field/koalabear",
	))

	// baby bear
	fftConfigs["2013265921"] = (config.NewConfig(

		"31",
		"440564289",
		"27",
		"github.com/consensys/gnark-crypto/field/babybear",
	))

}
