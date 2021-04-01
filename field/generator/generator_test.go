// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/field"
)

// integration test will create modulus for various field sizes and run tests

const rootDir = "integration_test"

func TestIntegration(t *testing.T) {
	os.RemoveAll(rootDir)
	err := os.MkdirAll(rootDir, 0700)
	defer os.RemoveAll(rootDir)
	if err != nil {
		t.Fatal(err)
	}

	var bits []int
	if testing.Short() {
		for i := 128; i <= 448; i += 64 {
			bits = append(bits, i-3, i-2, i-1, i, i+1)
		}
	} else {
		for i := 120; i < 704; i++ {
			bits = append(bits, i)
		}
	}

	moduli := make(map[string]string)
	for _, i := range bits {
		var q *big.Int
		var nbWords int
		if i%64 == 0 {
			q, _ = rand.Prime(rand.Reader, i)
			moduli[fmt.Sprintf("e_cios_%04d", i)] = q.String()
		} else {
			for {
				q, _ = rand.Prime(rand.Reader, i)
				nbWords = len(q.Bits())
				const B = (^uint64(0) >> 1) - 1
				if uint64(q.Bits()[nbWords-1]) <= B {
					break
				}
			}
			moduli[fmt.Sprintf("e_nocarry_%04d", i)] = q.String()
		}

	}

	for elementName, modulus := range moduli {
		// generate field
		childDir := filepath.Join(rootDir, elementName)
		fIntegration, err := field.NewField("integration", elementName, modulus)
		if err != nil {
			t.Fatal(elementName, err)
		}
		if err := GenerateFF(fIntegration, childDir); err != nil {
			t.Fatal(elementName, err)
		}
	}

	// run go test
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	packageDir := filepath.Join(wd, rootDir) + string(filepath.Separator) + "..."
	cmd := exec.Command("go", "test", packageDir)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		t.Fatal(err)
	}

}
