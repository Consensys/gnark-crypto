package poseidon2

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/stretchr/testify/require"
)

const width = 16

type testCase struct {
	In  [][]uint32 `json:"in"`
	Out []uint32   `json:"out"`
}

func TestVectors(t *testing.T) {

	f, err := os.Open("test-vectors.json")
	require.NoError(t, err)
	defer f.Close()

	var testCases []testCase
	require.NoError(t, json.NewDecoder(f).Decode(&testCases))

	hsh := NewMerkleDamgardHasher()
	require.Equal(t, width, GetDefaultParameters().Width, "unexpected hash width")
	require.Equal(t, width*koalabear.Bytes, 2*hsh.BlockSize(), "unexpected input block size")

	var buf [width * koalabear.Bytes / 2]byte

	for _, c := range testCases {
		hsh.Reset()
		require.Equal(t, width/2, len(c.Out), "unexpected output block size")

		for _, in := range c.In {
			require.Equal(t, width/2, len(in), "unexpected input block size")

			for i := range in {
				_, err = binary.Encode(buf[i*koalabear.Bytes:(i+1)*koalabear.Bytes], binary.BigEndian, in[i])
				require.NoError(t, err)
			}
			_, err = hsh.Write(buf[:])
			require.NoError(t, err)
		}

		res := bytes.NewReader(hsh.Sum(nil))
		var out [width / 2]uint32
		for i := range width / 2 {
			require.NoError(t, binary.Read(res, binary.BigEndian, &out[i]))
		}
		require.Equal(t, c.Out, out[:], "unexpected output on input %v", c.In)
	}
}
