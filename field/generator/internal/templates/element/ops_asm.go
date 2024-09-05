package element

// OpsAMD64 is included with AMD64 builds (regardless of architecture or if F.ASM is set)
const OpsAMD64 = `

{{if .ASM}}

//go:noescape
func MulBy3(x *{{.ElementName}})

//go:noescape
func MulBy5(x *{{.ElementName}})

//go:noescape
func MulBy13(x *{{.ElementName}})

//go:noescape
func mul(res,x,y *{{.ElementName}})

//go:noescape
func fromMont(res *{{.ElementName}})

//go:noescape
func reduce(res *{{.ElementName}})

// Butterfly sets
//  a = a + b (mod q)
//  b = a - b (mod q)
//go:noescape
func Butterfly(a, b *{{.ElementName}})

{{- if eq .NbWords 4}}
// Add adds two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Add(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Add: vectors don't have the same length")
	}
	addVec(&(*vector)[0], &a[0], &b[0], uint64(len(a)))
}

//go:noescape
func addVec(res, a, b *{{.ElementName}}, n uint64)
{{- end}}

// Mul z = x * y (mod q)
//
// x and y must be less than q
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	{{ mul_doc $.NoCarry }}
	mul(z, x, y)
	return z
}

// Square z = x * x (mod q)
//
// x must be less than q
func (z *{{.ElementName}}) Square(x *{{.ElementName}}) *{{.ElementName}} {
	// see Mul for doc.
	mul(z, x, x)
	return z
}

{{end}}



`
