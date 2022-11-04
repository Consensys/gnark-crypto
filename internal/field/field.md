
# Usage

At the root of your repo:
```bash
go get github.com/consensys/gnark-crypto/field
``` 

then in a `main.go`  (that can be called using a `go:generate` workflow):

```
generator.GenerateFF(packageName, structName, modulus, destinationPath, false)
```

The generated type has an API that's similar with `big.Int`

Example API signature
```go 
// Mul z = x * y mod q
func (z *Element) Mul(x, y *Element) *Element 
```

and can be used like so:

```go 
var a, b Element
a.SetUint64(2)
b.SetString("984896738")

a.Mul(a, b)

a.Sub(a, a)
 .Add(a, b)
 .Inv(a)
 
b.Exp(b, 42)
b.Neg(b)
```

