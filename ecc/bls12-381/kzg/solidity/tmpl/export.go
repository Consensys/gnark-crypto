package tmpl

import (
	"fmt"
	"html/template"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	kzg_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/kzg"
)

func ExportContract(srs *kzg_bls12381.SRS, w io.Writer) error {
	funcMap := template.FuncMap{
		"hex": func(i int) string {
			return fmt.Sprintf("0x%x", i)
		},
		"frstr": func(x fr.Element) string {
			// we use big.Int to always get a positive string.
			// not the most efficient hack, but it works better for .sol generation.
			bv := new(big.Int)
			x.BigInt(bv)
			return bv.String()
		},
		"fpstr": func(x fp.Element) string {
			bv := new(big.Int)
			x.BigInt(bv)
			return bv.String()
		},
		"add": func(i, j int) int {
			return i + j
		},
		"neg": func(x fp.Element) string {
			bp := fp.Modulus()
			var bx big.Int
			x.BigInt(&bx)
			bp.Sub(bp, &bx)
			return bp.String()
		},
	}

	t, err := template.New("t").Funcs(funcMap).Parse(SolidityKzg)
	if err != nil {
		return err
	}

	return t.Execute(w, struct {
		Vk kzg_bls12381.VerifyingKey
	}{
		Vk: srs.Vk,
	})
}
