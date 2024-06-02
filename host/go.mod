module github.com/cjr29/cpu1-simulator/host

go 1.22.3

require (
	github.com/beevik/cmd v0.2.0
	github.com/beevik/prefixtree v1.0.1
	github.com/cjr29/cpu1-simulator/asm v0.0.0
	github.com/cjr29/cpu1-simulator/cpu1 v0.0.0
	github.com/cjr29/cpu1-simulator/disasm v0.0.0
	github.com/cjr29/cpu1-simulator/term v0.0.0

)

require (
	github.com/cjr29/go6502/cpu v0.0.0-20240601120914-7d2368808de5 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace github.com/cjr29/cpu1-simulator/cpu1 v0.0.0 => ../cpu1

replace github.com/cjr29/cpu1-simulator/asm v0.0.0 => ../asm

replace github.com/cjr29/cpu1-simulator/term v0.0.0 => ../term

replace github.com/cjr29/cpu1-simulator/disasm v0.0.0 => ../disasm
