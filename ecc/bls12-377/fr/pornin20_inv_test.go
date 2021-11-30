package fr

import "testing"

func testInv(x Element) {
	var oldInv Element
	oldInv.InverseOld(&x)
	var inv Element
	inv.Inverse(&x)

	if !inv.Equal(&oldInv) {
		panic("mismatch")
	}
}

func TestInv0(t *testing.T) {
	testInv( Element{
		10939534727055711396,
		1205464551661532186,
		1416111280771473410,
		727939882724427866,
	})
}
