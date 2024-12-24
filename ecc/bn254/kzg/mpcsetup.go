package kzg

import (
	curve "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/mpcsetup"
	"io"
)

type MpcSetup struct {
	srs   SRS
	proof mpcsetup.UpdateProof
}

func InitializeSetup(N int) MpcSetup {
	var res MpcSetup
	_, _, g1, g2 := curve.Generators()

	res.srs.Pk.G1 = make([]curve.G1Affine, N)
	for i := range N {
		res.srs.Pk.G1[i] = g1
	}
	res.srs.Vk.G1 = g1
	res.srs.Vk.G2[0] = g2
	res.srs.Vk.G2[1] = g2

	return res
}

// WriteTo implements io.WriterTo
func (s *MpcSetup) WriteTo(w io.Writer) (int64, error) {
	return mpcsetup.WriteTo(w, s.srs.Vk.G2[1], s.proof)

}

func (s *MpcSetup) Contribute() {

	s.proof = mpcsetup.UpdateValues(nil, nil, 0, &s.srs.Vk.G2[1])
}
