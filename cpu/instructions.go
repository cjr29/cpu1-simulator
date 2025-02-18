// Copyright 2014-2018 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
*	Adapted by Chris Riddick, 2024 for new instruction set
 */

package cpu

import (
	"strings"
)

// An opsym is an internal symbol used to associate an opcode's data
// with its instructions.
type opsym byte

const (
	symADI0 opsym = iota
	symADI1
	symADI2
	symADI3
	symADI4
	symADI5
	symADI6
	symADI7
	symADIC
	symADM
	symADMC
	symADR
	symADRC
	symAND
	symANI
	symCALL
	symCMP
	symCPSR
	symDEC
	symEX
	symHALT
	symINC
	symLBR
	symLBRC
	symLBRQ
	symLBRZ
	symLDI0
	symLDI1
	symLDI2
	symLDI3
	symLDI4
	symLDI5
	symLDI6
	symLDI7
	symLDM
	symNOP
	symOR
	symORI
	symPOP0
	symPOP1
	symPOP2
	symPOP3
	symPOP4
	symPOP5
	symPOP6
	symPOP7
	symPUSH0
	symPUSH1
	symPUSH2
	symPUSH3
	symPUSH4
	symPUSH5
	symPUSH6
	symPUSH7
	symRESETQ0
	symRESETQ1
	symRESETQ2
	symRESETQ3
	symRESETQ4
	symRESETQ5
	symRESETQ6
	symRESETQ7
	symRET
	symSETQ0
	symSETQ1
	symSETQ2
	symSETQ3
	symSETQ4
	symSETQ5
	symSETQ6
	symSETQ7
	symSHL
	symSHLC
	symSHR
	symSHRC
	symSPSR
	symSTI0
	symSTI1
	symSTI2
	symSTI3
	symSTI4
	symSTI5
	symSTI6
	symSTI7
	symSUB
	symSUBC
	symSUBI
	symSUBIC
	symSUBM
	symSUBMC
	symXOR
	symXRI
)

type instfunc func(c *CPU, inst *Instruction, operand []byte)

// Emulator implementation for each opcode
type opcodeImpl struct {
	sym  opsym
	name string
	fn   [2]instfunc // NMOS=0, CMOS=1
}

