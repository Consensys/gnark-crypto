package gkr

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational/sumcheck"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSingleIdentityGateTwoInstancesProver(t *testing.T) {
	testCase := newTestCase(t, "../../rational_cases/single_identity_gate_two_instances.json")
	testCase.Transcript.Update(0)
	proof := Prove(testCase.Circuit, testCase.FullAssignment, testCase.Transcript)
	serialized, err := json.Marshal(proof)
	assert.NoError(t, err)
	fmt.Println(string(serialized))
	//printGkrProof(t, proof)
	assertProofEquals(t, testCase.Proof, proof)
}

func TestSingeIdentityGateTwoInstancesVerifier(t *testing.T) {
	testCase := newTestCase(t, "../../rational_cases/single_identity_gate_two_instances.json")
	testCase.Transcript.Update(0)
	success := Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
	assert.True(t, success)

	testCase = newTestCase(t, "../../rational_cases/single_identity_gate_two_instances.json")
	testCase.Transcript.Update(1)
	success = Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
	assert.False(t, success)

}

func TestLoadCircuit(t *testing.T) {
	if getCircuit(t, "../../rational_cases/resources/single_identity_gate.json") == nil {
		t.Fail()
	}
}

func TestLoadHash(t *testing.T) {
	h := MapHashTranscript{hashMap: getHash(t, "../../rational_cases/resources/hash.json")}

	var one small_rational.SmallRational
	one.SetOne()
	s := h.Next(one)
	fmt.Println(s.Text(10))
}

var hashCache = make(map[string]map[pair]small_rational.SmallRational)

func getHash(t *testing.T, path string) map[pair]small_rational.SmallRational {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if h, ok := hashCache[path]; ok {
		return h
	}
	if bytes, err := os.ReadFile(path); err == nil {
		var asMap map[string]interface{}
		if err := json.Unmarshal(bytes, &asMap); err != nil {
			t.Error(err)
		}

		res := make(map[pair]small_rational.SmallRational)

		for k, v := range asMap {
			var value small_rational.SmallRational
			if _, err := value.Set(v); err != nil {
				t.Error(err)
			}

			key := strings.Split(k, ",")
			var pair pair
			switch len(key) {
			case 1:
				pair.secondPresent = false
			case 2:
				pair.secondPresent = true
				if _, err := pair.second.Set(key[1]); err != nil {
					t.Error(err)
				}
			default:
				t.Error(fmt.Errorf("cannot parse %T as one or two field elements", v))
			}
			if _, err := pair.first.Set(key[0]); err != nil {
				t.Error(err)
			}

			res[pair] = value
		}
		hashCache[path] = res
		return res

	} else {
		t.Error(err)
	}
	return nil //Unreachable
}

type pair struct {
	first         small_rational.SmallRational
	second        small_rational.SmallRational
	secondPresent bool
}

type MapHashTranscript struct {
	hashMap         map[pair]small_rational.SmallRational
	stateValid      bool
	resultAvailable bool
	state           small_rational.SmallRational
}

func (m *MapHashTranscript) hash(x *small_rational.SmallRational, y *small_rational.SmallRational) small_rational.SmallRational {
	// Not too concerned with efficiency
	for k, v := range m.hashMap {
		if k.first.Equal(x) {
			if y == nil {
				if !k.secondPresent {
					return v
				}
			} else if k.secondPresent && k.second.Equal(y) {
				return v
			}
		}
	}

	if y == nil {
		panic("No hash available for input " + x.Text(10))
	} else {
		panic("No hash available for input " + x.Text(10) + "," + y.Text(10))
	}
}

