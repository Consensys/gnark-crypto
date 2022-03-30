package element

//TODO: Remove

const OpsNoAsmTemp = `
func add(z,  x, y *{{.ElementName}}) {
_addGeneric(z,x,y)
}

func sub(z,  x, y *{{.ElementName}}) {
	_subGeneric(z,x,y)
}

`
