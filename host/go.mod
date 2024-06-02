module github.com/cjr29/cpu1-simulator/host

go 1.22.3

require (
    github.com/cjr29/cpu1-simulator/cpu1 v0.0.0
	github.com/cjr29/cpu1-simulator/asm v0.0.0
	github.com/cjr29/cpu1-simulator/term v0.0.0
	github.com/cjr29/cpu1-simulator/disasm v0.0.0

)

replace github.com/cjr29/cpu1-simulator/cpu1 v0.0.0 => ../cpu1
replace github.com/cjr29/cpu1-simulator/asm v0.0.0 => ../asm
replace github.com/cjr29/cpu1-simulator/term v0.0.0 => ../term
replace github.com/cjr29/cpu1-simulator/disasm v0.0.0 => ../disasm
