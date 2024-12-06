package element

const OpsNoAsm = `

{{- if ne .NbWords 1}}
import "math/bits"
{{- end}}

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

func fromMont(z *{{.ElementName}} ) {
	_fromMontGeneric(z)
}

func reduce(z *{{.ElementName}})  {
	_reduceGeneric(z)
}

// Mul z = x * y (mod q)
{{- if $.NoCarry}}
//
// x and y must be less than q
{{- end }}
func (z *{{.ElementName}}) Mul(x, y *{{.ElementName}}) *{{.ElementName}} {
	{{- if eq $.NbWords 1}}
		{{ template "mul_cios_one_limb" dict "all" . "V1" "x" "V2" "y" }}
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

{{- if eq $.Word.BitSize 32}}
const mu uint64 = {{.Mu}}
func barrettReduce(v uint64) uint32 {
	qPrime := ((v >> (Bits - 1)) * mu) >> (Bits + 1)
	r := v - qPrime * q
	if r >= q {
		r -= q
	}
	return uint32(r)
}

func montReduce(v uint64) uint32 {
	m := (v  * qInvNeg ) % r
	t := uint32((v + m * q) >> rBits)
	if t >= q {
		t -= q
	}
	return t
}
{{- end}}

// Square z = x * x (mod q)
{{- if $.NoCarry}}
//
// x must be less than q
{{- end }}
func (z *{{.ElementName}}) Square(x *{{.ElementName}}) *{{.ElementName}} {
	// see Mul for algorithm documentation
	{{- if eq $.NbWords 1}}
		{{ template "mul_cios_one_limb" dict "all" . "V1" "x" "V2" "x" }}
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
