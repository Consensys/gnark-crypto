package generator

import (
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/fft/config"

	"github.com/consensys/bavard"
)

func Generate(conf config.FFT, outputDir string, bgen *bavard.BatchGenerator) error {

	conf.Package = "fft"

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

	if err := bgen.GenerateWithOptions(conf, conf.Package, "./template/", bavardOpts, entries...); err != nil {
		return err
	}

	// put the generator in the parent dir (fr)
	// TODO this should be in goff
	entries = []bavard.Entry{
		{File: filepath.Join(outputDir, "../generator.go"), Templates: []string{"fr.generator.go.tmpl"}},
	}
	fieldNameSplitted := strings.Split(conf.FieldPackagePath, "/")
	fieldName := fieldNameSplitted[len(fieldNameSplitted)-1]
	err := bgen.GenerateWithOptions(conf, fieldName, "./template/", bavardOpts, entries...)
	if err != nil {
		return err
	}

	// format the generated code
	cmd := exec.Command("gofmt", "-s", "-w", outputDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
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
