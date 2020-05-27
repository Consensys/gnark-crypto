package primefields

const TwoInv = `
import "github.com/consensys/gurvy/{{.Fpackage}}/fp"

// InitTwoInv set z.TwoInv to the inverse of 2 as an fp.Element
func (z *GenerateData) InitTwoInv() *GenerateData {
	var twoInv fp.Element
	twoInv.SetUint64(2).Inverse(&twoInv)
	z.TwoInv = twoInv[:]
	return z
}
`
