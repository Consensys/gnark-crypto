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

{{- if .ASMVector}}
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

// Sub subtracts two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Sub(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Sub: vectors don't have the same length")
	}
	subVec(&(*vector)[0], &a[0], &b[0], uint64(len(a)))
}

//go:noescape
func subVec(res, a, b *{{.ElementName}}, n uint64)

// ScalarMul multiplies a vector by a scalar element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) ScalarMul(a Vector, b *{{.ElementName}}) {
	if len(a) != len(*vector) {
		panic("vector.ScalarMul: vectors don't have the same length")
	}
	scalarMulVec(&(*vector)[0], &a[0], b, uint64(len(a)))
}

//go:noescape
func scalarMulVec(res, a, b *{{.ElementName}}, n uint64)

// Sum computes the sum of all elements in the vector.
func (vector *Vector) Sum() (res {{.ElementName}}) {
	if len(*vector) == 0 {
		return
	}
	// n := uint64(len(*vector))
	// const minN = 16*7 // AVX512 slower than generic for small n
	// const maxN = (1 << 32) - 1
	// if !supportAvx512 || n <= minN || n >= maxN {
	// 	// call sumVecGeneric
	// 	sumVecGeneric(&res, *vector)
	// 	return
	// }
	sumVec(&res, &(*vector)[0], uint64(len(*vector)))
	return
}

//go:noescape
func sumVec(res *{{.ElementName}}, a *{{.ElementName}}, n uint64)


//go:noescape
func innerProdVec(res *{{.ElementName}}, a,b *{{.ElementName}}, n uint64)

// InnerProduct computes the inner product of two vectors.
// It panics if the vectors don't have the same length.
func (vector *Vector) InnerProduct(other Vector) (res {{.ElementName}}) {
if len(other) == 0 {
return
}
	// n := uint64(len(*vector))
	// if n != uint64(len(other)) {
	// 	panic("vector.InnerProduct: vectors don't have the same length")
	// }
	// const minN = 16*7 // AVX512 slower than generic for small n
	// const maxN = (1 << 32) - 1
	// if !supportAvx512 || n <= minN || n >= maxN {
	// 	// call innerProductVecGeneric
	// 	innerProductVecGeneric(&res, *vector, other)
	// 	return
	// }
	innerProdVec(&res, &(*vector)[0], &other[0], uint64(len(*vector)))

	return
}

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
