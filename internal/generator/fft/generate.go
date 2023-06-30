package fft

import (
	"math/bits"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {

	conf.Package = "fft"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "domain_test.go"), Templates: []string{"tests/domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "domain.go"), Templates: []string{"domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft_test.go"), Templates: []string{"tests/fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft.go"), Templates: []string{"fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "options.go"), Templates: []string{"options.go.tmpl", "imports.go.tmpl"}},
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

	return bgen.GenerateWithOptions(conf, conf.Package, "./fft/template/", bavardOpts, entries...)
}