func (m *MapHashTranscript) Update(i ...interface{}) {
	if len(i) > 0 {
		for _, x := range i {

			var xElement small_rational.SmallRational
			if _, err := xElement.Set(x); err != nil {
				panic(err.Error())
			}
			if m.stateValid {
				m.state = m.hash(&xElement, &m.state)
			} else {
				m.state = m.hash(&xElement, nil)
			}

			m.stateValid = true
		}
	} else { //just hash the state itself
		if !m.stateValid {
			panic("nothing to hash")
		}
		m.state = m.hash(&m.state, nil)
	}
	m.resultAvailable = true
}

func (m *MapHashTranscript) Next(i ...interface{}) small_rational.SmallRational {

	if len(i) > 0 || !m.resultAvailable {
		m.Update(i...)
	}
	m.resultAvailable = false
	return m.state
}

func (m *MapHashTranscript) NextN(N int, i ...interface{}) []small_rational.SmallRational {

	if len(i) > 0 {
		m.Update(i...)
	}

	res := make([]small_rational.SmallRational, N)

	for n := range res {
		res[n] = m.Next()
	}

	return res
}

type WireInfo struct {
	Gate   string  `json:"gate"`
	Inputs [][]int `json:"inputs"`
}

type CircuitInfo [][]WireInfo

var circuitCache = make(map[string]Circuit)

