package gkr

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational/sumcheck"
	"os"
	"strings"
	"testing"
)

func TestLoadCircuit(t *testing.T) {
	if bytes, err := os.ReadFile("../../rational_cases/resources/single_identity_gate.json"); err == nil {
		var circuitInfo CircuitInfo
		if err := json.Unmarshal(bytes, &circuitInfo); err == nil {
			circuit := circuitInfo.toCircuit()
			circuit.Size()
		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
}

func TestLoadHash(t *testing.T) {
	h, err := getHash("../../rational_cases/resources/hash.json")
	if err != nil {
		t.Error(err)
	}
	var one small_rational.SmallRational
	one.SetOne()
	s := h.Next(one)
	fmt.Println(s.Text(10))
}

var transcriptCache = make(map[string]sumcheck.ArithmeticTranscript)

func getHash(path string) (sumcheck.ArithmeticTranscript, error) {
	if h, ok := transcriptCache[path]; ok {
		return h, nil
	}
	if bytes, err := os.ReadFile(path); err == nil {
		var asMap map[string]interface{}
		if err := json.Unmarshal(bytes, &asMap); err != nil {
			return nil, err
		}

		res := MapHashTranscript{
			hashMap: make(map[pair]small_rational.SmallRational),
		}
		for k, v := range asMap {
			var value small_rational.SmallRational
			if _, err := value.Set(v); err != nil {
				return nil, err
			}

			key := strings.Split(k, ",")
			var pair pair
			switch len(key) {
			case 1:
				pair.secondPresent = false
			case 2:
				pair.secondPresent = true
				if _, err := pair.second.Set(key[1]); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("cannot parse %T as one or two field elements", v)
			}
			if _, err := pair.first.Set(key[0]); err != nil {
				return nil, err
			}

			res.hashMap[pair] = value
		}
		transcriptCache[path] = &res
		return &res, nil

	} else {
		return nil, err
	}
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
			if (k.secondPresent && k.second.Equal(y)) || ( /*!k.secondPresent &&*/ y == nil) {
				return v
			}
		}
	}
	panic("No hash available for input " + x.Text(10))
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
	m.Update(i...)

	res := make([]small_rational.SmallRational, N)

	for n := range res {
		res[n] = m.Next(small_rational.SmallRational{})
	}

	return res
}

type WireInfo struct {
	Gate   string  `json:"gate"`
	Inputs [][]int `json:"inputs"`
}

type CircuitInfo [][]WireInfo

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
