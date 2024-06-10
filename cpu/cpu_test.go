package cpu_test

import (
	"os"
	"strings"
	"testing"

	"riddick.net/cpu1-simulator/asm"
	"riddick.net/cpu1-simulator/cpu"
)

func loadCPU(t *testing.T, asmString string) *cpu.CPU {
	b := strings.NewReader(asmString)
	r, sm, err := asm.Assemble(b, "test.asm", 0x1000, os.Stdout, 0)
	if err != nil {
		t.Error(err)
		return nil
	}

	mem := cpu.NewFlatMemory()
	cpu := cpu.NewCPU(cpu.NMOS, mem)
	mem.StoreBytes(sm.Origin, r.Code)
	cpu.SetPC(sm.Origin)
	return cpu
}

func stepCPU(cpu *cpu.CPU, steps int) {
	for i := 0; i < steps; i++ {
		cpu.Step()
	}
}

func runCPU(t *testing.T, asmString string, steps int) *cpu.CPU {
	cpu := loadCPU(t, asmString)
	if cpu != nil {
		stepCPU(cpu, steps)
	}
	return cpu
}

func expectPC(t *testing.T, cpu *cpu.CPU, pc uint16) {
	if cpu.Reg.PC != pc {
		t.Errorf("PC incorrect. exp: $%04X, got: $%04X", pc, cpu.Reg.PC)
	}
}

func expectQ(t *testing.T, cpu *cpu.CPU, b byte) {
	if cpu.Reg.Q != b {
		t.Errorf("Q incorrect. exp: $%02X, got: $%02X", b, cpu.Reg.Q)
	}
}

func expectCycles(t *testing.T, cpu *cpu.CPU, cycles uint64) {
	if cpu.Cycles != cycles {
		t.Errorf("Cycles incorrect. exp: %d, got: %d", cycles, cpu.Cycles)
	}
}

/* func expectACC(t *testing.T, cpu *cpu.CPU, acc byte) {
	if cpu.Reg.A != acc {
		t.Errorf("Accumulator incorrect. exp: $%02X, got: $%02X", acc, cpu.Reg.A)
	}
} */

func expectSP(t *testing.T, cpu *cpu.CPU, sp byte) {
	if cpu.Reg.SP != sp {
		t.Errorf("stack pointer incorrect. exp: %02X, got $%02X", sp, cpu.Reg.SP)
	}
}

func expectR(t *testing.T, cpu *cpu.CPU, acc byte, reg byte) {
	if cpu.Reg.R[reg] != acc {
		t.Errorf("Register incorrect. exp: $%02X, got: $%02X", acc, cpu.Reg.R[reg])
	}
}

func expectMem(t *testing.T, cpu *cpu.CPU, addr uint16, v byte) {
	got := cpu.Mem.LoadByte(addr)
	if got != v {
		t.Errorf("Memory at $%04X incorrect. exp: $%02X, got: $%02X", addr, v, got)
	}
}

func TestReg0(t *testing.T) {
	asm := `
	.ORG $1000
	LDI0 #$5E
	STI0 $1500`

	cpu := runCPU(t, asm, 2) // 2 steps
	if cpu == nil {
		return
	}

	expectPC(t, cpu, 0x1005)
	expectCycles(t, cpu, 6)
	expectR(t, cpu, 0x5e, 0)
	expectMem(t, cpu, 0x1500, 0x5e)
}

func TestStack(t *testing.T) {
	asm := `
	.ORG $1000
	LDI0 #$11
	PUSH0
	LDI0 #$12
	PUSH0
	LDI0 #$13
	PUSH0

	POP0
	STI0 $2000
	POP0
	STI0 $2001
	POP0
	STI0 $2002`

	cpu := loadCPU(t, asm)
	stepCPU(cpu, 6)

	expectSP(t, cpu, 0xfc)
	expectR(t, cpu, 0x13, 0)
	expectMem(t, cpu, 0x1ff, 0x11)
	expectMem(t, cpu, 0x1fe, 0x12)
	expectMem(t, cpu, 0x1fd, 0x13)

	stepCPU(cpu, 6)
	expectR(t, cpu, 0x11, 0)
	expectSP(t, cpu, 0xff)
	expectMem(t, cpu, 0x2000, 0x13)
	expectMem(t, cpu, 0x2001, 0x12)
	expectMem(t, cpu, 0x2002, 0x11)
}

// Test Q instructions
func TestQ(t *testing.T) {
	asm := `
	.ORG $1000
	SETQ0
	SETQ1
	SETQ2
	SETQ3
	SETQ4
	SETQ5
	SETQ6
	SETQ7`

	cpu := runCPU(t, asm, 8)
	if cpu == nil {
		return
	}

	expectPC(t, cpu, 0x1008)
	expectCycles(t, cpu, 8)
	expectQ(t, cpu, 0xff)
}

func TestUnusedCPU1(t *testing.T) {
	asm := `
	.ORG $1000
	.ARCH CPU1
	.DH 06
	.DH 07
	.DH 1c
	.DH 1d
	.DH 1e`

	cpu := runCPU(t, asm, 5)
	if cpu == nil {
		return
	}

	expectPC(t, cpu, 0x1005)
	expectCycles(t, cpu, 5)
}

// Test arithmetic
func TestArithmetic(t *testing.T) {
	asm := `
	.ORG $1000
	.ARCH CPU1
	LDI0	#$11
	LDI1	#$01
	ADR 	#$01
	`
	cpu := loadCPU(t, asm)
	stepCPU(cpu, 3)

	expectPC(t, cpu, 0x1006)
	expectCycles(t, cpu, 7)
	expectR(t, cpu, 0x12, 0)

}
