package poseidon

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/assert"
)

func elementFromString(v string) *fr.Element {
	n, success := new(big.Int).SetString(v, 10)
	if !success {
		panic("Error parsing hex number")
	}
	var e fr.Element
	e.SetBigInt(n)
	return &e
}

func TestPoseidon1(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("7764075183688725171230668857402392634761334547267776368103645048439717572548")
	length := 1
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon2(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("7142104613055408817911962100316808866448378443474503659992478482890339429929")
	length := 2
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon4(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("7817711165059374331357136443537800893307845083525445872661165200086166013245")
	length := 4
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon13(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("1709610050961943784828399921362905178787999827108026634048665681910636069934")
	length := 13
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon16(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("8319791455060392555425392842391403897548969645190976863995973180967774875286")
	length := 16
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon24(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("14281896993318141900551144554156181598834585543901557749703302979893059224887")
	length := 24
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon30(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("3706864405066113783363062549980271879113588784557216652303342540436728346372")
	length := 30
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestPoseidon256(t *testing.T) {
	// WARNING: No test vector to compare with
	expectedHash := elementFromString("3889232958018785041730045800798978544000060048890444628344970190264245196615")
	length := 256
	inputs := make([]*fr.Element, length)
	for i := 0; i < length; i++ {
		e := fr.NewElement((uint64)(i + 1))
		inputs[i] = &e
	}
	actualHash := Poseidon(inputs...)
	assert.True(t, actualHash.Equal(expectedHash), "%s != %s", actualHash, expectedHash)
}

func TestConsistency(t *testing.T) {
	// Check whether Poseidon returns the same value for the same input
	// Test vector https://extgit.iaik.tugraz.at/krypto/hadeshash/-/blob/master/code/test_vectors.txt
	inputsStr := []string{"1", "2", "3", "4"}
	inputs := make([]*fr.Element, len(inputsStr))
	for i := 0; i < len(inputsStr); i++ {
		inputs[i] = elementFromString(inputsStr[i])
	}
	actualHash1 := Poseidon(inputs...)
	actualHash2 := Poseidon(inputs...)
	assert.True(t, actualHash1.Equal(actualHash2), "%s != %s", actualHash1, actualHash2)
}

func TestPoseidonBytes(t *testing.T) {
	// Test vector https://extgit.iaik.tugraz.at/krypto/hadeshash/-/blob/master/code/test_vectors.txt
	expectedHash := elementFromString("7142104613055408817911962100316808866448378443474503659992478482890339429929")
	inputs := make([][]byte, 2)
	inputs[0] = make([]byte, 1)
	inputs[0][0] = 1
	inputs[1] = make([]byte, 1)
	inputs[1][0] = 2
	actualHash := PoseidonBytes(inputs...)
	actualHashEle := fr.Element{0, 0, 0, 0}
	actualHashEle.SetBytes(actualHash)
	assert.True(t, actualHashEle.Equal(expectedHash), "%s != %s", actualHashEle, expectedHash)
}

func TestDigest(t *testing.T) {
	expectedHash := elementFromString("7142104613055408817911962100316808866448378443474503659992478482890339429929")
	hFunc := NewPoseidon()
	inputs := make([][]byte, 2)
	inputs[0] = make([]byte, 1)
	inputs[0][0] = 1
	inputs[1] = make([]byte, 1)
	inputs[1][0] = 2
	hFunc.Write(inputs[0])
	hFunc.Write(inputs[1])
	actualHash := hFunc.Sum(nil)
	actualHashEle := fr.Element{0, 0, 0, 0}
	actualHashEle.SetBytes(actualHash)
	assert.True(t, actualHashEle.Equal(expectedHash), "%s != %s", actualHashEle, expectedHash)

	hFunc.Reset()
	bigNumber, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	inputs[0] = bigNumber.Bytes()
	n, err := hFunc.Write(inputs[0])
	assert.EqualError(t, err, "not support bytes bigger than modulus")
	assert.Equal(t, n, 0)
}