func getCircuit(t *testing.T, path string) Circuit {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if circuit, ok := circuitCache[path]; ok {
		return circuit
	}
	if bytes, err := os.ReadFile(path); err == nil {
		var circuitInfo CircuitInfo
		if err := json.Unmarshal(bytes, &circuitInfo); err == nil {
			circuit := circuitInfo.toCircuit()
			circuitCache[path] = circuit
			return circuit
		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
	return nil //unreachable
}

func (c CircuitInfo) toCircuit() (circuit Circuit) {
	isOutput := make(map[*Wire]interface{})
	circuit = make(Circuit, len(c))
	for i := len(c) - 1; i >= 0; i-- {
		circuit[i] = make(CircuitLayer, len(c[i]))
		for j, wireInfo := range c[i] {
			circuit[i][j].Gate = gates[wireInfo.Gate]
			circuit[i][j].Inputs = make([]*Wire, len(wireInfo.Inputs))
			isOutput[&circuit[i][j]] = nil
			for k, inputCoord := range wireInfo.Inputs {
				if len(inputCoord) != 2 {
					panic("circuit wire has two coordinates")
				}
				input := &circuit[inputCoord[0]][inputCoord[1]]
				input.NumOutputs++
				circuit[i][j].Inputs[k] = input
				delete(isOutput, input)
			}
			if (i == len(c)-1) != (len(circuit[i][j].Inputs) == 0) {
				panic("wire is input if and only if in last layer")
			}
		}
	}

	for k := range isOutput {
		k.NumOutputs = 1
	}

	return
}

var gates map[string]Gate

func init() {
	gates = make(map[string]Gate)
	gates["identity"] = identityGate{}
	gates["mul"] = mulGate{}
}

type TestCase struct {
	Circuit         Circuit
	Transcript      sumcheck.ArithmeticTranscript
	Proof           Proof
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
}

type TestCaseInfo struct {
	Hash    string          `json:"hash"`
	Circuit string          `json:"circuit"`
	Input   [][]interface{} `json:"input"`
	Output  [][]interface{} `json:"output"`
	Proof   PrintableProof  `json:"proof"`
}

type ParsedTestCase struct {
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
	Proof           Proof
	Hash            map[pair]small_rational.SmallRational
	Circuit         Circuit
}

var parsedTestCases = make(map[string]*ParsedTestCase)

func newTestCase(t *testing.T, path string) *TestCase {
	path, err := filepath.Abs(path)
	assert.NoError(t, err)
	dir := filepath.Dir(path)

	parsedCase, ok := parsedTestCases[path]
	if !ok {
		if bytes, err := os.ReadFile(path); err == nil {
			var info TestCaseInfo
			err = json.Unmarshal(bytes, &info)
			if err != nil {
				t.Error(err)
			}

			circuit := getCircuit(t, filepath.Join(dir, info.Circuit))
			hash := getHash(t, filepath.Join(dir, info.Hash))
			proof := unmarshalProof(t, info.Proof)

			fullAssignment := make(WireAssignment)
			inOutAssignment := make(WireAssignment)
			assignmentSize := len(info.Input[0])

			{
				i := len(circuit) - 1

				assert.Equal(t, len(circuit[i]), len(info.Input), "Input layer not the same size as input vector")

				for j := range circuit[i] {
					wire := &circuit[i][j]
					wireAssignment := sliceToElementSlice(t, info.Input[j])
					fullAssignment[wire] = wireAssignment
					inOutAssignment[wire] = wireAssignment
				}
			}

			for i := len(circuit) - 2; i >= 0; i-- {
				for j := range circuit[i] {
					wire := &circuit[i][j]
					assignment := make(polynomial.MultiLin, assignmentSize)
					in := make([]small_rational.SmallRational, len(wire.Inputs))
					for k := range assignment {
						for l, inputWire := range circuit[i][j].Inputs {
							in[l] = fullAssignment[inputWire][k]
						}
						assignment[k] = wire.Gate.Evaluate(in...)
					}

					fullAssignment[wire] = assignment
				}
			}

			assert.Equal(t, len(circuit[0]), len(info.Output), "Output layer not the same size as output vector")
			for j := range circuit[0] {
				wire := &circuit[0][j]
				inOutAssignment[wire] = sliceToElementSlice(t, info.Output[j])
				assertSliceEquals(t, inOutAssignment[wire], fullAssignment[wire])
			}

			parsedCase = &ParsedTestCase{
				FullAssignment:  fullAssignment,
				InOutAssignment: inOutAssignment,
				Proof:           proof,
				Hash:            hash,
				Circuit:         circuit,
			}

			parsedTestCases[path] = parsedCase
		} else {
			t.Error(err)
		}
	}

	return &TestCase{
		Circuit:         parsedCase.Circuit,
		Transcript:      &MapHashTranscript{hashMap: parsedCase.Hash},
		FullAssignment:  parsedCase.FullAssignment,
		InOutAssignment: parsedCase.InOutAssignment,
		Proof:           parsedCase.Proof,
	}
}

func sliceToElementSlice(t *testing.T, slice []interface{}) (elementSlice []small_rational.SmallRational) {
	elementSlice = make([]small_rational.SmallRational, len(slice))
	for i, v := range slice {
		if _, err := elementSlice[i].Set(v); err != nil {
			t.Error(err)
		}
	}
	return
}

func assertSliceEquals(t *testing.T, a []small_rational.SmallRational, b []small_rational.SmallRational) {
	assert.Equal(t, len(a), len(b))
	for i := range a {
		if !a[i].Equal(&b[i]) {
			t.Error(a[i].String(), "≠", b[i].String())
		}
	}
}

func assertProofEquals(t *testing.T, expected Proof, seen Proof) {
	assert.Equal(t, len(expected), len(seen))
	for i, x := range expected {
		xSeen := seen[i]
		assert.Equal(t, len(x), len(xSeen))
		for j, y := range x {
			ySeen := xSeen[j]

			if ySeen.FinalEvalProof == nil {
				assert.Equal(t, 0, len(y.FinalEvalProof.([]small_rational.SmallRational)))
			} else {
				assert.Equal(t, y.FinalEvalProof, ySeen.FinalEvalProof)
			}
			assert.Equal(t, len(y.PartialSumPolys), len(ySeen.PartialSumPolys))
			for k, z := range y.PartialSumPolys {
				zSeen := ySeen.PartialSumPolys[k]
				assertSliceEquals(t, z, zSeen)
			}
		}
	}
}

type PrintableProof [][]PrintableSumcheckProof

type PrintableSumcheckProof struct {
	FinalEvalProof  interface{}     `json:"finalEvalProof"`
	PartialSumPolys [][]interface{} `json:"partialSumPolys"`
}

func unmarshalProof(t *testing.T, printable PrintableProof) (proof Proof) {
	proof = make(Proof, len(printable))
	for i := range printable {
		proof[i] = make([]sumcheck.Proof, len(printable[i]))
		for j, printableSumcheck := range printable[i] {
			finalEvalProof := []small_rational.SmallRational(nil)

			if printableSumcheck.FinalEvalProof != nil {
				finalEvalSlice := reflect.ValueOf(printableSumcheck.FinalEvalProof)
				finalEvalProof = make([]small_rational.SmallRational, finalEvalSlice.Len())
				for k := range finalEvalProof {
					_, err := finalEvalProof[k].Set(finalEvalSlice.Index(k).Interface())
					assert.NoError(t, err)
				}
			}

			proof[i][j] = sumcheck.Proof{
				PartialSumPolys: make([]polynomial.Polynomial, len(printableSumcheck.PartialSumPolys)),
				FinalEvalProof:  finalEvalProof,
			}
			for k := range printableSumcheck.PartialSumPolys {
				proof[i][j].PartialSumPolys[k] = sliceToElementSlice(t, printableSumcheck.PartialSumPolys[k])
			}
		}
	}
	return
}

/*type PrintableProof [][]PrintableSumcheckProof

type PrintableSumcheckProof struct {
	FinalEvalProof interface{}`json:"finalEvalProof"`
	PartialSumPolys []polynomial.Polynomial `json:"partialSumPolys"`
}

func toPrintableProof(t *testing.T, proof Proof) (printable PrintableProof){
	printable = make(PrintableProof, len(proof))

	for i, layer := range proof {
		printable[i] = make([]PrintableSumcheckProof, len(layer))
		for j, wire := range layer {

			assert.Nil(t, wire.FinalEvalProof)
			printable[i][j] = make([]polynomial.Polynomial, len(wire.PartialSumPolys))
			for k, poly := range wire.PartialSumPolys {
				printable[i][j][k] = make(polynomial.Polynomial, len(poly))
				for l := range poly {
					printable[i][j][k][l] = poly[l]
				}
			}
		}
	}
	serialized, err := json.Marshal(printable)
	assert.NoError(t, err)
	fmt.Println(serialized)
}*/

/*func toPrintableProof(t *testing.T, proof Proof) {
	forPrint := make([][][]polynomial.Polynomial, len(proof))

	for i, layer := range proof {
		forPrint[i] = make([][]polynomial.Polynomial, len(layer))
		for j, wire := range layer {
			assert.Nil(t, wire.FinalEvalProof)
			forPrint[i][j] = make([]polynomial.Polynomial, len(wire.PartialSumPolys))
			for k, poly := range wire.PartialSumPolys {
				forPrint[i][j][k] = make(polynomial.Polynomial, len(poly))
				for l := range poly {
					forPrint[i][j][k][l] = poly[l]
				}
			}
		}
	}
	serialized, err := json.Marshal(forPrint)
	assert.NoError(t, err)
	fmt.Println(serialized)
}
*/
/*func printProof(proof Proof) {
	fmt.Println("[")
	for i, layer := range proof {
		fmt.Println("[")

		for j, wire := range layer {
			fmt.Println("[")
			for k, poly := range wire.PartialSumPolys {
				fmt.Println("[")
				for l := range poly {
					fmt.Print(poly[l].String())
					if l+1 != len(poly) {
						fmt.Print(",")
					}
				}
				fmt.Print("[")
				if k+1 != len(wire.PartialSumPolys) {
					fmt.Print(",")
				}
				fmt.Println()
			}
			fmt.Print("]")
			if j+1 != len(layer) {
				fmt.Pr
			}
		}

		fmt.Println("]")
		if i != len(proof) {
			fmt.Println(",")
		}
	}
}*/
