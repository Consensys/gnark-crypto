// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) Loop(n amd64.Register, fn func()) {
	lblLoop := f.NewLabel("loop")
	lblDone := f.NewLabel("done")

	// while n > 0, do:
	f.LABEL(lblLoop)
	f.TESTQ(n, n)
	f.JEQ(lblDone)
	f.DECQ(n)

	fn()

	f.JMP(lblLoop)
	f.LABEL(lblDone)
}

// LabelRegisters write comment with friendler name to registers
func (f *FFAmd64) LabelRegisters(name string, r ...amd64.Register) {
	switch len(r) {
	case 0:
		return
	case 1:
		f.Comment(fmt.Sprintf("%s -> %s", name, string(r[0])))
	default:
		for i := 0; i < len(r); i++ {
			f.Comment(fmt.Sprintf("%s[%d] -> %s", name, i, string(r[i])))
		}
	}
	// f.WriteLn("")
}

func (f *FFAmd64) ReduceElement(t, scratch []amd64.Register, avoidGlobal bool) {
	var buf bytes.Buffer
	var err error
	if f.qStack == nil && avoidGlobal {
		if len(scratch) != len(t)+1 {
			panic(fmt.Sprintf("expected %d scratch registers, got %d", len(t)+1, len(scratch)))
		}
		const tmplReduce = `// reduce element({{- range $i, $a := .A}}{{$a}}{{- if ne $.Last $i}},{{ end}}{{- end}}) using temp registers ({{- range $i, $b := .B}}{{$b}}{{- if ne $.Last $i}},{{ end}}{{- end}})
	REDUCE_NOGLOBAL({{- range $i, $a := .A}}{{$a}},{{- end}}
		{{- range $i, $b := .B}}{{$b}}{{- if ne $.Last $i}},{{ end}}{{- end}})`

		err = template.Must(template.New("").
			Parse(tmplReduce)).Execute(&buf, struct {
			A, B []amd64.Register
			Last int
		}{t, scratch, len(scratch) - 1})
	} else {
		scratch = scratch[:len(t)] // ensure we do not use more scratch registers than t

		// we can use f.qAt(...)
		var Q []string
		for i := 0; i < f.NbWords; i++ {
			Q = append(Q, f.qAt(i))
		}

		const tmplReduce = `// reduce element({{- range $i, $a := .A}}{{$a}}{{- if ne $.Last $i}},{{ end}}{{- end}}) using temp registers ({{- range $i, $b := .B}}{{$b}}{{- if ne $.Last $i}},{{ end}}{{- end}})
	REDUCE({{- range $i, $a := .A}}{{$a}},{{- end}}
		{{- range $i, $b := .B}}{{$b}},{{- end}}
		{{- range $i, $q := .Q}}{{$q}}{{- if ne $.Last $i}},{{ end}}{{- end}})`

		err = template.Must(template.New("").
			Parse(tmplReduce)).Execute(&buf, struct {
			A, B []amd64.Register
			Q    []string
			Last int
		}{t, scratch, Q, len(scratch) - 1})
	}

	if err != nil {
		panic(err)
	}

	f.WriteLn(buf.String())
	f.WriteLn("")
}

// This template uses ·qElement (global variable)
const tmplReduceDefine = `

#define REDUCE(	{{- range $i := .NbWordsIndexesFull}}ra{{$i}},{{- end}}
				{{- range $i := .NbWordsIndexesFull}}rb{{$i}},{{- end}}
				{{- range $i := .NbWordsIndexesFull}}q{{$i}}{{- if ne $.NbWordsLastIndex $i}},{{- end}}{{- end}}) \
	MOVQ ra0, rb0;  \
	SUBQ    q0, ra0; \
	{{- range $i := .NbWordsIndexesNoZero}}
	MOVQ ra{{$i}}, rb{{$i}};  \
	SBBQ  q{{$i}}, ra{{$i}}; \
	{{- end}}
	{{- range $i := .NbWordsIndexesFull}}
	CMOVQCS rb{{$i}}, ra{{$i}};  \
	{{- end}}
`