var impl = []opcodeImpl{
	{symADI0, "ADI0", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI1, "ADI1", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI2, "ADI2", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI3, "ADI3", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI4, "ADI4", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI5, "ADI5", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI6, "ADI6", [2]instfunc{(*CPU).adi, (*CPU).adi}},
	{symADI7, "ADI7", [2]instfunc{(*CPU).adi, (*CPU).adi}},

	{symADM, "ADM", [2]instfunc{(*CPU).adm, (*CPU).adm}},

	{symADR, "ADR", [2]instfunc{(*CPU).adr, (*CPU).adr}},

	{symAND, "AND", [2]instfunc{(*CPU).and, (*CPU).and}},
	{symANI, "ANI", [2]instfunc{(*CPU).ani, (*CPU).ani}},
	{symCALL, "CALL", [2]instfunc{(*CPU).call, (*CPU).call}},
	{symCMP, "CMP", [2]instfunc{(*CPU).cmp, (*CPU).cmp}},
	{symCPSR, "CPSR", [2]instfunc{(*CPU).cpsr, (*CPU).cpsr}},
	{symDEC, "DEC", [2]instfunc{(*CPU).dec, (*CPU).dec}},
	{symEX, "EX", [2]instfunc{(*CPU).ex, (*CPU).ex}},
	{symHALT, "HALT", [2]instfunc{(*CPU).halt, (*CPU).halt}},
	{symINC, "INC", [2]instfunc{(*CPU).inc, (*CPU).inc}},
	{symLBR, "LBR", [2]instfunc{(*CPU).lbr, (*CPU).lbr}},
	{symLBRC, "LBRC", [2]instfunc{(*CPU).lbrc, (*CPU).lbrc}},
	{symLBRZ, "LBRZ", [2]instfunc{(*CPU).lbrz, (*CPU).lbrz}},
	{symLBRQ, "LBRQ", [2]instfunc{(*CPU).lbrq, (*CPU).lbrq}},
	{symLDI0, "LDI0", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI1, "LDI1", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI2, "LDI2", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI3, "LDI3", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI4, "LDI4", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI5, "LDI5", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI6, "LDI6", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDI7, "LDI7", [2]instfunc{(*CPU).ldi, (*CPU).ldi}},
	{symLDM, "LDM", [2]instfunc{(*CPU).ldm, (*CPU).ldm}},
	{symNOP, "NOP", [2]instfunc{(*CPU).nop, (*CPU).nop}},
	{symOR, "OR", [2]instfunc{(*CPU).or, (*CPU).or}},
	{symORI, "ORI", [2]instfunc{(*CPU).ori, (*CPU).ori}},
	{symPOP0, "POP0", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP1, "POP1", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP2, "POP2", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP3, "POP3", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP4, "POP4", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP5, "POP5", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP6, "POP6", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPOP7, "POP7", [2]instfunc{(*CPU).popr, (*CPU).popr}},
	{symPUSH0, "PUSH0", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH1, "PUSH1", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH2, "PUSH2", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH3, "PUSH3", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH4, "PUSH4", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH5, "PUSH5", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH6, "PUSH6", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symPUSH7, "PUSH7", [2]instfunc{(*CPU).pushr, (*CPU).pushr}},
	{symRESETQ0, "RESETQ0", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ1, "RESETQ1", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ2, "RESETQ2", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ3, "RESETQ3", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ4, "RESETQ4", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ5, "RESETQ5", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ6, "RESETQ6", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRESETQ7, "RESETQ7", [2]instfunc{(*CPU).resetq, (*CPU).resetq}},
	{symRET, "RET", [2]instfunc{(*CPU).ret, (*CPU).ret}},
	{symSETQ0, "SETQ0", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ1, "SETQ1", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ2, "SETQ2", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ3, "SETQ3", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ4, "SETQ4", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ5, "SETQ5", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ6, "SETQ6", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSETQ7, "SETQ7", [2]instfunc{(*CPU).setq, (*CPU).setq}},
	{symSHL, "SHL", [2]instfunc{(*CPU).shl, (*CPU).shl}},
	{symSHLC, "SHLC", [2]instfunc{(*CPU).shlc, (*CPU).shlc}},
	{symSHR, "SHR", [2]instfunc{(*CPU).shr, (*CPU).shr}},
	{symSHRC, "SHRC", [2]instfunc{(*CPU).shrc, (*CPU).shrc}},
	{symSPSR, "SPSR", [2]instfunc{(*CPU).spsr, (*CPU).spsr}},
	{symSTI0, "STI0", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI1, "STI1", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI2, "STI2", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI3, "STI3", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI4, "STI4", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI5, "STI5", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI6, "STI6", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSTI7, "STI7", [2]instfunc{(*CPU).sti, (*CPU).sti}},
	{symSUB, "SUB", [2]instfunc{(*CPU).sub, (*CPU).sub}},

	{symSUBI, "SUBI", [2]instfunc{(*CPU).subi, (*CPU).subi}},

	{symSUBM, "SUBM", [2]instfunc{(*CPU).subm, (*CPU).subm}},

	{symXOR, "XOR", [2]instfunc{(*CPU).xor, (*CPU).xor}},
	{symXRI, "XRI", [2]instfunc{(*CPU).xri, (*CPU).xri}},
}

// Mode describes a memory addressing mode.
type Mode byte

// All possible memory addressing modes
const (
	IMM Mode = iota // Immediate
	IMP             // Implied (no operand)
	REL             // Relative
	ZPG             // Zero Page
	ZPX             // Zero Page,X
	ZPY             // Zero Page,Y
	ABS             // Absolute, using 2-byte operand as address
	ABX             // Absolute,X
	ABY             // Absolute,Y
	IND             // (Indirect)
	IDX             // (Indirect,X)
	IDY             // (Indirect),Y
	ACC             // Accumulator (no operand)
)

// Opcode data for an (opcode, mode) pair
type opcodeData struct {
	sym      opsym // internal opcode symbol
	mode     Mode  // addressing mode
	opcode   byte  // opcode hex value
	length   byte  // length of opcode + operand in bytes
	cycles   byte  // number of CPU cycles to execute command
	bpcycles byte  // additional CPU cycles if command crosses page boundary
	cmos     bool  // whether the opcode/mode pair is valid only on 65C02
}

