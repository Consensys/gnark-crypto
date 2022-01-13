package element

const OpsNoAsm = `
// /!\ WARNING /!\
// this code has not been audited and is provided as-is. In particular, 
// there is no security guarantees such as constant time implementation 
// or side-channel attack resistance
// /!\ WARNING /!\

// TODO: Discuss why this has to be in place, and why x is not a receiver

{{ $mulConsts := list 3 5 11 13 }}
{{- range $i := $mulConsts }}

// MulBy{{$i}} x *= {{$i}}
func MulBy{{$i}}(x *{{$.ElementName}}) {
	mulByConstant(x, {{$i}})
}

{{- end}}


// Butterfly sets 
// a = a + b
// b = a - b 
func Butterfly(a, b *{{.ElementName}}) {
	_butterflyGeneric(a, b)
}

func mul(z, x, y *{{.ElementName}}) {
	_mulGeneric(z, x, y)
}


// FromMont converts z in place (i.e. mutates) from Montgomery to regular representation
// sets and returns z = z * 1
func fromMont(z *{{.ElementName}} ) {
	_fromMontGeneric(z)
}

func add(z,  x, y *{{.ElementName}}) {
	_addGeneric(z,x,y)
}

func double(z,  x *{{.ElementName}}) {
	_doubleGeneric(z,x)
}


func sub(z,  x, y *{{.ElementName}}) {
	_subGeneric(z,x,y)
}

func neg(z,  x *{{.ElementName}}) {
	_negGeneric(z,x)
}


func reduce(z *{{.ElementName}})  {
	_reduceGeneric(z)
}


`
