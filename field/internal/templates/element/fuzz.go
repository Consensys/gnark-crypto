package element

// Fuzz when build tag is provided, we expose Generic methods to be used by ecc/ package fuzzing functions
const Fuzz = `

// MulGeneric is a wrapper exposed and used for fuzzing purposes only
func MulGeneric(z,x,y *{{.ElementName}}) {
	_mulGeneric(z, x, y)
}

// FromMontGeneric is a wrapper exposed and used for fuzzing purposes only
func FromMontGeneric(z *{{.ElementName}}) {
	_fromMontGeneric(z)
}

// AddGeneric is a wrapper exposed and used for fuzzing purposes only
func AddGeneric(z,  x, y *{{.ElementName}}) {
	_addGeneric(z, x,y)
}

// DoubleGeneric is a wrapper exposed and used for fuzzing purposes only
func DoubleGeneric(z,  x *{{.ElementName}}) {
	_doubleGeneric(z, x)
}

// SubGeneric is a wrapper exposed and used for fuzzing purposes only
func SubGeneric(z,  x, y *{{.ElementName}}) {
	_subGeneric(z,x,y)
}

// NegGeneric is a wrapper exposed and used for fuzzing purposes only
func NegGeneric(z,  x *{{.ElementName}}) {
	_negGeneric(z,x)
}

// ReduceGeneric is a wrapper exposed and used for fuzzing purposes only
func ReduceGeneric(z *{{.ElementName}})  {
	_reduceGeneric(z)
}


`
