package generator

import (
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/sis/config"
)

func Generate(conf config.SIS, outputDir string, bgen *bavard.BatchGenerator) error {

	conf.Package = "sis"

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "sis_fft.go"), Templates: []string{"fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "sis.go"), Templates: []string{"sis.go.tmpl"}},
		{File: filepath.Join(outputDir, "sis_test.go"), Templates: []string{"sis.test.go.tmpl"}},
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

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

	err := bgen.GenerateWithOptions(conf, conf.Package, "./template/", bavardOpts, entries...)
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
