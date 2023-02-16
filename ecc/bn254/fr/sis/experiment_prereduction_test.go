package sis_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

/*
	Just a piece of code independent from the rest to assess if an approach for sis
	works. The idea is that we limb-expand field elements in 6 limbs of 64 bits (instead of 4 normally).
	To give an example, the field element 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	would be represented as

		0x00000fffffffffff
		0x00000fffffffffff
		0x00000fffffffffff
		0x00000fffffffffff
		0x00000fffffffffff
		0x00000fffffffffff

	Thus each limbs has 20 bits of margin. Multiplying by a small number of bits, reduces the margin
	down to 17 bits. This leaves enough room to perform 2^17 additions before we need to `reset`. Since
	reduction will happen "only occasionally" and since we are only interested in performances, we will
	neglect it in the benchmarks.
*/

const (
	MAX_NORM     int = 8
	LOG_MAX_NORM int = 3
	// We have that 254 / 3 = 85, but in practice each field element is given in limbs
	// of uint64. In the end, it is simpler to work with 22 limbs per uint64, than it
	// is to work with large integers. If we wanted to have something more optimal we
	// would need to implement it by hand.
	NUM_LIMBS_FIELD            int = 88
	NUM_LIMBS_U64              int = 22
	NUM_LIMBS_PER_FIELD_IN_KEY int = 6
	NUM_FIELD_ELEMENT_KEY      int = 2
	NUM_LIMBS_PER_KEY_ENTRY    int = NUM_LIMBS_PER_FIELD_IN_KEY * NUM_FIELD_ELEMENT_KEY
)

type ExpandedKeyEntry = [NUM_LIMBS_PER_KEY_ENTRY]uint64

func BenchmarkLimbIdea(b *testing.B) {

	vecSize := 1 << 20

	key := GenerateKey(vecSize)

	vec := make([]fr.Element, vecSize)
	for i := range vec {
		// Worst case
		vec[i].SetInt64(-1)
	}

	b.ResetTimer()
	for _cnt := 0; _cnt < b.N; _cnt++ {
		HashLimbExpanded(key, vec)
	}

}

/*
The key is in limb expanded form, thus each entry will contain 12 u64 instead of two field element. The user
pass the "number" of field element he wishes to hash.
*/
func GenerateKey(nbField int) [][NUM_LIMBS_PER_KEY_ENTRY]uint64 {

	/*
		Since the logTwoBound is 2^3, we need 85 entries to represent a field element
	*/
	key := make([][NUM_LIMBS_PER_KEY_ENTRY]uint64, nbField*NUM_LIMBS_FIELD)

	for i := range key {
		for j := 0; j < NUM_LIMBS_PER_KEY_ENTRY; j++ {
			// Not really random but that does not matter for the performances
			key[i][j] = uint64(0x00000fffffffffff - NUM_LIMBS_PER_KEY_ENTRY*i - j)
		}
	}

	return key
}

/*
We assume that each input is already short. That way, we are not impacted by the cost
of "splitting" each field element into limbs. Since the limb are very small, we fit them
on uint8 integers. The limbs can be considered as
*/
func HashLimbExpanded(key []ExpandedKeyEntry, inputFields []fr.Element) ExpandedKeyEntry {

	var res ExpandedKeyEntry

	for i := 0; i < len(inputFields); i++ {

		xRegular := inputFields[i]
		xRegular = xRegular.Bits()

		var xLimb uint64

		pos := 0

		for j := 0; j < 4; j++ {

			xU64 := xRegular[j]

			// We unroll the loop by taking "2" entries at a time

			for k := 0; k < NUM_LIMBS_U64/2; k++ {

				xLimb = xU64 & 7 // This takes the 3 last bits of xU64
				xU64 >>= 3       // We shift, so that, next time we get the 3 next bits etc..

				// Load to key portion at once
				res[0] += key[pos][0] * xLimb
				res[1] += key[pos][1] * xLimb
				res[2] += key[pos][2] * xLimb
				res[3] += key[pos][3] * xLimb
				res[4] += key[pos][4] * xLimb
				res[5] += key[pos][5] * xLimb
				res[6] += key[pos][6] * xLimb
				res[7] += key[pos][7] * xLimb
				res[8] += key[pos][8] * xLimb
				res[9] += key[pos][9] * xLimb
				res[10] += key[pos][10] * xLimb
				res[11] += key[pos][11] * xLimb

				pos++
				xLimb = xU64 & 7 // This takes the 3 last bits of xU64
				xU64 >>= 3       // We shift, so that, next time we get the 3 next bits etc..

				// Load to key portion at once
				res[0] += key[pos][0] * xLimb
				res[1] += key[pos][1] * xLimb
				res[2] += key[pos][2] * xLimb
				res[3] += key[pos][3] * xLimb
				res[4] += key[pos][4] * xLimb
				res[5] += key[pos][5] * xLimb
				res[6] += key[pos][6] * xLimb
				res[7] += key[pos][7] * xLimb
				res[8] += key[pos][8] * xLimb
				res[9] += key[pos][9] * xLimb
				res[10] += key[pos][10] * xLimb
				res[11] += key[pos][11] * xLimb

				pos++

			}
		}
	}

	return res
}
