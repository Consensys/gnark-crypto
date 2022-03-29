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
func add(res,x,y *{{.ElementName}})

//go:noescape
func sub(res,x,y *{{.ElementName}})

//go:noescape
func neg(res,x *{{.ElementName}})

//go:noescape
func double(res,x *{{.ElementName}})

//go:noescape
func mul(res,x,y *{{.ElementName}})

//go:noescape
func fromMont(res *{{.ElementName}})

//go:noescape
func reduce(res *{{.ElementName}})

//go:noescape
func Butterfly(a, b *{{.ElementName}})

{{end}}



`

//OpsARM64 is included with ARM64 builds (regardless of architecture or if F.ASM is set)
const OpsARM64 = `

{{if .ASM}}

//go:noescape
func add(res, x, y *{{.ElementName}})

{{end}}
`
