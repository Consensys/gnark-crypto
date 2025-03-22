package poseidon2

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	fr "github.com/consensys/gnark-crypto/field/babybear"
	"github.com/stretchr/testify/require"
)

func TestPlonky3Interop(t *testing.T) {
	assert := require.New(t)

	width := []int{16, 24}
	for _, w := range width {
		var h *Permutation
		if w == 16 {
			h = NewPermutation(16, 8, 13)
			h.setHorizenRoundKeys16()
		} else {
			h = NewPermutation(24, 8, 21)
			h.setHorizenRoundKeys24()
		}
		fName := fmt.Sprintf("poseidon2_babybear_%d_test_vectors.csv", w)
		t.Logf("running test vectors from %s", fName)
		file, err := os.Open(fName)
		assert.NoError(err)
		// read poseidon2_babybear_XX_test_vectors.csv
		// it is structured;
		// 1 line of header
		// 1line = 32 uint64; 16 for input, 16 for expected result
		// this file was generated using plonky3 babybear implementation at commit de2b3b7
		// using default_babybear_poseidon2_16() and random inputs

		r := csv.NewReader(file)
		r.Read() // skip header
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			assert.NoError(err)
			assert.Equal(len(record), w*2)
			input := make([]fr.Element, w)
			expected := make([]fr.Element, w)
			for i := 0; i < w; i++ {
				v, err := strconv.Atoi(record[i])
				assert.NoError(err)

				input[i].SetUint64(uint64(v))
			}
			for i := w; i < w*2; i++ {
				v, err := strconv.Atoi(record[i])
				assert.NoError(err)
				expected[i-w].SetUint64(uint64(v))
			}

			h.Permutation(input[:])

			for i := range input {
				assert.Equal(input[i], expected[i])
			}
		}
	}

}

func (h *Permutation) setHorizenRoundKeys16() {
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

func (h *Permutation) setHorizenRoundKeys24() {

	externalInitial := [][]uint64{
		{
			0x0fa20c37, 0x0795bb97, 0x12c60b9c, 0x0eabd88e, 0x096485ca, 0x07093527, 0x1b1d4e50,
			0x30a01ace, 0x3bd86f5a, 0x69af7c28, 0x3f94775f, 0x731560e8, 0x465a0ecd, 0x574ef807,
			0x62fd4870, 0x52ccfe44, 0x14772b14, 0x4dedf371, 0x260acd7c, 0x1f51dc58, 0x75125532,
			0x686a4d7b, 0x54bac179, 0x31947706,
		},
		{
			0x29799d3b, 0x6e01ae90, 0x203a7a64, 0x4f7e25be, 0x72503f77, 0x45bd3b69, 0x769bd6b4,
			0x5a867f08, 0x4fdba082, 0x251c4318, 0x28f06201, 0x6788c43a, 0x4c6d6a99, 0x357784a8,
			0x2abaf051, 0x770f7de6, 0x1794b784, 0x4796c57a, 0x724b7a10, 0x449989a7, 0x64935cf1,
			0x59e14aac, 0x0e620bb8, 0x3af5a33b,
		},
		{
			0x4465cc0e, 0x019df68f, 0x4af8d068, 0x08784f82, 0x0cefdeae, 0x6337a467, 0x32fa7a16,
			0x486f62d6, 0x386a7480, 0x20f17c4a, 0x54e50da8, 0x2012cf03, 0x5fe52950, 0x09afb6cd,
			0x2523044e, 0x5c54d0ef, 0x71c01f3c, 0x60b2c4fb, 0x4050b379, 0x5e6a70a5, 0x418543f5,
			0x71debe56, 0x1aad2994, 0x3368a483,
		},
		{
			0x07a86f3a, 0x5ea43ff1, 0x2443780e, 0x4ce444f7, 0x146f9882, 0x3132b089, 0x197ea856,
			0x667030c3, 0x2317d5dc, 0x0c2c48a7, 0x56b2df66, 0x67bd81e9, 0x4fcdfb19, 0x4baaef32,
			0x0328d30a, 0x6235760d, 0x12432912, 0x0a49e258, 0x030e1b70, 0x48caeb03, 0x49e4d9e9,
			0x1051b5c6, 0x6a36dbbe, 0x4cff27a5,
		},
	}

	externalFinal := [][]uint64{
		{
			0x032959ad, 0x2b18af6a, 0x55d3dc8c, 0x43bd26c8, 0x0c41595f, 0x7048d2e2, 0x00db8983,
			0x2af563d7, 0x6e84758f, 0x611d64e1, 0x1f9977e2, 0x64163a0a, 0x5c5fc27b, 0x02e22561,
			0x3a2d75db, 0x1ba7b71a, 0x34343f64, 0x7406b35d, 0x19df8299, 0x6ff4480a, 0x514a81c8,
			0x57ab52ce, 0x6ad69f52, 0x3e0c0e0d,
		},
		{
			0x48126114, 0x2a9d62cc, 0x17441f23, 0x485762bb, 0x2f218674, 0x06fdc64a, 0x0861b7f2,
			0x3b36eee6, 0x70a11040, 0x04b31737, 0x3722a872, 0x2a351c63, 0x623560dc, 0x62584ab2,
			0x382c7c04, 0x3bf9edc7, 0x0e38fe51, 0x376f3b10, 0x5381e178, 0x3afc61c7, 0x5c1bcb4d,
			0x6643ce1f, 0x2d0af1c1, 0x08f583cc,
		},
		{
			0x5d6ff60f, 0x6324c1e5, 0x74412fb7, 0x70c0192e, 0x0b72f141, 0x4067a111, 0x57388c4f,
			0x351009ec, 0x0974c159, 0x539a58b3, 0x038c0cff, 0x476c0392, 0x3f7bc15f, 0x4491dd2c,
			0x4d1fef55, 0x04936ae3, 0x58214dd4, 0x683c6aad, 0x1b42f16b, 0x6dc79135, 0x2d4e71ec,
			0x3e2946ea, 0x59dce8db, 0x6cee892a,
		},
		{
			0x47f07350, 0x7106ce93, 0x3bd4a7a9, 0x2bfe636a, 0x430011e9, 0x001cd66a, 0x307faf5b,
			0x0d9ef3fe, 0x6d40043a, 0x2e8f470c, 0x1b6865e8, 0x0c0e6c01, 0x4d41981f, 0x423b9d3d,
			0x410408cc, 0x263f0884, 0x5311bbd0, 0x4dae58d8, 0x30401cea, 0x09afa575, 0x4b3d5b42,
			0x63ac0b37, 0x5fe5bb14, 0x5244e9d4,
		},
	}

	internal := []uint64{
		0x1da78ec2, 0x730b0924, 0x3eb56cf3, 0x5bd93073, 0x37204c97, 0x51642d89, 0x66e943e8, 0x1a3e72de,
		0x70beb1e9, 0x30ff3b3f, 0x4240d1c4, 0x12647b8d, 0x65d86965, 0x49ef4d7c, 0x47785697, 0x46b3969f,
		0x5c7b7a0e, 0x7078fc60, 0x4f22d482, 0x482a9aee, 0x6beb839d,
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