// This template uses a spare scratch register and uses define (const_q0, const_q1, etc.)
// it allows to avoid global variable usage (no R15 clobbering)
const tmplReduceDefineNoGlobal = `

#define REDUCE_NOGLOBAL(	{{- range $i := .NbWordsIndexesFull}}ra{{$i}},{{- end}}
				{{- range $i := .NbWordsIndexesFull}}rb{{$i}},{{- end}}, scratch0) \
	MOVQ ra0, rb0;  \
	MOVQ $const_q0, scratch0;  \
	SUBQ    scratch0, ra0; \
	{{- range $i := .NbWordsIndexesNoZero}}
	MOVQ ra{{$i}}, rb{{$i}};  \
	MOVQ $const_q{{$i}}, scratch0;  \
	SBBQ  scratch0, ra{{$i}}; \
	{{- end}}
	{{- range $i := .NbWordsIndexesFull}}
	CMOVQCS rb{{$i}}, ra{{$i}};  \
	{{- end}}
`

func (f *FFAmd64) GenerateReduceDefine(avoidGlobal ...bool) {
	tmplReduceDefine := tmplReduceDefine
	if len(avoidGlobal) > 0 && avoidGlobal[0] {
		tmplReduceDefine = tmplReduceDefineNoGlobal
	}
	tmpl := template.Must(template.New("").
		Funcs(helpers()).
		Parse(tmplReduceDefine))

	// execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, f); err != nil {
		panic(err)
	}

	f.WriteLn(buf.String())
}

func (f *FFAmd64) Mov(i1, i2 interface{}, offsets ...int) {
	var o1, o2 int
	if len(offsets) >= 1 {
		o1 = offsets[0]
		if len(offsets) >= 2 {
			o2 = offsets[1]
		}
	}
	switch c1 := i1.(type) {
	case []uint64:
		switch c2 := i2.(type) {
		default:
			panic("unsupported")
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(c1[i+o1], c2[i+o2])
			}
		}
	case amd64.Register:
		switch c2 := i2.(type) {
		case amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(c1.At(i+o1), c2.At(i+o2))
			}
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(c1.At(i+o1), c2[i+o2])
			}
		default:
			panic("unsupported")
		}
	case []amd64.Register:
		switch c2 := i2.(type) {
		case amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(c1[i+o1], c2.At(i+o2))
			}
		case []amd64.Register:
			// f.copyElement(c1[o1:], c2[o2:])
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(c1[i+o1], c2[i+o2])
			}
		default:
			panic("unsupported")
		}
	default:
		panic("unsupported")
	}

}

func (f *FFAmd64) Add(i1, i2 interface{}, offsets ...int) {
	var o1, o2 int
	if len(offsets) >= 1 {
		o1 = offsets[0]
		if len(offsets) >= 2 {
			o2 = offsets[1]
		}
	}
	switch c1 := i1.(type) {

	case amd64.Register:
		switch c2 := i2.(type) {
		default:
			panic("unsupported")
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				if i == 0 {
					f.ADDQ(c1.At(i+o1), c2[i+o2])
				} else {
					f.ADCQ(c1.At(i+o1), c2[i+o2])
				}
			}
		}
	case []amd64.Register:
		switch c2 := i2.(type) {
		default:
			panic("unsupported")
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				if i == 0 {
					f.ADDQ(c1[i+o1], c2[i+o2])
				} else {
					f.ADCQ(c1[i+o1], c2[i+o2])
				}
			}
		}
	default:
		panic("unsupported")
	}
}

func (f *FFAmd64) Sub(i1, i2 interface{}, offsets ...int) {
	var o1, o2 int
	if len(offsets) >= 1 {
		o1 = offsets[0]
		if len(offsets) >= 2 {
			o2 = offsets[1]
		}
	}
	switch c1 := i1.(type) {

	case amd64.Register:
		switch c2 := i2.(type) {
		default:
			panic("unsupported")
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				if i == 0 {
					f.SUBQ(c1.At(i+o1), c2[i+o2])
				} else {
					f.SBBQ(c1.At(i+o1), c2[i+o2])
				}
			}
		}
	case []amd64.Register:
		switch c2 := i2.(type) {
		default:
			panic("unsupported")
		case []amd64.Register:
			for i := 0; i < f.NbWords; i++ {
				if i == 0 {
					f.SUBQ(c1[i+o1], c2[i+o2])
				} else {
					f.SBBQ(c1[i+o1], c2[i+o2])
				}
			}
		}
	default:
		panic("unsupported")
	}
}

// Template helpers (txt/template)
func helpers() template.FuncMap {
	// functions used in template
	return template.FuncMap{
		"mul": mul,
		"imm": imm,
		"sub": sub,
	}
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

func imm(t uint64) string {
	switch t {
	case 0:
		return "$0"
	case 1:
		return "$1"
	default:
		return fmt.Sprintf("$%#016x", t)
	}
}
