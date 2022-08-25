package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// This is an implementation of Fiat-Shamir optimized for in-circuit verifiers.
// i.e. the hashes used operate on and return field elements.

type ArithmeticTranscript interface {
	NextFromElements(_ []fr.Element) fr.Element // NextFromElements does not directly incorporate the number of elements in the hash. The size must be fixed and checked by the verifier.
	NextFromBytes(_ []byte) fr.Element          // NextFromElements does not directly incorporate the size of input in the hash. The size must be fixed and checked by the verifier.
}

func NextChallenge(transcript ArithmeticTranscript, input interface{}) fr.Element {
	switch i := input.(type) {
	case []byte:
		return transcript.NextFromBytes(i)
	case []fr.Element:
		return transcript.NextFromElements(i)
	case fr.Element:
		return transcript.NextFromElements([]fr.Element{i})
	case *fr.Element:
		return transcript.NextFromElements([]fr.Element{*i})

	default:
		panic("invalid hash input type")
	}
}

func NextFromBytes(transcript ArithmeticTranscript, bytes []byte, count int) []fr.Element {
	res := make([]fr.Element, count)

	for i := 0; i < count; i++ {
		res[i] = transcript.NextFromBytes(bytes)
		bytes = nil
	}

	return res
}
