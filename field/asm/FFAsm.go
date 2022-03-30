package asm

import (
	"github.com/consensys/gnark-crypto/field"
)

func (f *FFAsm64) StackSize(maxNbRegistersNeeded, nbRegistersReserved, minStackSize int) int {
	got := f.NbRegisters - nbRegistersReserved
	r := got - maxNbRegistersNeeded
	if r >= 0 {
		return minStackSize
	}
	r *= -8
	return max(r, minStackSize)
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (f *FFAsm64) AssertCleanStack(reservedStackSize, minStackSize int) {
	if f.NbElementsOnStack != 0 {
		panic("missing f.Push stack elements")
	}
	if reservedStackSize < minStackSize {
		panic("invalid minStackSize or reservedStackSize")
	}
	usedStackSize := f.MaxOnStack * 8
	if usedStackSize > reservedStackSize {
		panic("using more stack size than reserved")
	} else if max(usedStackSize, minStackSize) < reservedStackSize {
		// this panic is for dev purposes as this may be by design for alignment
		panic("reserved more stack size than needed")
	}

	f.MaxOnStack = 0
}

type FFAsm64 struct {
	*field.Field
	NbElementsOnStack int
	MaxOnStack        int
	NbRegisters       int
}
