package multisethash

import (
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/octobear"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

// Poseidon2-sponge vector ECMSH (Section 4.3 "preferred concrete derivation").
//
// The digest is a vector of N = 23 ECMSH accumulators. The N ordinates are
// derived by absorbing (domainTag, msg) into a width-16 Poseidon2 sponge with
// rate 8, then squeezing 3 permutations worth of output (24 elements, 23 used).
// Each squeezed element u is reduced into [0, floor(p / (2*T))), giving a slot
// s and ordinate y = T*s + k for some k in [0, T). With T = 256, every y stays
// below p/2 so the image is inverse-free.
const (
	pqN            = 23
	pqT            = 256
	pqWidth        = 16 // Poseidon2 state width
	pqSqueezeRate  = 8  // koalabear elements consumed per permutation
	pqPermutations = 3  // ceil(pqN / pqSqueezeRate)
)

// pqDomainTag is the 8-byte ASCII domain separator absorbed before the
// message.
var pqDomainTag = [8]byte{'E', 'C', 'M', 'S', 'H', '_', 'P', 'Q'}

// pqReducerBound = floor(p / (2*T)) is the upper bound on the slot s
// extracted from each squeezed koalabear element. With p = 2^31 - 2^24 + 1
// and T = 256, this is floor(2130706433 / 512) = 4161536.
var (
	errPqSlotOutOfRange = errors.New("octobear vector multiset hash: Poseidon2 slot out of range")

	pqReducerBound = func() *big.Int {
		p := koalabear.Modulus()
		denom := big.NewInt(2 * pqT)
		return new(big.Int).Div(p, denom)
	}()

	pqPermOnce sync.Once
	pqPermImpl *poseidon2.Permutation
)

// pqPerm returns the lazily-initialized width-16 Poseidon2 permutation
// shared across all map calls.
func pqPerm() *poseidon2.Permutation {
	pqPermOnce.Do(func() {
		pqPermImpl = poseidon2.NewPermutation(pqWidth, 6, 21)
	})
	return pqPermImpl
}

// Poseidon2Accumulator holds the N affine accumulator points for the
// Poseidon2-sponge vector ECMSH.
type Poseidon2Accumulator struct {
	sum [pqN]octobear.G1Affine
}

// NewPoseidon2Accumulator returns a zero (all-infinity) Poseidon2Accumulator.
func NewPoseidon2Accumulator() Poseidon2Accumulator {
	var a Poseidon2Accumulator
	for i := range a.sum {
		a.sum[i].SetInfinity()
	}
	return a
}

// Insert maps msg to N curve points via the Poseidon2 sponge domain separator
// and adds each point to the corresponding accumulator coordinate.
func (a *Poseidon2Accumulator) Insert(msg uint64) error {
	pts, _, err := MapPoseidon2(msg)
	if err != nil {
		return err
	}
	for i := range a.sum {
		a.sum[i].Add(&a.sum[i], &pts[i])
	}
	return nil
}

// Remove maps msg to N curve points and subtracts each from the corresponding
// accumulator coordinate.
func (a *Poseidon2Accumulator) Remove(msg uint64) error {
	pts, _, err := MapPoseidon2(msg)
	if err != nil {
		return err
	}
	var neg octobear.G1Affine
	for i := range a.sum {
		neg.Neg(&pts[i])
		a.sum[i].Add(&a.sum[i], &neg)
	}
	return nil
}

// Digest returns the current vector of accumulator points.
func (a *Poseidon2Accumulator) Digest() [pqN]octobear.G1Affine {
	return a.sum
}

// Reset clears the accumulator to the all-infinity state.
func (a *Poseidon2Accumulator) Reset() {
	for i := range a.sum {
		a.sum[i].SetInfinity()
	}
}

// HashPoseidon2 returns the Poseidon2-sponge vector ECMSH of msgs.
func HashPoseidon2(msgs []uint64) ([pqN]octobear.G1Affine, error) {
	acc := NewPoseidon2Accumulator()
	for _, msg := range msgs {
		if err := acc.Insert(msg); err != nil {
			return [pqN]octobear.G1Affine{}, err
		}
	}
	return acc.Digest(), nil
}

// MapPoseidon2 deterministically maps msg to N curve points using a
// Poseidon2 sponge over the koalabear field. It returns the N points and
// the per-coordinate tweak offsets k_i in [0, T) that produced them.
func MapPoseidon2(msg uint64) ([pqN]octobear.G1Affine, [pqN]uint8, error) {
	var (
		pts     [pqN]octobear.G1Affine
		offsets [pqN]uint8
	)

	squeezed, err := squeezePoseidon2(msg)
	if err != nil {
		return pts, offsets, err
	}

	_, b := octobear.CurveCoefficients()
	var tmp big.Int
	for i := 0; i < pqN; i++ {
		squeezed[i].BigInt(&tmp)
		tmp.Mod(&tmp, pqReducerBound)
		baseY := tmp.Uint64() * pqT

		p, k, err := mapAtBase(baseY, pqT, &b)
		if err != nil {
			return pts, offsets, err
		}
		pts[i] = p
		offsets[i] = k
	}
	return pts, offsets, nil
}

// MapAtSlot is a public helper used by the gnark in-circuit Poseidon2 vector
// ECMSH gadget. Given a slot s = u mod ⌊p/(2T)⌋ (already range-reduced by the
// caller — typically the in-circuit code after a Poseidon2 squeeze), it scans
// k in [0, pqT) and returns the first octobear curve point whose ordinate is
// y = pqT*s + k in the base subfield. The slot must satisfy
// pqT*s + (pqT-1) < p/2 to preserve inverse-freeness; this is automatic when
// s < ⌊p/(2T)⌋.
func MapAtSlot(slot uint64) (octobear.G1Affine, uint8, error) {
	if slot >= pqReducerBound.Uint64() {
		return octobear.G1Affine{}, 0, errPqSlotOutOfRange
	}
	_, b := octobear.CurveCoefficients()
	return mapAtBase(slot*pqT, pqT, &b)
}

// PqReducerBound returns ⌊p/(2T)⌋, the upper bound on the slot s produced by
// the Poseidon2 sponge range-reduction. Exported for the gnark in-circuit
// range-reduction constraint.
func PqReducerBound() *big.Int {
	return new(big.Int).Set(pqReducerBound)
}

// PqDomainTag returns the 8-byte ASCII domain separator absorbed before the
// message by MapPoseidon2. Exported so the in-circuit
// sponge can absorb the same bytes.
func PqDomainTag() [8]byte {
	return pqDomainTag
}

// squeezePoseidon2 absorbs the domain tag and msg into a width-16 sponge with
// rate 8 and returns the first pqPermutations * pqSqueezeRate squeezed
// koalabear elements.
func squeezePoseidon2(msg uint64) ([pqPermutations * pqSqueezeRate]koalabear.Element, error) {
	var (
		state    [pqWidth]koalabear.Element
		squeezed [pqPermutations * pqSqueezeRate]koalabear.Element
	)

	// Absorb (domainTag, msg) into the rate part (state[0:pqSqueezeRate]).
	// The 8-byte tag occupies state[0..1] as two 32-bit big-endian halves
	// (each < p, so SetUint64 is injective on the tag's domain). The 64-bit
	// msg is split into four 16-bit big-endian chunks across state[2..5];
	// each chunk is < 2^16 < p, so this encoding is injective for the full
	// uint64 domain. state[6..7] and the capacity state[8..15] stay zero.
	//
	// Note: a 32-bit-half encoding would not be injective here because
	// koalabear has p = 2^31 - 2^24 + 1 < 2^32, so e.g. msg = 0 and msg = p
	// would absorb the same field elements (mod p) and collide.
	state[0].SetUint64(uint64(binary.BigEndian.Uint32(pqDomainTag[0:4])))
	state[1].SetUint64(uint64(binary.BigEndian.Uint32(pqDomainTag[4:8])))
	state[2].SetUint64(uint64(uint16(msg >> 48)))
	state[3].SetUint64(uint64(uint16(msg >> 32)))
	state[4].SetUint64(uint64(uint16(msg >> 16)))
	state[5].SetUint64(uint64(uint16(msg)))

	perm := pqPerm()
	for i := 0; i < pqPermutations; i++ {
		if err := perm.Permutation(state[:]); err != nil {
			return squeezed, err
		}
		copy(squeezed[i*pqSqueezeRate:(i+1)*pqSqueezeRate], state[:pqSqueezeRate])
	}
	return squeezed, nil
}
