package element

// OpsAMD64 is included with AMD64 builds (regardless of architecture or if F.ASM is set)
const OpsAMD64 = `

import (
	_ "{{.ASMPackagePath}}"
	"github.com/consensys/gnark-crypto/utils/cpu"
)

var supportAdx = cpu.SupportADX

//go:noescape
func MulBy3(x *{{.ElementName}})

//go:noescape
func MulBy5(x *{{.ElementName}})

//go:noescape
func MulBy13(x *{{.ElementName}})

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

`

const OpsARM64 = `
import (
	_ "{{.ASMPackagePath}}"
)


// Butterfly sets
//  a = a + b (mod q)
//  b = a - b (mod q)
{{- if le .NbWords 6}}
//go:noescape
func Butterfly(a, b *{{.ElementName}})
{{- else}}
func Butterfly(a, b *{{.ElementName}}) {
	_butterflyGeneric(a, b)
}
{{- end}}

//go:noescape
func mul(res,x,y *{{.ElementName}})

// Mul z = x * y (mod q)
//
// x and y must be less than q
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	mul(z,x,y)
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


{{ $mulConsts := list 3 5 11 13 }}
{{- range $i := $mulConsts }}

// MulBy{{$i}} x *= {{$i}} (mod q)
func MulBy{{$i}}(x *{{$.ElementName}}) {
	{{- if eq 1 $.NbWords}}
	var y {{$.ElementName}}
	y.SetUint64({{$i}})
	x.Mul(x, &y)
	{{- else}}
		{{- if eq $i 3}}
			_x := *x
			x.Double(x).Add(x, &_x)
		{{- else if eq $i 5}}
			_x := *x
			x.Double(x).Double(x).Add(x, &_x)
		{{- else if eq $i 11}}
			var y = {{$.ElementName}}{
				{{- range $i := $.Eleven}}
				{{$i}},{{end}}
			}
			x.Mul(x, &y)
		{{- else if eq $i 13}}
			var y = {{$.ElementName}}{
				{{- range $i := $.Thirteen}}
				{{$i}},{{end}}
			}
			x.Mul(x, &y)
		{{- else }}
			NOT IMPLEMENTED
		{{- end}}
	{{- end}}
}

{{- end}}

func fromMont(z *{{.ElementName}} ) {
	_fromMontGeneric(z)
}

//go:noescape
func reduce(res *{{.ElementName}})
`

const IncludeASM = `

// We include the hash to force the Go compiler to recompile: {{.Hash}}
#include "{{.IncludePath}}"

`
