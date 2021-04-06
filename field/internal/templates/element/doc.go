package element

const Doc = `
// Package {{.PackageName}} contains field arithmetic operations for modulus = 0x{{shorten .ModulusHex}}.
// 
// The API is similar to math/big (big.Int), but the operations are significantly faster (up to 20x for the modular multiplication on amd64, see also https://hackmd.io/@zkteam/modular_multiplication)
// 
// The modulus is hardcoded in all the operations.
// 
// Field elements are represented as an array, and assumed to be in Montgomery form in all methods:
// 	type {{.ElementName}} [{{.NbWords}}]uint64
//
// Example API signature
// 	// Mul z = x * y mod q
// 	func (z *Element) Mul(x, y *Element) *Element
//
// and can be used like so:
// 	var a, b Element
// 	a.SetUint64(2)
// 	b.SetString("984896738")
// 	a.Mul(a, b)
// 	a.Sub(a, a)
// 	 .Add(a, b)
// 	 .Inv(a)
// 	b.Exp(b, new(big.Int).SetUint64(42))
//
// Modulus
// 	0x{{.ModulusHex}} // base 16
// 	{{.Modulus}} // base 10
package {{.PackageName}}
`