// All valid (opcode, mode) pairs
var data = []opcodeData{
	{symADI0, IMM, 0x88, 2, 3, 0, false},
	{symADI1, IMM, 0x89, 2, 3, 0, false},
	{symADI2, IMM, 0x8a, 2, 3, 0, false},
	{symADI3, IMM, 0x8b, 2, 3, 0, false},
	{symADI4, IMM, 0x8c, 2, 3, 0, false},
	{symADI5, IMM, 0x8d, 2, 3, 0, false},
	{symADI6, IMM, 0x8e, 2, 3, 0, false},
	{symADI7, IMM, 0x8f, 2, 3, 0, false},

	{symADM, ABS, 0x90, 3, 4, 0, false},
	{symADM, ABS, 0x91, 3, 4, 0, false},
	{symADM, ABS, 0x92, 3, 4, 0, false},
	{symADM, ABS, 0x93, 3, 4, 0, false},
	{symADM, ABS, 0x94, 3, 4, 0, false},
	{symADM, ABS, 0x95, 3, 4, 0, false},
	{symADM, ABS, 0x96, 3, 4, 0, false},
	{symADM, ABS, 0x97, 3, 4, 0, false},

	{symADR, IMM, 0x80, 2, 3, 0, false},

	{symAND, IMM, 0x86, 2, 3, 0, false},

	{symANI, IMM, 0x50, 2, 3, 0, false},
	{symANI, IMM, 0x51, 2, 3, 0, false},
	{symANI, IMM, 0x52, 2, 3, 0, false},
	{symANI, IMM, 0x53, 2, 3, 0, false},
	{symANI, IMM, 0x54, 2, 3, 0, false},
	{symANI, IMM, 0x55, 2, 3, 0, false},
	{symANI, IMM, 0x56, 2, 3, 0, false},
	{symANI, IMM, 0x57, 2, 3, 0, false},

	{symCALL, ABS, 0x02, 3, 6, 0, false},

	{symCMP, IMM, 0x85, 2, 3, 0, false},

	{symDEC, IMP, 0x30, 1, 1, 0, false},
	{symDEC, IMP, 0x31, 1, 1, 0, false},
	{symDEC, IMP, 0x32, 1, 1, 0, false},
	{symDEC, IMP, 0x33, 1, 1, 0, false},
	{symDEC, IMP, 0x34, 1, 1, 0, false},
	{symDEC, IMP, 0x35, 1, 1, 0, false},
	{symDEC, IMP, 0x36, 1, 1, 0, false},
	{symDEC, IMP, 0x37, 1, 1, 0, false},

	{symEX, IMM, 0x84, 2, 3, 0, false},

	{symHALT, IMP, 0x01, 1, 1, 0, false},

	{symINC, IMP, 0x28, 1, 1, 0, false},
	{symINC, IMP, 0x29, 1, 1, 0, false},
	{symINC, IMP, 0x2a, 1, 1, 0, false},
	{symINC, IMP, 0x2b, 1, 1, 0, false},
	{symINC, IMP, 0x2c, 1, 1, 0, false},
	{symINC, IMP, 0x2d, 1, 1, 0, false},
	{symINC, IMP, 0x2e, 1, 1, 0, false},
	{symINC, IMP, 0x2f, 1, 1, 0, false},

	{symLBR, ABS, 0x18, 3, 4, 0, false},

	{symLBRC, ABS, 0x1a, 3, 4, 0, false},

	{symLBRQ, ABS, 0xb0, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb1, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb2, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb3, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb4, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb5, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb6, 3, 4, 0, false},
	{symLBRQ, ABS, 0xb7, 3, 4, 0, false},

	{symLBRZ, ABS, 0x1b, 3, 4, 0, false},

	{symLDI0, IMM, 0xe0, 2, 2, 0, false},
	{symLDI1, IMM, 0xe1, 2, 2, 0, false},
	{symLDI2, IMM, 0xe2, 2, 2, 0, false},
	{symLDI3, IMM, 0xe3, 2, 2, 0, false},
	{symLDI4, IMM, 0xe4, 2, 2, 0, false},
	{symLDI5, IMM, 0xe5, 2, 2, 0, false},
	{symLDI6, IMM, 0xe6, 2, 2, 0, false},
	{symLDI7, IMM, 0xe7, 2, 2, 0, false},

	{symLDM, ABS, 0xf0, 3, 4, 0, false},
	{symLDM, ABS, 0xf1, 3, 4, 0, false},
	{symLDM, ABS, 0xf2, 3, 4, 0, false},
	{symLDM, ABS, 0xf3, 3, 4, 0, false},
	{symLDM, ABS, 0xf4, 3, 4, 0, false},
	{symLDM, ABS, 0xf5, 3, 4, 0, false},
	{symLDM, ABS, 0xf6, 3, 4, 0, false},
	{symLDM, ABS, 0xf7, 3, 4, 0, false},

	{symNOP, IMP, 0x00, 1, 1, 0, false},

	{symOR, IMM, 0x87, 2, 2, 0, false},

	{symORI, IMM, 0x58, 2, 2, 0, false},
	{symORI, IMM, 0x59, 2, 2, 0, false},
	{symORI, IMM, 0x5a, 2, 2, 0, false},
	{symORI, IMM, 0x5b, 2, 2, 0, false},
	{symORI, IMM, 0x5c, 2, 2, 0, false},
	{symORI, IMM, 0x5d, 2, 2, 0, false},
	{symORI, IMM, 0x5e, 2, 2, 0, false},
	{symORI, IMM, 0x5f, 2, 2, 0, false},

	{symPOP0, IMP, 0x48, 1, 2, 0, false},
	{symPOP1, IMP, 0x49, 1, 2, 0, false},
	{symPOP2, IMP, 0x4a, 1, 2, 0, false},
	{symPOP3, IMP, 0x4b, 1, 2, 0, false},
	{symPOP4, IMP, 0x4c, 1, 2, 0, false},
	{symPOP5, IMP, 0x4d, 1, 2, 0, false},
	{symPOP6, IMP, 0x4e, 1, 2, 0, false},
	{symPOP7, IMP, 0x4f, 1, 2, 0, false},

	{symPUSH0, IMP, 0x40, 1, 2, 0, false},
	{symPUSH1, IMP, 0x41, 1, 2, 0, false},
	{symPUSH2, IMP, 0x42, 1, 2, 0, false},
	{symPUSH3, IMP, 0x43, 1, 2, 0, false},
	{symPUSH4, IMP, 0x44, 1, 2, 0, false},
	{symPUSH5, IMP, 0x45, 1, 2, 0, false},
	{symPUSH6, IMP, 0x46, 1, 2, 0, false},
	{symPUSH7, IMP, 0x47, 1, 2, 0, false},

	{symRESETQ0, IMP, 0x10, 1, 1, 0, false},
	{symRESETQ1, IMP, 0x11, 1, 1, 0, false},
	{symRESETQ2, IMP, 0x12, 1, 1, 0, false},
	{symRESETQ3, IMP, 0x13, 1, 1, 0, false},
	{symRESETQ4, IMP, 0x14, 1, 1, 0, false},
	{symRESETQ5, IMP, 0x15, 1, 1, 0, false},
	{symRESETQ6, IMP, 0x16, 1, 1, 0, false},
	{symRESETQ7, IMP, 0x17, 1, 1, 0, false},

	{symRET, IMP, 0x03, 1, 1, 6, false},

	{symSETQ0, IMP, 0x38, 1, 1, 0, false},
	{symSETQ1, IMP, 0x39, 1, 1, 0, false},
	{symSETQ2, IMP, 0x3a, 1, 1, 0, false},
	{symSETQ3, IMP, 0x3b, 1, 1, 0, false},
	{symSETQ4, IMP, 0x3c, 1, 1, 0, false},
	{symSETQ5, IMP, 0x3d, 1, 1, 0, false},
	{symSETQ6, IMP, 0x3e, 1, 1, 0, false},
	{symSETQ7, IMP, 0x3f, 1, 1, 0, false},

	{symSHL, IMP, 0x78, 1, 1, 0, false},
	{symSHL, IMP, 0x79, 1, 1, 0, false},
	{symSHL, IMP, 0x7a, 1, 1, 0, false},
	{symSHL, IMP, 0x7b, 1, 1, 0, false},
	{symSHL, IMP, 0x7c, 1, 1, 0, false},
	{symSHL, IMP, 0x7d, 1, 1, 0, false},
	{symSHL, IMP, 0x7e, 1, 1, 0, false},
	{symSHL, IMP, 0x7f, 1, 1, 0, false},

	{symSHLC, IMP, 0x20, 1, 1, 0, false},
	{symSHLC, IMP, 0x21, 1, 1, 0, false},
	{symSHLC, IMP, 0x22, 1, 1, 0, false},
	{symSHLC, IMP, 0x23, 1, 1, 0, false},
	{symSHLC, IMP, 0x24, 1, 1, 0, false},
	{symSHLC, IMP, 0x25, 1, 1, 0, false},
	{symSHLC, IMP, 0x26, 1, 1, 0, false},
	{symSHLC, IMP, 0x27, 1, 1, 0, false},

	{symSHR, IMP, 0x68, 1, 1, 0, false},
	{symSHR, IMP, 0x69, 1, 1, 0, false},
	{symSHR, IMP, 0x6a, 1, 1, 0, false},
	{symSHR, IMP, 0x6b, 1, 1, 0, false},
	{symSHR, IMP, 0x6c, 1, 1, 0, false},
	{symSHR, IMP, 0x6d, 1, 1, 0, false},
	{symSHR, IMP, 0x6e, 1, 1, 0, false},
	{symSHR, IMP, 0x6f, 1, 1, 0, false},

	{symSHRC, IMP, 0x70, 1, 1, 0, false},
	{symSHRC, IMP, 0x71, 1, 1, 0, false},
	{symSHRC, IMP, 0x72, 1, 1, 0, false},
	{symSHRC, IMP, 0x73, 1, 1, 0, false},
	{symSHRC, IMP, 0x74, 1, 1, 0, false},
	{symSHRC, IMP, 0x75, 1, 1, 0, false},
	{symSHRC, IMP, 0x76, 1, 1, 0, false},
	{symSHRC, IMP, 0x77, 1, 1, 0, false},

	{symCPSR, IMM, 0x05, 2, 2, 0, false},
	{symSPSR, IMM, 0x04, 2, 2, 0, false},

	{symSTI0, ABS, 0xe8, 3, 4, 0, false},
	{symSTI1, ABS, 0xe9, 3, 4, 0, false},
	{symSTI2, ABS, 0xea, 3, 4, 0, false},
	{symSTI3, ABS, 0xeb, 3, 4, 0, false},
	{symSTI4, ABS, 0xec, 3, 4, 0, false},
	{symSTI5, ABS, 0xed, 3, 4, 0, false},
	{symSTI6, ABS, 0xee, 3, 4, 0, false},
	{symSTI7, ABS, 0xef, 3, 4, 0, false},

	{symSUB, IMM, 0x82, 2, 2, 0, false},

	{symSUBI, IMM, 0xb8, 2, 2, 0, false},
	{symSUBI, IMM, 0xb9, 2, 2, 0, false},
	{symSUBI, IMM, 0xba, 2, 2, 0, false},
	{symSUBI, IMM, 0xbb, 2, 2, 0, false},
	{symSUBI, IMM, 0xbc, 2, 2, 0, false},
	{symSUBI, IMM, 0xbd, 2, 2, 0, false},
	{symSUBI, IMM, 0xbe, 2, 2, 0, false},
	{symSUBI, IMM, 0xbf, 2, 2, 0, false},

	{symSUBM, ABS, 0xc0, 3, 4, 0, false},
	{symSUBM, ABS, 0xc1, 3, 4, 0, false},
	{symSUBM, ABS, 0xc2, 3, 4, 0, false},
	{symSUBM, ABS, 0xc3, 3, 4, 0, false},
	{symSUBM, ABS, 0xc4, 3, 4, 0, false},
	{symSUBM, ABS, 0xc5, 3, 4, 0, false},
	{symSUBM, ABS, 0xc6, 3, 4, 0, false},
	{symSUBM, ABS, 0xc7, 3, 4, 0, false},

	{symXOR, IMM, 0x19, 2, 2, 0, false},

	{symXRI, IMM, 0x60, 2, 2, 0, false},
	{symXRI, IMM, 0x61, 2, 2, 0, false},
	{symXRI, IMM, 0x62, 2, 2, 0, false},
	{symXRI, IMM, 0x63, 2, 2, 0, false},
	{symXRI, IMM, 0x64, 2, 2, 0, false},
	{symXRI, IMM, 0x65, 2, 2, 0, false},
	{symXRI, IMM, 0x66, 2, 2, 0, false},
	{symXRI, IMM, 0x67, 2, 2, 0, false},
}

