package config

type Field struct {
	Name    string
	Modulus string

	// GenerateExtensionE8 enables generation of field/extensions/e8.go.
	GenerateExtensionE8 bool

	// CustomExtensionCbrt keeps extension Cbrt methods in hand-maintained files.
	CustomExtensionCbrt bool
}

var Fields []Field

func addField(f Field) {
	Fields = append(Fields, f)
}

func init() {
	addField(Field{
		Name:    "goldilocks",
		Modulus: "0xFFFFFFFF00000001",
	})
	addField(Field{
		Name:                "koalabear",
		Modulus:             "0x7f000001",
		GenerateExtensionE8: true,
		CustomExtensionCbrt: true,
	})
	addField(Field{
		Name:    "babybear",
		Modulus: "0x78000001",
	})
}
