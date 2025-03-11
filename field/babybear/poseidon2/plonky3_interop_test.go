package poseidon2

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"testing"

	fr "github.com/consensys/gnark-crypto/field/babybear"
	"github.com/stretchr/testify/require"
)

func TestPlonky3Interop(t *testing.T) {
	assert := require.New(t)

	h := NewPermutation(16, 8, 13)
	h.setHorizenRoundKeys()

	// read poseidon2_babybear_16_test_vectors.csv
	// it is structured;
	// 1 line of header
	// 1line = 32 uint64; 16 for input, 16 for expected result
	// this file was generated using plonky3 babybear implementation at commit de2b3b7
	// using default_babybear_poseidon2_16() and random inputs

	file, err := os.Open("poseidon2_babybear_16_test_vectors.csv")
	assert.NoError(err)
	r := csv.NewReader(file)
	r.Read() // skip header
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		assert.NoError(err)
		assert.Equal(len(record), 32)
		var input, expected [16]fr.Element
		for i := 0; i < 16; i++ {
			v, err := strconv.Atoi(record[i])
			assert.NoError(err)

			input[i].SetUint64(uint64(v))
		}
		for i := 16; i < 32; i++ {
			v, err := strconv.Atoi(record[i])
			assert.NoError(err)
			expected[i-16].SetUint64(uint64(v))
		}

		h.Permutation(input[:])

		for i := 0; i < 16; i++ {
			assert.Equal(input[i], expected[i])
		}
	}

}

func (h *Permutation) setHorizenRoundKeys() {
	// See https://github.com/HorizenLabs/poseidon2/blob/main/plain_implementations/src/poseidon2/poseidon2_instance_babybear.rs
	// and https://github.com/Plonky3/Plonky3/blob/de2b3b788660ef225af6d9d9f6df8c03b1001e05/baby-bear/src/poseidon2.rs#L115
	externalInitial := [][]uint64{
		{
			0x69cbb6af, 0x46ad93f9, 0x60a00f4e, 0x6b1297cd, 0x23189afe, 0x732e7bef, 0x72c246de,
			0x2c941900, 0x0557eede, 0x1580496f, 0x3a3ea77b, 0x54f3f271, 0x0f49b029, 0x47872fe1,
			0x221e2e36, 0x1ab7202e,
		},
		{
			0x487779a6, 0x3851c9d8, 0x38dc17c0, 0x209f8849, 0x268dcee8, 0x350c48da, 0x5b9ad32e,
			0x0523272b, 0x3f89055b, 0x01e894b2, 0x13ddedde, 0x1b2ef334, 0x7507d8b4, 0x6ceeb94e,
			0x52eb6ba2, 0x50642905,
		},
		{
			0x05453f3f, 0x06349efc, 0x6922787c, 0x04bfff9c, 0x768c714a, 0x3e9ff21a, 0x15737c9c,
			0x2229c807, 0x0d47f88c, 0x097e0ecc, 0x27eadba0, 0x2d7d29e4, 0x3502aaa0, 0x0f475fd7,
			0x29fbda49, 0x018afffd,
		},
		{
			0x0315b618, 0x6d4497d1, 0x1b171d9e, 0x52861abd, 0x2e5d0501, 0x3ec8646c, 0x6e5f250a,
			0x148ae8e6, 0x17f5fa4a, 0x3e66d284, 0x0051aa3b, 0x483f7913, 0x2cfe5f15, 0x023427ca,
			0x2cc78315, 0x1e36ea47,
		},
	}

	externalFinal := [][]uint64{
		{
			0x7290a80d, 0x6f7e5329, 0x598ec8a8, 0x76a859a0, 0x6559e868, 0x657b83af, 0x13271d3f,
			0x1f876063, 0x0aeeae37, 0x706e9ca6, 0x46400cee, 0x72a05c26, 0x2c589c9e, 0x20bd37a7,
			0x6a2d3d10, 0x20523767,
		},
		{
			0x5b8fe9c4, 0x2aa501d6, 0x1e01ac3e, 0x1448bc54, 0x5ce5ad1c, 0x4918a14d, 0x2c46a83f,
			0x4fcf6876, 0x61d8d5c8, 0x6ddf4ff9, 0x11fda4d3, 0x02933a8f, 0x170eaf81, 0x5a9c314f,
			0x49a12590, 0x35ec52a1,
		},
		{
			0x58eb1611, 0x5e481e65, 0x367125c9, 0x0eba33ba, 0x1fc28ded, 0x066399ad, 0x0cbec0ea,
			0x75fd1af0, 0x50f5bf4e, 0x643d5f41, 0x6f4fe718, 0x5b3cbbde, 0x1e3afb3e, 0x296fb027,
			0x45e1547b, 0x4a8db2ab,
		},
		{
			0x59986d19, 0x30bcdfa3, 0x1db63932, 0x1d7c2824, 0x53b33681, 0x0673b747, 0x038a98a3,
			0x2c5bce60, 0x351979cd, 0x5008fb73, 0x547bca78, 0x711af481, 0x3f93bf64, 0x644d987b,
			0x3c8bcd87, 0x608758b8,
		},
	}

	internal := []uint64{
		0x5a8053c0, 0x693be639, 0x3858867d, 0x19334f6b, 0x128f0fd8, 0x4e2b1ccb, 0x61210ce0, 0x3c318939,
		0x0b5b2f22, 0x2edb11d5, 0x213effdf, 0x0cac4606, 0x241af16d,
	}

	p := h.params

	for i := 0; i < p.NbFullRounds/2; i++ {
		for j := 0; j < p.Width; j++ {
			p.RoundKeys[i][j].SetUint64(externalInitial[i][j])
		}
	}

	for i := p.NbFullRounds / 2; i < p.NbPartialRounds+p.NbFullRounds/2; i++ {
		// internal
		p.RoundKeys[i][0].SetUint64(internal[i-p.NbFullRounds/2])
	}

	for i := p.NbPartialRounds + p.NbFullRounds/2; i < p.NbPartialRounds+p.NbFullRounds; i++ {
		for j := 0; j < p.Width; j++ {
			p.RoundKeys[i][j].SetUint64(externalFinal[i-p.NbPartialRounds-p.NbFullRounds/2][j])
		}
	}

}