// Unused opcodes
type unused struct {
	opcode byte
	mode   Mode
	length byte
	cycles byte
}

var unusedData = []unused{
	{0x06, IMP, 1, 1},
	{0x07, IMP, 1, 1},
	{0x08, IMP, 1, 1},
	{0x09, IMP, 1, 1},
	{0x0a, IMP, 1, 1},
	{0x0b, IMP, 1, 1},
	{0x0c, IMP, 1, 1},
	{0x0d, IMP, 1, 1},
	{0x0e, IMP, 1, 1},
	{0x0f, IMP, 1, 1},
	{0x1c, IMP, 1, 1},
	{0x1d, IMP, 1, 1},
	{0x1e, IMP, 1, 1},
	{0x1f, IMP, 1, 1},
	{0x81, IMP, 1, 1},
	{0x83, IMP, 1, 1},
	{0x98, IMP, 1, 1},
	{0x99, IMP, 1, 1},
	{0x9a, IMP, 1, 1},
	{0x9b, IMP, 1, 1},
	{0x9c, IMP, 1, 1},
	{0x9d, IMP, 1, 1},
	{0x9e, IMP, 1, 1},
	{0x9f, IMP, 1, 1},
	{0xa0, IMP, 1, 1},
	{0xa1, IMP, 1, 1},
	{0xa2, IMP, 1, 1},
	{0xa3, IMP, 1, 1},
	{0xa4, IMP, 1, 1},
	{0xa5, IMP, 1, 1},
	{0xa6, IMP, 1, 1},
	{0xa7, IMP, 1, 1},
	{0xa8, IMP, 1, 1},
	{0xa9, IMP, 1, 1},
	{0xaa, IMP, 1, 1},
	{0xab, IMP, 1, 1},
	{0xac, IMP, 1, 1},
	{0xad, IMP, 1, 1},
	{0xae, IMP, 1, 1},
	{0xaf, IMP, 1, 1},
	{0xc8, IMP, 1, 1},
	{0xc9, IMP, 1, 1},
	{0xca, IMP, 1, 1},
	{0xcb, IMP, 1, 1},
	{0xcc, IMP, 1, 1},
	{0xcd, IMP, 1, 1},
	{0xce, IMP, 1, 1},
	{0xcf, IMP, 1, 1},
	{0xd0, IMP, 1, 1},
	{0xd1, IMP, 1, 1},
	{0xd2, IMP, 1, 1},
	{0xd3, IMP, 1, 1},
	{0xd4, IMP, 1, 1},
	{0xd5, IMP, 1, 1},
	{0xd6, IMP, 1, 1},
	{0xd7, IMP, 1, 1},
	{0xd8, IMP, 1, 1},
	{0xd9, IMP, 1, 1},
	{0xda, IMP, 1, 1},
	{0xdb, IMP, 1, 1},
	{0xdc, IMP, 1, 1},
	{0xdd, IMP, 1, 1},
	{0xde, IMP, 1, 1},
	{0xdf, IMP, 1, 1},
	{0xf8, IMP, 1, 1},
	{0xf9, IMP, 1, 1},
	{0xfa, IMP, 1, 1},
	{0xfb, IMP, 1, 1},
	{0xfc, IMP, 1, 1},
	{0xfd, IMP, 1, 1},
	{0xfe, IMP, 1, 1},
	{0xff, IMP, 1, 1},
}

