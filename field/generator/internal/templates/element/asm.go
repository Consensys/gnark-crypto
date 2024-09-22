package element

// Asm ...
const Asm = `
import 	"golang.org/x/sys/cpu"

var (
	supportAdx = cpu.X86.HasADX && cpu.X86.HasBMI2
	_ = supportAdx
	{{- if eq .NbWords 4}}
	supportAvx512 = cpu.X86.HasAVX512 && cpu.X86.HasAVX512DQ
	_ = supportAvx512
	{{- end}}
)
`

// AsmNoAdx ...
const AsmNoAdx = `

// note: this is needed for test purposes, as dynamically changing supportAdx doesn't flag
// certain errors (like fatal error: missing stackmap)
// this ensures we test all asm path.
var (
	supportAdx = false
	_ = supportAdx
	{{- if eq .NbWords 4}}
	supportAvx512 = false
	_ = supportAvx512
	{{- end}}
)
`
