package bls12381

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	testDir                = "testing/bls"
	deserializationG1Tests = filepath.Join(testDir, "deserialization_G1/*")
)

func TestDeserializationG1(t *testing.T) {

	type Test struct {
		Input struct {
			PubKeyHexStr string `yaml:"pubkey"`
		}
		IsValidPredicate *bool `yaml:"output"`
	}
	tests, err := filepath.Glob(deserializationG1Tests)
	require.NoError(t, err)
	for _, testPath := range tests {
		t.Run(testPath, func(t *testing.T) {
			testFile, err := os.Open(testPath)
			require.NoError(t, err)
			test := Test{}
			err = yaml.NewDecoder(testFile).Decode(&test)
			require.NoError(t, testFile.Close())
			require.NoError(t, err)
			//fmt.Println(testFile)
			testCaseValid := test.IsValidPredicate != nil
			byts, err := hex.DecodeString(test.Input.PubKeyHexStr)
			if err != nil && testCaseValid {
				panic(err)
			}

			var point G1Affine
			_, err = point.SetBytes(byts[:])
			if err == nil && !testCaseValid {
				panic("err should not be nil")
			}

		})
	}
}
