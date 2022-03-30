package arm64

import (
	"fmt"
	"github.com/consensys/bavard"
	"github.com/consensys/bavard/arm64"
	"github.com/consensys/gnark-crypto/field/asm"
	"io"

	"github.com/consensys/gnark-crypto/field"
)

func NewFFArm64(w io.Writer, F *field.Field) *FFArm64 {
	return &FFArm64{asm.FFAsm64{Field: F}, arm64.NewArm64(w)}
}

type FFArm64 struct {
	asm.FFAsm64
	*arm64.Arm64 //TODO: Introduce an Arm64 type

}

// Generate generates assembly code for the base field provided to goff
// see internal/templates/ops*
func Generate(w io.Writer, F *field.Field) error {
	f := NewFFArm64(w, F)
	f.WriteLn(bavard.Apache2Header("ConsenSys Software Inc.", 2020))

	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"\n")

	//f.GenerateDefines()
	f.generateGlobals()

	// add
	//TODO: It requires field size < 960
	f.generateAdd()

	// sub
	f.generateSub()

	// double
	f.generateDouble()
	/*
		// neg
		f.generateNeg()

		// reduce
		f.generateReduce()

		// mul by constants
		f.generateMulBy3()
		f.generateMulBy5()
		f.generateMulBy13()

		// fft butterflies
		f.generateButterfly()*/

	return nil
}

// Copy pasted from AMD64. TODO: Merge?
func (f *FFArm64) generateGlobals() {

	f.Comment("modulus q")

	for i, w := range f.Q {
		f.WriteLn(fmt.Sprintf("DATA q<>+%d(SB)/8, $%d", 8*i, w))
	}
	f.WriteLn(fmt.Sprintf("GLOBL q<>(SB), (RODATA+NOPTR), $%d", 8*f.NbWords))

	f.Comment("qInv0 q'[0]")
	f.WriteLn(fmt.Sprintf("DATA qInv0<>(SB)/8, $%d", f.QInverse[0]))
	f.WriteLn("GLOBL qInv0<>(SB), (RODATA+NOPTR), $8")

}
