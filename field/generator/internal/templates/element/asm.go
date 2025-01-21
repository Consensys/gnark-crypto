package element

// Asm ...
const Asm = `
import 	"golang.org/x/sys/cpu"

var (
	supportAdx = cpu.X86.HasADX && cpu.X86.HasBMI2
	_ = supportAdx
)
`

const Avx = `
import 	"golang.org/x/sys/cpu"

var (
	supportAvx512 = {{- if not .F31 }}supportAdx && {{- end}}cpu.X86.HasAVX512 && cpu.X86.HasAVX512DQ  && cpu.X86.HasAVX512VBMI2
	_ = supportAvx512
)
`

const NoAvx = `
const supportAvx512 = false
`

// AsmNoAdx ...
const AsmNoAdx = `

// note: this is needed for test purposes, as dynamically changing supportAdx doesn't flag
// certain errors (like fatal error: missing stackmap)
// this ensures we test all asm path.
var (
	supportAdx = false
	_ = supportAdx
)
`
