package config

type Field struct {
	Name    string
	Modulus string
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
		Name:    "koalabear",
		Modulus: "0x7f000001",
	})
	addField(Field{
		Name:    "babybear",
		Modulus: "0x78000001",
	})
}
