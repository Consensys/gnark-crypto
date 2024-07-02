package mimc_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiMCFiatShamir(t *testing.T) {
	fs := fiatshamir.NewTranscript(mimc.NewMiMC(), "c0")
	zero := make([]byte, mimc.BlockSize)
	err := fs.Bind("c0", zero)
	assert.NoError(t, err)
	_, err = fs.ComputeChallenge("c0")
	assert.NoError(t, err)
}

func TestByteOrder(t *testing.T) {
	assert := require.New(t)

	var buf [fr.Bytes]byte
	// if the 31 first bytes are FF, it's a valid FF in little endian, but not in big endian
	for i := 0; i < fr.Bytes-1; i++ {
		buf[i] = 0xFF
	}
	_, err := fr.BigEndian.Element(&buf)
	assert.Error(err)
	_, err = fr.LittleEndian.Element(&buf)
	assert.NoError(err)

	{
		// hashing buf with big endian should fail
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.BigEndian))
		_, err := mimcHash.Write(buf[:])
		assert.Error(err)
	}

	{
		// hashing buf with little endian should succeed
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.LittleEndian))
		_, err := mimcHash.Write(buf[:])
		assert.NoError(err)
	}

	buf = [fr.Bytes]byte{}
	// if the 31 bytes are FF, it's a valid FF in big endian, but not in little endian
	for i := 1; i < fr.Bytes; i++ {
		buf[i] = 0xFF
	}
	_, err = fr.BigEndian.Element(&buf)
	assert.NoError(err)
	_, err = fr.LittleEndian.Element(&buf)
	assert.Error(err)

	{
		// hashing buf with big endian should succeed
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.BigEndian))
		_, err := mimcHash.Write(buf[:])
		assert.NoError(err)
	}

	{
		// hashing buf with little endian should fail
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.LittleEndian))
		_, err := mimcHash.Write(buf[:])
		assert.Error(err)
	}

}
