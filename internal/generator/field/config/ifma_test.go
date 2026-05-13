package config

import (
	"fmt"
	"testing"
)

func TestIFMAConstants(t *testing.T) {
	// BLS12-377 fr field
	F, err := NewFieldConfig("fr", "Element", "8444461749428370424248824938781546531375899335154063827935233455917409239041", false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("=== BLS12-377 fr field IFMA constants ===")
	fmt.Printf("NbWords: %d\n", F.NbWords)
	fmt.Printf("NbBits: %d\n", F.NbBits)
	fmt.Println()

	fmt.Println("Q in radix-64:")
	for i, v := range F.Q {
		fmt.Printf("  q[%d] = 0x%016x\n", i, v)
	}
	fmt.Println()

	fmt.Println("Q in radix-52:")
	for i, v := range F.QRadix52 {
		fmt.Printf("  q52[%d] = 0x%013x\n", i, v)
	}
	fmt.Println()

	fmt.Printf("Barrett mu (shift=%d): 0x%x\n", F.BarrettShift52, F.MuBarrett52)

	// Verify against expected values from Python
	expectedQ52 := []uint64{0x1800000000001, 0xfed00000010a1, 0xc37b00159aa76, 0xa55660b44d1e5, 0x012ab655e9a2c}
	expectedMu := uint64(0x36d9)

	fmt.Println("\n=== Verification ===")
	for i, v := range F.QRadix52 {
		if v != expectedQ52[i] {
			t.Errorf("MISMATCH q52[%d]: got 0x%x, expected 0x%x", i, v, expectedQ52[i])
		}
	}
	if F.MuBarrett52 != expectedMu {
		t.Errorf("MISMATCH mu: got 0x%x, expected 0x%x", F.MuBarrett52, expectedMu)
	}
}
