package common

import (
	"fmt"
	"reflect"
	"text/template"
)

// Funcs returns the common functions for the templates
func Funcs() template.FuncMap {
	return template.FuncMap{
		"shorten": func(input string) string {
			if len(input) > 15 {
				return input[:6] + "..." + input[len(input)-6:]
			}
			return input
		},
		"ltu64": func(a, b uint64) bool {
			return a < b
		},
		"last": func(x int, a any) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"add": func(a, b any) int {
			return anyToInt(a) + anyToInt(b)
		},
		"sub": func(a, b any) int {
			return anyToInt(a) - anyToInt(b)
		},
		"mul": func(a, b any) int {
			return anyToInt(a) * anyToInt(b)
		},
		"div": func(a, b any) int {
			return anyToInt(a) / anyToInt(b)
		},
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"iterate": func(start, end any) []int {
			s := anyToInt(start)
			e := anyToInt(end)
			if e <= s {
				return nil
			}
			res := make([]int, e-s)
			for i := range res {
				res[i] = s + i
			}
			return res
		},
		"reverse": func(input any) any {
			rv := reflect.ValueOf(input)
			if rv.Kind() != reflect.Slice {
				return input
			}
			l := rv.Len()
			res := reflect.MakeSlice(rv.Type(), l, l)
			for i := range l {
				res.Index(l - 1 - i).Set(rv.Index(i))
			}
			return res.Interface()
		},
	}
}

func anyToInt(i any) int {
	switch v := i.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case uint64:
		return int(v)
	case int32:
		return int(v)
	case uint32:
		return int(v)
	default:
		return 0
	}
}
