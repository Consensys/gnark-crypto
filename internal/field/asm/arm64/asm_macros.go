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

package arm64

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/consensys/bavard/arm64"
)

// LabelRegisters write comment with friendler name to registers
func (f *FFArm64) LabelRegisters(name string, r ...arm64.Register) {
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

// TODO @gbotrel: figure out if interleaving MOVQ and SUBQ or CMOVQ and MOVQ instructions makes sense
const tmplDefines = `

// modulus q
{{- range $i, $w := .Q}}
DATA q<>+{{mul $i 8}}(SB)/8, {{imm $w}}
{{- end}}
GLOBL q<>(SB), (RODATA+NOPTR), ${{mul 8 $.NbWords}}

`

// func (f *FFArm64) reduce(t, v []arm64.Register) {

// 	f.MOVD(f.qAt(0), v[0])
// 	f.SUBS(v[0], t[0], v[0])
// 	for i := 1; i < f.NbWords; i++ {
// 		f.MOVD(f.qAt(i), v[i])
// 		f.SBCS(v[i], t[i], v[i])
// 	}

// 	for i := 0; i < f.NbWords; i++ {
// 		f.CSEL("CS", v[i], t[i], t[i])
// 	}
// }

func (f *FFArm64) GenerateDefines() {
	tmpl := template.Must(template.New("").
		Funcs(helpers()).
		Parse(tmplDefines))

	// execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, f); err != nil {
		panic(err)
	}

	f.WriteLn(buf.String())
}

func (f *FFArm64) Mov(i1, i2 interface{}, offsets ...int) {
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
		case []arm64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVD(c1[i+o1], c2[i+o2])
			}
		}
	case arm64.Register:
		switch c2 := i2.(type) {
		case arm64.Register:
			for i := 0; i < f.NbWords; i++ {
				f.MOVD(c1.At(i+o1), c2.At(i+o2))
			}
		case []arm64.Register:
			// for i := 0; i < f.NbWords; i++ {
			// 	f.MOVD(c1.At(i+o1), c2[i+o2])
			// }
			for i := 0; i < f.NbWords-1; i += 2 {
				f.LDP(c1.At(i+o1), c2[i+o2], c2[i+o2+1] /*, fmt.Sprintf("%s, %s = q[%d], q[%d]", q[i].Name(), q[i+1].Name(), i, i+1)*/)
			}
			if f.NbWords%2 == 1 {
				i := f.NbWords - 1
				f.MOVD(c1.At(i+o1), c2[i+o2] /*, fmt.Sprintf("%s = q[%d]", q[i].Name(), i)*/)
			}
		default:
			panic("unsupported")
		}
	case []arm64.Register:
		switch c2 := i2.(type) {
		case arm64.Register:
			// for i := 0; i < f.NbWords; i++ {
			// 	f.MOVD(c1[i+o1], c2.At(i+o2))
			// }
			for i := 0; i < f.NbWords-1; i += 2 {
				f.STP(c1[i+o1], c1[i+o1+1], c2.At(i+o2))
			}
			if f.NbWords%2 == 1 {
				i := f.NbWords - 1
				f.MOVD(c1[i+o1], c2.At(i+o2))
			}
		case []arm64.Register:
			// f.copyElement(c1[o1:], c2[o2:])
			for i := 0; i < f.NbWords; i++ {
				f.MOVD(c1[i+o1], c2[i+o2])
			}
		default:
			panic("unsupported")
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

// madd0 (hi, -) = a*b + c
func (f *FFArm64) madd0(hi, a, b, c, TMP interface{}) {
	f.UMULH(a, b, hi)
	f.MUL(a, b, TMP)
	f.ADDS(c, TMP, "ZR")
	f.ADC("ZR", hi, hi)
}

// madd1 (hi, lo) = a*b + c
func (f *FFArm64) madd1(hi, lo, a, b, c, TMP interface{}) {
	f.MUL(a, b, TMP)
	f.ADDS(c, TMP, lo)
	f.UMULH(a, b, hi)
	f.ADC("ZR", hi, hi)
}

// madd2 (hi, lo) = a*b + c + d
func (f *FFArm64) madd2(hi, lo, a, b, c, d, TMP interface{}) {
	f.MUL(a, b, TMP)
	f.ADDS(c, TMP, TMP)
	f.UMULH(a, b, hi)
	f.ADC("ZR", hi, hi)
	f.ADDS(d, TMP, lo)
	f.ADC("ZR", hi, hi)
}

func (f *FFArm64) madd4(hi, lo, a, b, c, d, e, TMP, TMP2 interface{}) {
	f.MUL(a, b, TMP)
	f.UMULH(a, b, TMP2)
	f.ADDS(c, TMP, TMP)
	f.ADC("ZR", TMP2, TMP2)
	f.ADDS(d, TMP, lo)
	f.ADC(e, TMP2, hi)
}

func (f *FFArm64) callTemplate(templateName string, ops ...interface{}) {
	f.Write(templateName)
	f.Write("(")
	for i := 0; i < len(ops); i++ {
		f.Write(arm64.Operand(ops[i]))
		if i+1 < len(ops) {
			f.Write(", ")
		}
	}
	f.WriteLn(")")
}
