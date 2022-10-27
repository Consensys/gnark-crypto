package test_vector_utils

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

var hashCache = make(map[string]HashMap)

func GetHash(path string) (HashMap, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if h, ok := hashCache[path]; ok {
		return h, nil
	}
	var bytes []byte
	if bytes, err = os.ReadFile(path); err == nil {
		var asMap map[string]interface{}
		if err = json.Unmarshal(bytes, &asMap); err != nil {
			return nil, err
		}

		res := make(HashMap, 0, len(asMap))

		for k, v := range asMap {
			var entry RationalTriplet
			if _, err = entry.value.SetInterface(v); err != nil {
				return nil, err
			}

			key := strings.Split(k, ",")

			switch len(key) {
			case 1:
				entry.key2Present = false
			case 2:
				entry.key2Present = true
				if _, err = entry.key2.SetInterface(key[1]); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("cannot parse %T as one or two field elements", v)
			}
			if _, err = entry.key1.SetInterface(key[0]); err != nil {
				return nil, err
			}

			res = append(res, &entry)
		}

		sort.Slice(res, func(i, j int) bool {
			return res[i].CmpKey(res[j]) <= 0
		})

		hashCache[path] = res

		return res, nil

	} else {
		return nil, err
	}
}

type HashMap []*RationalTriplet

type MapHashTranscript struct {
	HashMap         HashMap
	stateValid      bool
	resultAvailable bool
	state           small_rational.SmallRational
}

type RationalTriplet struct {
	key1        small_rational.SmallRational
	key2        small_rational.SmallRational
	key2Present bool
	value       small_rational.SmallRational
	used        bool
}

func (t *RationalTriplet) CmpKey(o *RationalTriplet) int {
	if cmp1 := t.key1.Cmp(&o.key1); cmp1 != 0 {
		return cmp1
	}

	if t.key2Present {
		if o.key2Present {
			return t.key2.Cmp(&o.key2)
		}
		return 1
	} else {
		if o.key2Present {
			return -1
		}
		return 0
	}
}

func (m HashMap) Hash(x *small_rational.SmallRational, y *small_rational.SmallRational) small_rational.SmallRational {

	toFind := RationalTriplet{
		key1:        *x,
		key2Present: y != nil,
	}

	if y != nil {
		toFind.key2 = *y
	}

	i := sort.Search(len(m), func(i int) bool { return m[i].CmpKey(&toFind) >= 0 })

	if i < len(m) && m[i].CmpKey(&toFind) == 0 {
		m[i].used = true
		return m[i].value
	}

	if y == nil {
		panic("No hash available for input " + x.Text(10))
	} else {
		panic("No hash available for input " + x.Text(10) + "," + y.Text(10))
	}
}

func (m HashMap) UnusedEntries() []interface{} {
	unused := make([]interface{}, 0)
	for _, v := range m {
		if !v.used {
			var vInterface interface{}
			if v.key2Present {
				vInterface = []interface{}{ElementToInterface(&v.key1), ElementToInterface(&v.key2)}
			} else {
				vInterface = ElementToInterface(&v.key1)
			}
			unused = append(unused, vInterface)
		}
	}
	return unused
}

func (m *MapHashTranscript) Update(i ...interface{}) {
	if len(i) > 0 {
		for _, x := range i {

			var xElement small_rational.SmallRational
			if _, err := xElement.SetInterface(x); err != nil {
				panic(err.Error())
			}
			if m.stateValid {
				m.state = m.HashMap.Hash(&xElement, &m.state)
			} else {
				m.state = m.HashMap.Hash(&xElement, nil)
			}

			m.stateValid = true
		}
	} else { //just hash the state itself
		if !m.stateValid {
			panic("nothing to hash")
		}
		m.state = m.HashMap.Hash(&m.state, nil)
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

func SliceToElementSlice(slice []interface{}) ([]small_rational.SmallRational, error) {
	elementSlice := make([]small_rational.SmallRational, len(slice))
	for i, v := range slice {
		if _, err := elementSlice[i].SetInterface(v); err != nil {
			return nil, err
		}
	}
	return elementSlice, nil
}

func SliceEquals(a []small_rational.SmallRational, b []small_rational.SmallRational) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch %d≠%d", len(a), len(b))
	}
	for i := range a {
		if !a[i].Equal(&b[i]) {
			return fmt.Errorf("at index %d: %s ≠ %s", i, a[i].String(), b[i].String())
		}
	}
	return nil
}

func ElementToInterface(x *small_rational.SmallRational) interface{} {
	text := x.Text(10)
	if len(text) < 10 && !strings.Contains(text, "/") {
		if res, err := strconv.Atoi(text); err != nil {
			panic("error: " + err.Error())
		} else {
			return res
		}
	}
	return text
}

func ElementSliceToInterfaceSlice(x interface{}) []interface{} {
	if x == nil {
		return nil
	}

	X := reflect.ValueOf(x)

	res := make([]interface{}, X.Len())
	for i := range res {
		xI := X.Index(i).Interface().(small_rational.SmallRational)
		res[i] = ElementToInterface(&xI)
	}
	return res
}

func ElementSliceSliceToInterfaceSliceSlice(x interface{}) [][]interface{} {
	if x == nil {
		return nil
	}

	X := reflect.ValueOf(x)

	res := make([][]interface{}, X.Len())
	for i := range res {
		res[i] = ElementSliceToInterfaceSlice(X.Index(i).Interface())
	}

	return res
}