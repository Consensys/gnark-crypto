package element

const TestVector = `


import (
	"testing"
	"github.com/stretchr/testify/require"
	"sort"
	"reflect"
	"bytes"
	"fmt"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestVectorSort(t *testing.T) {
	assert := require.New(t)

	v := make(Vector, 3)
	v[0].SetUint64(2)
	v[1].SetUint64(3)
	v[2].SetUint64(1)

	sort.Sort(v)

	assert.Equal("[1,2,3]", v.String())
}

func TestVectorRoundTrip(t *testing.T) {
	assert := require.New(t)

	v1 := make(Vector, 3)
	v1[0].SetUint64(2)
	v1[1].SetUint64(3)
	v1[2].SetUint64(1)

	b, err := v1.MarshalBinary()
	assert.NoError(err)

	var v2,v3 Vector

	err = v2.UnmarshalBinary(b)
	assert.NoError(err)

	err = v3.unmarshalBinaryAsync(b)
	assert.NoError(err)

	assert.True(reflect.DeepEqual(v1,v2))
	assert.True(reflect.DeepEqual(v3,v2))
}

func TestVectorEmptyRoundTrip(t *testing.T) {
	assert := require.New(t)

	v1 := make(Vector, 0)

	b, err := v1.MarshalBinary()
	assert.NoError(err)

	var v2, v3 Vector

	err = v2.UnmarshalBinary(b)
	assert.NoError(err)

	err = v3.unmarshalBinaryAsync(b)
	assert.NoError(err)

	assert.True(reflect.DeepEqual(v1,v2))
	assert.True(reflect.DeepEqual(v3,v2))
}

func (vector *Vector) unmarshalBinaryAsync(data []byte) error {
	r := bytes.NewReader(data)
	_, err, chErr := vector.AsyncReadFrom(r)
	if err != nil {
		return err
	}
	return <-chErr
}



func TestVectorOps(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = 5
	} else {
		parameters.MinSuccessfulTests = 100
	}
	properties := gopter.NewProperties(parameters)

	addVector := func(a, b Vector) bool {
		c := make(Vector, len(a))
		c.Add(a, b)
		
		for i := 0; i < len(a); i++ {
			var tmp {{.ElementName}}
			tmp.Add(&a[i], &b[i])
			if !tmp.Equal(&c[i]) {
				return false
			}
		}
		return true
	}

	subVector := func(a, b Vector) bool {
		c := make(Vector, len(a))
		c.Sub(a, b)
		
		for i := 0; i < len(a); i++ {
			var tmp {{.ElementName}}
			tmp.Sub(&a[i], &b[i])
			if !tmp.Equal(&c[i]) {
				return false
			}
		}
		return true
	}

	scalarMulVector := func(a Vector, b {{.ElementName}}) bool {
		c := make(Vector, len(a))
		c.ScalarMul(a, &b)
		
		for i := 0; i < len(a); i++ {
			var tmp {{.ElementName}}
			tmp.Mul(&a[i], &b)
			if !tmp.Equal(&c[i]) {
				return false
			}
		}
		return true
	}

	sumVector := func(a Vector) bool {
		var sum {{.ElementName}}
		computed := a.Sum()
		for i := 0; i < len(a); i++ {
			sum.Add(&sum, &a[i])
		}

		return sum.Equal(&computed)
	}

	innerProductVector := func(a, b Vector) bool {
		computed := a.InnerProduct(b)
		var innerProduct {{.ElementName}}
		for i := 0; i < len(a); i++ {
			var tmp {{.ElementName}}
			tmp.Mul(&a[i], &b[i])
			innerProduct.Add(&innerProduct, &tmp)
		}

		return innerProduct.Equal(&computed)
	}

	sizes := []int{1, 2, 3, 4, 7, 9, 64, 65, 117, 127, 128, 129, 130, 131, 1024}

	for _, size := range sizes {
		properties.Property(fmt.Sprintf("vector addition %d", size), prop.ForAll(
			addVector,
			genVector(size),
			genVector(size),
		))

		properties.Property(fmt.Sprintf("vector subtraction %d", size), prop.ForAll(
			subVector,
			genVector(size),
			genVector(size),
		))

		properties.Property(fmt.Sprintf("vector scalar multiplication %d", size), prop.ForAll(
			scalarMulVector,
			genVector(size),
			gen{{.ElementName}}(),
		))

		properties.Property(fmt.Sprintf("vector sum %d", size), prop.ForAll(
			sumVector,
			genVector(size),
		))

		properties.Property(fmt.Sprintf("vector inner product %d", size), prop.ForAll(
			innerProductVector,
			genVector(size),
			genVector(size),
		))

	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func BenchmarkVectorOps(b *testing.B) {
	// note; to benchmark against "no asm" version, use the following
	// build tag: -tags purego
	const N = 1<<20
	a1 := make(Vector, N)
	b1 := make(Vector, N)
	c1 := make(Vector, N)
	var mixer {{.ElementName}}
	mixer.SetRandom()
	for i := 1; i < N; i++ {
		a1[i-1].SetUint64(uint64(i)).
			Mul(&a1[i-1], &mixer)
		b1[i-1].SetUint64(^uint64(i)).
			Mul(&b1[i-1], &mixer)
	}

	b.Run("Add", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c1.Add(a1, b1)
		}
	})

	b.Run("Sub", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c1.Sub(a1, b1)
		}
	})

	b.Run("ScalarMul", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c1.ScalarMul(a1, &b1[0])
		}
	})

	b.Run("Sum", func(b *testing.B) {
		b.ResetTimer()
		var sum {{.ElementName}}
		for i := 0; i < b.N; i++ {
			sum = c1.Sum()
		}
		_ = sum
	})

	b.Run("InnerProduct", func(b *testing.B) {
		b.ResetTimer()
		var innerProduct {{.ElementName}}
		for i := 0; i < b.N; i++ {
			innerProduct = a1.InnerProduct(b1)
		}
		_ = innerProduct
	})
}

func genVector(size int) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		g := make(Vector, size)
		mixer := {{.ElementName}}{
			{{- range $i := .NbWordsIndexesFull}}
			genParams.NextUint64(),{{end}}
		}
		if qElement[{{.NbWordsLastIndex}}] != ^uint64(0) {
			mixer[{{.NbWordsLastIndex}}] %= (qElement[{{.NbWordsLastIndex}}] +1 )
		}
		

		for !mixer.smallerThanModulus() {
			mixer = {{.ElementName}}{
				{{- range $i := .NbWordsIndexesFull}}
				genParams.NextUint64(),{{end}}
			}
			if qElement[{{.NbWordsLastIndex}}] != ^uint64(0) {
				mixer[{{.NbWordsLastIndex}}] %= (qElement[{{.NbWordsLastIndex}}] +1 )
			}
		}

		for i := 1; i <= size; i++ {
			g[i-1].SetUint64(uint64(i)).
				Mul(&g[i-1], &mixer)
		}

		genResult := gopter.NewGenResult(g, gopter.NoShrinker)
		return genResult
	}
}

`
