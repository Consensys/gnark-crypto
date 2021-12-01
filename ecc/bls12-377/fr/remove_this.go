package fr

import (
	"fmt"
	"testing"
)

func testInv(x *Element) {
	var oldInv Element
	oldInv.InverseOld(x)
	var inv Element
	inv.Inverse(x)

	if !inv.Equal(&oldInv) {
		var ratio Element
		ratio.Mul(x, &inv)
		fmt.Println("off by", ratio)
		ratio.Mul(x, &oldInv)
		fmt.Println(ratio, "is one")
		panic("mismatch")
	}
}

func TestInv0(t *testing.T) {
	testInv(&Element{
		10939534727055711396,
		1205464551661532186,
		1416111280771473410,
		727939882724427866,
	})
}

func TestInv(t *testing.T) {
	var x Element
	for i := 0; i < 1000; i++ {
		x.SetRandom()
		//fmt.Println("trying", i, x)
		testInv(&x)
	}
}
