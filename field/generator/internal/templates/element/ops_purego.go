package element

const OpsNoAsm = `

import "math/bits"

{{ $mulConsts := list 3 5 13 }}
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

{{- if $.F31}}
// Mul2ExpNegN multiplies x by -1/2^n
//
// Since the Montgomery constant is 2^32, the Montgomery form of 1/2^n is
// 2^{32-n}. Montgomery reduction works provided the input is < 2^32 so this
// works for 0 <= n <= 32.
//
// N.B. n must be < 33.
func (z *{{.ElementName}}) Mul2ExpNegN(x *{{.ElementName}}, n uint32) *{{.ElementName}} {
       v := uint64(x[0]) << (32 - n)
       z[0] = montReduce(v)
       return z
}
{{- end}}

func fromMont(z *{{.ElementName}} ) {
	_fromMontGeneric(z)
}

func reduce(z *{{.ElementName}})  {
	_reduceGeneric(z)
}

{{- if $.F31}}
func montReduce(v uint64) uint32 {
	const (
		R    = 1 << 32
		qInv = R - qInvNeg
	)

	m := uint32(v) * qInv
	t, borrow := bits.Sub64(v, uint64(m)*q, 0)

	res := uint32(t / R)
	if borrow != 0 {
		res += q
	}
	return res
}
{{- end}}

// Mul z = x * y (mod q)
{{- if $.NoCarry}}
//
// x and y must be less than q
{{- end }}
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	{{- if eq $.NbWords 1}}
		{{- if $.F31}}
			v := uint64(x[0]) * uint64(y[0])
			z[0] = montReduce(v)
		{{- else}}
			{{ template "mul_cios_one_limb" dict "all" . "V1" "x" "V2" "y" }}
		{{- end}}
	{{- else }}
		{{ mul_doc $.NoCarry }}
		{{- if $.NoCarry}}
			{{ template "mul_nocarry" dict "all" . "V1" "x" "V2" "y"}}
		{{- else}}
			{{ template "mul_cios" dict "all" . "V1" "x" "V2" "y" "ReturnZ" true}}
		{{- end}}
		{{ template "reduce"  . }}
	{{- end }}
	return z
}

// Square z = x * x (mod q)
{{- if $.NoCarry}}
//
// x must be less than q
{{- end }}
func (z *{{.ElementName}}) Square(x *{{.ElementName}}) *{{.ElementName}} {
	// see Mul for algorithm documentation
	{{- if eq $.NbWords 1}}
		{{- if $.F31}}
			v := uint64(x[0]) * uint64(x[0])
			z[0] = montReduce(v)
		{{- else}}
			{{ template "mul_cios_one_limb" dict "all" . "V1" "x" "V2" "x" }}
		{{- end}}
	{{- else }}
		{{- if $.NoCarry}}
			{{ template "mul_nocarry" dict "all" . "V1" "x" "V2" "x"}}
		{{- else}}
			{{ template "mul_cios" dict "all" . "V1" "x" "V2" "x" "ReturnZ" true}}
		{{- end}}
		{{ template "reduce"  . }}
	{{- end }}
	return z
}

// Butterfly sets
//  a = a + b (mod q)
//  b = a - b (mod q)
func Butterfly(a, b *{{.ElementName}}) {
	_butterflyGeneric(a, b)
}

`
