package element

const Doc = `
// Package {{.PackageName}} contains field arithmetic operations for modulus = 0x{{shorten .ModulusHex}}.
// 
// The API is similar to math/big (big.Int), but the operations are significantly faster (up to 20x). 
//
// Additionally {{.PackageName}}.Vector offers an API to manipulate []{{.ElementName}}{{- if .GenerateVectorOpsAMD64}} using AVX512{{- if .GenerateVectorOpsARM64}}/NEON{{- end}} instructions if available{{- end}}.
//
// The modulus is hardcoded in all the operations.
// 
// Field elements are represented as an array, and assumed to be in Montgomery form in all methods:
// 	type {{.ElementName}} [{{.NbWords}}]{{.Word.TypeLower}}
//
// Usage
//
// Example API signature:
// 	// Mul z = x * y (mod q)
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
// Modulus q =
//
// 	q[base10] = {{.Modulus}}
// 	q[base16] = 0x{{.ModulusHex}}
//
// Warning
//
// There is no security guarantees such as constant time implementation or side-channel attack resistance.
// This code is provided as-is. Partially audited, see https://github.com/Consensys/gnark/tree/master/audits
// for more details.
package {{.PackageName}}
`
