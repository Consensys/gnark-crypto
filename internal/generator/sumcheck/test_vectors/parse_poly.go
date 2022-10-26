package test_vectors

/*
import (
	"fmt"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"strings"
	"unicode"
)

type stringReader struct {
	index int
	str   []rune
}

func (r *stringReader) peek() rune {
	return r.str[r.index]
}

func (r *stringReader) next() rune {
	res := r.peek()
	r.index++
	return res
}

func (r *stringReader) end() bool {
	return r.index >= len(r.str)
}

func (r *stringReader) increment() {
	r.index++
}

func (r *stringReader) skipSpaces() {
	for unicode.IsSpace(r.peek()) {
		r.increment()
	}
}

func (r *stringReader) peekAlphaNumeric() (rune, bool) {
	res := r.peek()
	if unicode.IsLetter(res) || unicode.IsDigit(res) || res == '_' {
		return res, true
	}
	return res, false
}

func (r *stringReader) expect(categoryName string, oneOf ...func(rune) bool) (rune, error) {
	res := r.next()

	for _, f := range oneOf {
		if f(res) {
			return res, nil
		}
	}

	return ' ', fmt.Errorf("at index %d expected %s but got \"%c\"", r.index-1, categoryName, res)
}

func (r *stringReader) expectLetter() (rune, error) {
	return r.expect("letter", unicode.IsLetter)
}

func (r *stringReader) expectDigit() (rune, error) {
	return r.expect("digit", unicode.IsDigit)
}

func (r *stringReader) expectLetterOrDigit() (rune, error) {
	return r.expect("alphanumeric", unicode.IsLetter, unicode.IsDigit, runeEquals('_')) //TODO: Allow subscripts
}

func (r *stringReader) expectIdentifier() (string, error) {

	const errorFormat = "identifier error: %s"

	var sb strings.Builder

	if c, err := r.expectLetter(); err == nil {
		sb.WriteRune(c)
	} else {
		return "", fmt.Errorf(errorFormat, err.Error())
	}

	for {
		if c, ok := r.peekAlphaNumeric(); ok {
			sb.WriteRune(c)
			r.increment()
		} else {
			break
		}
	}
	return sb.String(), nil
}

//func expectInt()

func runeEquals(r rune) func(rune) bool {
	return func(x rune) bool {
		return r == x
	}
}

func defineVariable(reader *stringReader, varsMap map[string]int) error {
	reader.skipSpaces()
	i := reader.index
	if s, err := reader.expectIdentifier(); err == nil {
		if _, found := varsMap[s]; found {
			return fmt.Errorf("variable \"%s\" redeclared at %d", s, i)
		}
		varsMap[s] = len(varsMap)
	}
	return nil
}

func parseVariables(reader *stringReader) (map[string]int, error) {

	varsMap := make(map[string]int)

	reader.skipSpaces()
	if reader.peek() == '↦' {
		reader.increment()
		return varsMap, nil
	}

	if err := defineVariable(varsMap, reader); err != nil {
		return nil, err
	}

	reader.skipSpaces()

	for {
		if reader.peek() == '↦' {
			reader.increment()
			return varsMap, nil
		}
		if _, err := reader.expect("\",\"", runeEquals(',')); err != nil {
			return nil, err
		}
		if err := defineVariable(varsMap, reader); err != nil {
			return nil, err
		}
		reader.skipSpaces()
	}
}

type monomial struct {
	coeff  small_rational.SmallRational
	powers []int
}

func readInt(reader *stringReader) (string, error) {
	var sb strings.Builder
	reader.skipSpaces()
	if c := reader.peek(); c == '+' || c == '-' {
		sb.WriteRune(c)
		reader.skipSpaces()
	}

	if c, err := reader.expectDigit(); err != nil {
		return "", err
	} else {
		sb.WriteRune(c)
	}

	for unicode.IsDigit(reader.peek()) {
		sb.WriteRune(reader.next())
	}

	return sb.String(), nil
}

func parseNumber(reader *stringReader) (*small_rational.SmallRational, error) {
	var res small_rational.SmallRational
	nom, err := readInt(reader)
	if err != nil {
		return &res, err
	}
	reader.skipSpaces()
	if reader.peek() == '/' {
		var denom string
		denom, err = readInt(reader)
		if err != nil {
			return &res, err
		}
		return res.SetInterface(nom + "/" + denom)
	} else {
		return res.SetInterface(nom)
	}
}

func parseMonomialVars(reader *stringReader, varsMap map[string]int) {

}

func parseMonomial(reader *stringReader, varsMap map[string]int) (monomial, error) {
	var name string
	var err error
	var res monomial

	if unicode.IsLetter(reader.peek()) {
		if name, err = reader.expectIdentifier(); err != nil {
			return res, err
		}

	}
}

func ParsePolynomial(str string) ([]monomial, error) {
	reader := stringReader{
		index: 0,
		str:   []rune(str),
	}

	variables, err := parseVariables(&reader)

	if err != nil {
		return nil, err
	}

}*/
