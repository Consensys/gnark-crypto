package fft

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/stretchr/testify/require"
)

//go:noescape
func tbed2(a, b, res []uint32)

func TestTwiddles(t *testing.T) {
	t.Skip("skipping test for now")
	assert := require.New(t)

	// 0 to 15
	tt := []uint32{1, 2}
	expected := []uint32{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	res := make([]uint32, 16)
	tbed(tt, res)

	assert.Equal(expected, res)

	// ok let's say now each pair of vector v0 v1
	// such that
	// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
	// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]
	// with a0 ... a15 = 0...15 and b0 ... b15 = 16...31
	// we want to reconstruct res
	// res = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15 | b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	a := []uint32{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30}
	b := []uint32{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}
	res = make([]uint32, 32)
	tbed2(a, b, res)

	expected = []uint32{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23,
		24, 25, 26, 27, 28, 29, 30, 31,
	}

	fmt.Println("a")
	for i := 0; i < 16; i++ {
		fmt.Printf("%2d ", a[i])
	}
	fmt.Println()
	fmt.Println("b")
	for i := 0; i < 16; i++ {
		fmt.Printf("%2d ", b[i])
	}
	fmt.Println()
	fmt.Println("res")
	for i := 0; i < 16; i++ {
		fmt.Printf("%2d ", res[i])
	}
	fmt.Println()
	for i := 16; i < 32; i++ {
		fmt.Printf("%2d ", res[i])
	}
	fmt.Println()

	// assert.Equal(expected, res)

}

//go:noescape
func permuteID(a, b []uint32)

func TestPermuteID(t *testing.T) {
	assert := require.New(t)
	a := []uint32{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30}
	b := []uint32{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}

	expectedA := []uint32{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30}
	expectedB := []uint32{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}

	permuteID(a, b)

	assert.Equal(expectedA, a)
	assert.Equal(expectedB, b)

}

func TestFFTDITAVX(t *testing.T) {
	assert := require.New(t)

	// 0 to 256
	a := make([]koalabear.Element, 256)
	b := make([]koalabear.Element, 256)
	for i := range a {
		a[i][0] = uint32(i)
		b[i][0] = uint32(i)
	}

	// new domain
	domain := NewDomain(256)

	kerDITNP_256generic(a, domain.twiddles, 0)
	kerDITNP_256_avx512(b, domain.twiddles, 0)

	for i := range a {
		assert.Equal(a[i][0], b[i][0])
	}
}

func TestSisShuffle(t *testing.T) {
	assert := require.New(t)

	// 0 to 256
	a := make([]koalabear.Element, 512)
	b := make([]koalabear.Element, 512)
	for i := range a {
		a[i][0] = uint32(i)
		b[i][0] = uint32(i)
	}

	// new domain
	SISShuffle(a)
	SISUnshuffle(a)

	for i := range a {
		assert.Equal(a[i][0], b[i][0])
	}
}
