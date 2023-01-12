package mimc

import (
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMiMCFiatShamir(t *testing.T) {
	fs := fiatshamir.NewTranscript(NewMiMC(), "c0")
	zero := make([]byte, BlockSize)
	err := fs.Bind("c0", zero)
	assert.NoError(t, err)
	_, err = fs.ComputeChallenge("c0")
	assert.NoError(t, err)
}