// An Instruction describes a CPU instruction, including its name,
// its addressing mode, its opcode value, its operand size, and its CPU cycle
// cost.
type Instruction struct {
	Name     string   // all-caps name of the instruction
	Mode     Mode     // addressing mode
	Opcode   byte     // hexadecimal opcode value
	Length   byte     // combined size of opcode and operand, in bytes
	Cycles   byte     // number of CPU cycles to execute the instruction
	BPCycles byte     // additional cycles required if boundary page crossed
	fn       instfunc // emulator implementation of the function
}

// An InstructionSet defines the set of all possible instructions that
// can run on the emulated CPU.
type InstructionSet struct {
	Arch         Architecture
	instructions [256]Instruction          // all instructions by opcode
	variants     map[string][]*Instruction // variants of each instruction
}

// Lookup retrieves a CPU instruction corresponding to the requested opcode.
func (s *InstructionSet) Lookup(opcode byte) *Instruction {
	return &s.instructions[opcode]
}

// GetInstructions returns all CPU instructions whose name matches the
// provided string.
func (s *InstructionSet) GetInstructions(name string) []*Instruction {
	return s.variants[strings.ToUpper(name)]
}

// Create an instruction set for a CPU architecture.
func newInstructionSet(arch Architecture) *InstructionSet {
	set := &InstructionSet{Arch: arch}

	// Create a map from symbol to implementation for fast lookups.
	//log.Println("***** newInstructionSet ...")
	symToImpl := make(map[opsym]*opcodeImpl, len(impl))
	for i := range impl {
		symToImpl[impl[i].sym] = &impl[i]
		//log.Println("name: ", impl[i].name)
	}

	// Create a map from instruction name to the slice of all instruction
	// variants matching that name.
	set.variants = make(map[string][]*Instruction)

	unusedName := "???"

	// For each instruction, create a list of opcode variants valid for
	// the architecture.
	for _, d := range data {
		inst := &set.instructions[d.opcode]

		// If opcode has only a CMOS implementation and this is NMOS, create
		// an unused instruction for it.
		if d.cmos && arch != CMOS {
			inst.Name = unusedName
			inst.Mode = d.mode
			inst.Opcode = d.opcode
			inst.Length = d.length
			inst.Cycles = d.cycles
			inst.BPCycles = 0
			inst.fn = (*CPU).unusedn
			continue
		}

		impl := symToImpl[d.sym]
		if impl.fn[arch] == nil {
			continue // some opcodes have no architecture implementation
		}

		inst.Name = impl.name
		inst.Mode = d.mode
		inst.Opcode = d.opcode
		inst.Length = d.length
		inst.Cycles = d.cycles
		inst.BPCycles = d.bpcycles
		inst.fn = impl.fn[arch]

		set.variants[inst.Name] = append(set.variants[inst.Name], inst)
	}

	// Add unused opcodes to the instruction set. This information is useful
	// mostly for 65c02, where unused operations do something predicable
	// (i.e., eat cycles and nothing else).
	for _, u := range unusedData {
		inst := &set.instructions[u.opcode]
		inst.Name = unusedName
		inst.Mode = u.mode
		inst.Opcode = u.opcode
		inst.Length = u.length
		inst.Cycles = u.cycles
		inst.BPCycles = 0
		switch arch {
		case NMOS:
			inst.fn = (*CPU).unusedn
		case CMOS:
			inst.fn = (*CPU).unusedc
		}
	}

	for i := 0; i < 256; i++ {
		//fmt.Printf("set.instructions[i] = %s; opcode = x%02x\n", set.instructions[i].Name, set.instructions[i].Opcode)
		if set.instructions[i].Name == "" {
			panic("missing instruction")
		}
	}
	return set
}

var instructionSets [2]*InstructionSet

// GetInstructionSet returns an instruction set for the requested CPU
// architecture.
func GetInstructionSet(arch Architecture) *InstructionSet {
	//log.Println("***** Entered GetInstructionSet, arch= ", arch)
	if instructionSets[arch] == nil {
		// Lazy-create the instruction set.
		//log.Println("***** Create the instruction set")
		instructionSets[arch] = newInstructionSet(arch)
	}
	return instructionSets[arch]
}
