// Copyright 2014-2018 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpu implements a 6502 CPU instruction
// set and emulator.
package cpu

import (
	"fmt"
	"log"
	"os"
)

// Architecture selects the CPU chip: 6502 or 65c02
type Architecture byte

const (
	// NMOS 6502 CPU
	NMOS Architecture = iota

	// CMOS 65c02 CPU
	CMOS
)

// BrkHandler is an interface implemented by types that wish to be notified
// when a BRK instruction is about to be executed.
type BrkHandler interface {
	OnBrk(cpu *CPU)
}

// CPU represents a single 6502 CPU. It contains a pointer to the
// memory associated with the CPU.
type CPU struct {
	Arch        Architecture    // CPU architecture
	Reg         Registers       // CPU registers
	Mem         Memory          // assigned memory
	Cycles      uint64          // total executed CPU cycles
	LastPC      uint16          // Previous program counter
	InstSet     *InstructionSet // Instruction set used by the CPU
	pageCrossed bool
	deltaCycles int8
	debugger    *Debugger
	brkHandler  BrkHandler
	storeByte   func(cpu *CPU, addr uint16, v byte)
}

// Interrupt vectors
const (
	vectorNMI   = 0xfffa
	vectorReset = 0xfffc
	vectorIRQ   = 0xfffe
	vectorBRK   = 0xfffe
)

// NewCPU creates an emulated 6502 CPU bound to the specified memory.
func NewCPU(arch Architecture, m Memory) *CPU {
	LogFile, err := os.OpenFile("CPU1.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	infoLogger := log.New(LogFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger.Println("***** Entered cpu.NewCPU()")

	cpu := &CPU{
		Arch:      arch,
		Mem:       m,
		InstSet:   GetInstructionSet(arch),
		storeByte: (*CPU).storeByteNormal,
	}

	cpu.Reg.Init()
	return cpu
}

// SetPC updates the CPU program counter to 'addr'.
func (cpu *CPU) SetPC(addr uint16) {
	cpu.Reg.PC = addr
}

// GetInstruction returns the instruction opcode at the requested address.
func (cpu *CPU) GetInstruction(addr uint16) *Instruction {
	opcode := cpu.Mem.LoadByte(addr)
	return cpu.InstSet.Lookup(opcode)
}

// NextAddr returns the address of the next instruction following the
// instruction at addr.
func (cpu *CPU) NextAddr(addr uint16) uint16 {
	opcode := cpu.Mem.LoadByte(addr)
	inst := cpu.InstSet.Lookup(opcode)
	return addr + uint16(inst.Length)
}

// Step the cpu by one instruction.
func (cpu *CPU) Step() {
	// Grab the next opcode at the current PC
	//log.Printf("CPU Step. PC = x%04x\n", cpu.Reg.PC)
	opcode := cpu.Mem.LoadByte(cpu.Reg.PC)

	// Look up the instruction data for the opcode
	inst := cpu.InstSet.Lookup(opcode)

	// If the instruction is undefined, reset the CPU (for now).
	if inst.fn == nil {
		cpu.reset()
		return
	}

	// If a BRK instruction is about to be executed and a BRK handler has been
	// installed, call the BRK handler instead of executing the instruction.
	if inst.Opcode == 0x00 && cpu.brkHandler != nil {
		cpu.brkHandler.OnBrk(cpu)
		return
	}

	// Fetch the operand (if any) and advance the PC
	var buf [2]byte
	operand := buf[:inst.Length-1]
	cpu.Mem.LoadBytes(cpu.Reg.PC+1, operand)
	cpu.LastPC = cpu.Reg.PC
	cpu.Reg.PC += uint16(inst.Length)

	// Execute the instruction
	cpu.pageCrossed = false
	cpu.deltaCycles = 0
	inst.fn(cpu, inst, operand)

	// Update the CPU cycle counter, with special-case logic
	// to handle a page boundary crossing
	cpu.Cycles += uint64(int8(inst.Cycles) + cpu.deltaCycles)
	if cpu.pageCrossed {
		cpu.Cycles += uint64(inst.BPCycles)
	}

	// Update the debugger so it handle breakpoints.
	if cpu.debugger != nil {
		cpu.debugger.onUpdatePC(cpu, cpu.Reg.PC)
	}
}

// AttachBrkHandler attaches a handler that is called whenever the BRK
// instruction is executed.
func (cpu *CPU) AttachBrkHandler(handler BrkHandler) {
	cpu.brkHandler = handler
}

// AttachDebugger attaches a debugger to the CPU. The debugger receives
// notifications whenever the CPU executes an instruction or stores a byte
// to memory.
func (cpu *CPU) AttachDebugger(debugger *Debugger) {
	cpu.debugger = debugger
	cpu.storeByte = (*CPU).storeByteDebugger
}

// DetachDebugger detaches the currently debugger from the CPU.
func (cpu *CPU) DetachDebugger() {
	cpu.debugger = nil
	cpu.storeByte = (*CPU).storeByteNormal
}

// Load a byte value from using the requested addressing mode
// and the operand to determine where to load it from.
func (cpu *CPU) load(mode Mode, operand []byte) byte {
	switch mode {
	case IMM:
		return operand[0]
	case ZPG:
		zpaddr := operandToAddress(operand)
		return cpu.Mem.LoadByte(zpaddr)
	// case ZPX:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.X)
	// 	return cpu.Mem.LoadByte(zpaddr)
	// case ZPY:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.Y)
	// 	return cpu.Mem.LoadByte(zpaddr)
	case ABS:
		addr := operandToAddress(operand)
		return cpu.Mem.LoadByte(addr)
	// case ABX:
	// 	addr := operandToAddress(operand)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.X)
	// 	return cpu.Mem.LoadByte(addr)
	// case ABY:
	// 	addr := operandToAddress(operand)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.Y)
	// 	return cpu.Mem.LoadByte(addr)
	// case IDX:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.X)
	// 	addr := cpu.Mem.LoadAddress(zpaddr)
	// 	return cpu.Mem.LoadByte(addr)
	// case IDY:
	// 	zpaddr := operandToAddress(operand)
	// 	addr := cpu.Mem.LoadAddress(zpaddr)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.Y)
	// 	return cpu.Mem.LoadByte(addr)
	// case ACC:
	// 	return cpu.Reg.A
	default:
		panic("Invalid addressing mode")
	}
}

// Load a 16-bit address value from memory using the requested addressing mode
// and the 16-bit instruction operand.
func (cpu *CPU) loadAddress(mode Mode, operand []byte) uint16 {
	switch mode {
	case ABS:
		return operandToAddress(operand)
	case IND:
		addr := operandToAddress(operand)
		return cpu.Mem.LoadAddress(addr)
	default:
		panic("Invalid addressing mode")
	}
}

// Store a byte value using the specified addressing mode and the
// variable-sized instruction operand to determine where to store it.
func (cpu *CPU) store(mode Mode, operand []byte, v byte) {
	switch mode {
	case ZPG:
		zpaddr := operandToAddress(operand)
		cpu.storeByte(cpu, zpaddr, v)
	// case ZPX:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.X)
	// 	cpu.storeByte(cpu, zpaddr, v)
	// case ZPY:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.Y)
	// 	cpu.storeByte(cpu, zpaddr, v)
	case ABS:
		addr := operandToAddress(operand)
		cpu.storeByte(cpu, addr, v)
	// case ABX:
	// 	addr := operandToAddress(operand)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.X)
	// 	cpu.storeByte(cpu, addr, v)
	// case ABY:
	// 	addr := operandToAddress(operand)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.Y)
	// 	cpu.storeByte(cpu, addr, v)
	// case IDX:
	// 	zpaddr := operandToAddress(operand)
	// 	zpaddr = offsetZeroPage(zpaddr, cpu.Reg.X)
	// 	addr := cpu.Mem.LoadAddress(zpaddr)
	// 	cpu.storeByte(cpu, addr, v)
	// case IDY:
	// 	zpaddr := operandToAddress(operand)
	// 	addr := cpu.Mem.LoadAddress(zpaddr)
	// 	addr, cpu.pageCrossed = offsetAddress(addr, cpu.Reg.Y)
	// 	cpu.storeByte(cpu, addr, v)
	// case ACC:
	// 	cpu.Reg.A = v
	default:
		panic("Invalid addressing mode")
	}
}

// Execute a branch using the instruction operand.
// func (cpu *CPU) branch(operand []byte) {
// 	offset := operandToAddress(operand)
// 	oldPC := cpu.Reg.PC
// 	if offset < 0x80 {
// 		cpu.Reg.PC += uint16(offset)
// 	} else {
// 		cpu.Reg.PC -= uint16(0x100 - offset)
// 	}
// 	cpu.deltaCycles++
// 	if ((cpu.Reg.PC ^ oldPC) & 0xff00) != 0 {
// 		cpu.deltaCycles++
// 	}
// }

// Store the byte value 'v' add the address 'addr'.
func (cpu *CPU) storeByteNormal(addr uint16, v byte) {
	cpu.Mem.StoreByte(addr, v)
}

// Store the byte value 'v' add the address 'addr'.
func (cpu *CPU) storeByteDebugger(addr uint16, v byte) {
	cpu.debugger.onDataStore(cpu, addr, v)
	cpu.Mem.StoreByte(addr, v)
}

// Push the address 'addr' onto the stack.
func (cpu *CPU) pushAddress(addr uint16) {
	cpu.push(byte(addr >> 8))
	cpu.push(byte(addr))
}

// Pop a 16-bit address off the stack.
func (cpu *CPU) popAddress() uint16 {
	lo := cpu.pop()
	hi := cpu.pop()
	return uint16(lo) | (uint16(hi) << 8)
}

// Pop a value from the stack and return it.
func (cpu *CPU) pop() byte {
	cpu.Reg.SP++
	return cpu.Mem.LoadByte(stackAddress(cpu.Reg.SP))
}

// Push a value 'v' onto the stack.
func (cpu *CPU) push(v byte) {
	cpu.storeByte(cpu, stackAddress(cpu.Reg.SP), v)
	cpu.Reg.SP--
}

// Set bit in byte
func bitSet(b byte, nbit byte) byte {
	b = b | (1 << (nbit))
	return b
}

// Clear bit in byte
func bitClear(b byte, nbit byte) byte {
	b = b & ^(1 << (nbit))
	return b
}

// Update the Zero and Negative flags based on the value of 'v'.
func (cpu *CPU) updateNZ(v byte) {
	cpu.Reg.Zero = (v == 0)
	cpu.Reg.Sign = ((v & 0x80) != 0)
}

// Decode Registers X and Y from an opcode (byte)
func (cpu *CPU) getRegXY(v byte) (byte, byte) {
	x := (v & 0b01110000) >> 4
	y := (v & 0b00000111)
	return x, y
}

// Decode Register # from  3 lsb of an opcode (byte)
func (cpu *CPU) getReg(v byte) byte {
	r := (v & 0b00000111)
	return r
}

// Handle a handleInterrupt by storing the program counter and status flags on
// the stack. Then switch the program counter to the requested address.
func (cpu *CPU) handleInterrupt(brk bool, addr uint16) {
	cpu.pushAddress(cpu.Reg.PC)
	cpu.push(cpu.Reg.SavePS(brk))

	cpu.Reg.InterruptDisable = true
	if cpu.Arch == CMOS {
		cpu.Reg.Decimal = false
	}

	cpu.Reg.PC = cpu.Mem.LoadAddress(addr)
}

// Generate a maskable IRQ (hardware) interrupt request.
func (cpu *CPU) irq() {
	if !cpu.Reg.InterruptDisable {
		cpu.handleInterrupt(false, vectorIRQ)
	}
}

// Generate a non-maskable interrupt.
func (cpu *CPU) nmi() {
	cpu.handleInterrupt(false, vectorNMI)
}

// Generate a reset signal.
func (cpu *CPU) reset() {
	cpu.Reg.PC = cpu.Mem.LoadAddress(vectorReset)
}

// 2's Complement Add with Carry
func (cpu *CPU) twosCompAdd(a byte, b byte) byte {
	// x := uint32(a)
	// y := uint32(b)
	// carry := boolToUint32(cpu.Reg.Carry)
	// v := x + y + carry
	// cpu.Reg.Carry = (v >= 0x100)
	// cpu.Reg.Overflow = (((x & 0x80) == (y & 0x80)) && ((x & 0x80) != (v & 0x80)))
	//x := ^a
	//cpu.updateNZ(byte(v))
	//return byte(v)
	return 0
}

// Add with carry (CMOS)
/* func (cpu *CPU) adcc(inst *Instruction, operand []byte) {
	acc := uint32(cpu.Reg.A)
	add := uint32(cpu.load(inst.Mode, operand))
	carry := boolToUint32(cpu.Reg.Carry)
	var v uint32

	cpu.Reg.Overflow = (((acc ^ add) & 0x80) == 0)

	switch cpu.Reg.Decimal {
	case true:
		cpu.deltaCycles++

		lo := (acc & 0x0f) + (add & 0x0f) + carry

		var carrylo uint32
		if lo >= 0x0a {
			carrylo = 0x10
			lo -= 0xa
		}

		hi := (acc & 0xf0) + (add & 0xf0) + carrylo

		if hi >= 0xa0 {
			cpu.Reg.Carry = true
			if hi >= 0x180 {
				cpu.Reg.Overflow = false
			}
			hi -= 0xa0
		} else {
			cpu.Reg.Carry = false
			if hi < 0x80 {
				cpu.Reg.Overflow = false
			}
		}

		v = hi | lo

	case false:
		v = acc + add + carry
		if v >= 0x100 {
			cpu.Reg.Carry = true
			if v >= 0x180 {
				cpu.Reg.Overflow = false
			}
		} else {
			cpu.Reg.Carry = false
			if v < 0x80 {
				cpu.Reg.Overflow = false
			}
		}
	}

	cpu.Reg.A = byte(v)
	cpu.updateNZ(cpu.Reg.A)
} */

// Add with carry (NMOS)
/* func (cpu *CPU) adcn(inst *Instruction, operand []byte) {
	acc := uint32(cpu.Reg.A)
	add := uint32(cpu.load(inst.Mode, operand))
	carry := boolToUint32(cpu.Reg.Carry)
	var v uint32

	switch cpu.Reg.Decimal {
	case true:
		lo := (acc & 0x0f) + (add & 0x0f) + carry

		var carrylo uint32
		if lo >= 0x0a {
			carrylo = 0x10
			lo -= 0x0a
		}

		hi := (acc & 0xf0) + (add & 0xf0) + carrylo

		if hi >= 0xa0 {
			cpu.Reg.Carry = true
			hi -= 0xa0
		} else {
			cpu.Reg.Carry = false
		}

		v = hi | lo

		cpu.Reg.Overflow = ((acc^v)&0x80) != 0 && ((acc^add)&0x80) == 0

	case false:
		v = acc + add + carry
		cpu.Reg.Carry = (v >= 0x100)
		cpu.Reg.Overflow = (((acc & 0x80) == (add & 0x80)) && ((acc & 0x80) != (v & 0x80)))
	}

	cpu.Reg.A = byte(v)
	cpu.updateNZ(cpu.Reg.A)
} */

// Arithmetic Shift Left
// func (cpu *CPU) asl(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Carry = ((v & 0x80) == 0x80)
// 	v = v << 1
// 	cpu.updateNZ(v)
// 	cpu.store(inst.Mode, operand, v)
// 	if cpu.Arch == CMOS && inst.Mode == ABX && !cpu.pageCrossed {
// 		cpu.deltaCycles--
// 	}
// }

// Branch if Carry Clear
// func (cpu *CPU) bcc(inst *Instruction, operand []byte) {
// 	if !cpu.Reg.Carry {
// 		cpu.branch(operand)
// 	}
// }

// Branch if Carry Set
// func (cpu *CPU) bcs(inst *Instruction, operand []byte) {
// 	if cpu.Reg.Carry {
// 		cpu.branch(operand)
// 	}
// }

// Bit Test
// func (cpu *CPU) bit(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Zero = ((v & cpu.Reg.A) == 0)
// 	cpu.Reg.Sign = ((v & 0x80) != 0)
// 	cpu.Reg.Overflow = ((v & 0x40) != 0)
// }

// Branch if MInus (negative)
// func (cpu *CPU) bmi(inst *Instruction, operand []byte) {
// 	if cpu.Reg.Sign {
// 		cpu.branch(operand)
// 	}
// }

// Branch if Not Equal (not zero)
// func (cpu *CPU) bne(inst *Instruction, operand []byte) {
// 	if !cpu.Reg.Zero {
// 		cpu.branch(operand)
// 	}
// }

// Branch if PLus (positive)
// func (cpu *CPU) bpl(inst *Instruction, operand []byte) {
// 	if !cpu.Reg.Sign {
// 		cpu.branch(operand)
// 	}
// }

// Break
// func (cpu *CPU) brk(inst *Instruction, operand []byte) {
// 	cpu.Reg.PC++
// 	cpu.handleInterrupt(true, vectorBRK)
// }

// Branch if oVerflow Clear
// func (cpu *CPU) bvc(inst *Instruction, operand []byte) {
// 	if !cpu.Reg.Overflow {
// 		cpu.branch(operand)
// 	}
// }

// Branch if oVerflow Set
// func (cpu *CPU) bvs(inst *Instruction, operand []byte) {
// 	if cpu.Reg.Overflow {
// 		cpu.branch(operand)
// 	}
// }

// Clear Carry flag
// func (cpu *CPU) clc(inst *Instruction, operand []byte) {
// 	cpu.Reg.Carry = false
// }

// Clear Decimal flag
// func (cpu *CPU) cld(inst *Instruction, operand []byte) {
// 	cpu.Reg.Decimal = false
// }

// Clear InterruptDisable flag
// func (cpu *CPU) cli(inst *Instruction, operand []byte) {
// 	cpu.Reg.InterruptDisable = false
// }

// Clear oVerflow flag
// func (cpu *CPU) clv(inst *Instruction, operand []byte) {
// 	cpu.Reg.Overflow = false
// }

// Compare to X register
// func (cpu *CPU) cpx(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Carry = (cpu.Reg.X >= v)
// 	cpu.updateNZ(cpu.Reg.X - v)
// }

// Compare to Y register
// func (cpu *CPU) cpy(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Carry = (cpu.Reg.Y >= v)
// 	cpu.updateNZ(cpu.Reg.Y - v)
// }

// Decrement X register
// func (cpu *CPU) dex(inst *Instruction, operand []byte) {
// 	cpu.Reg.X--
// 	cpu.updateNZ(cpu.Reg.X)
// }

// Decrement Y register
// func (cpu *CPU) dey(inst *Instruction, operand []byte) {
// 	cpu.Reg.Y--
// 	cpu.updateNZ(cpu.Reg.Y)
// }

// Boolean XOR
// func (cpu *CPU) eor(inst *Instruction, operand []byte) {
// 	cpu.Reg.A ^= cpu.load(inst.Mode, operand)
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Increment X register
// func (cpu *CPU) inx(inst *Instruction, operand []byte) {
// 	cpu.Reg.X++
// 	cpu.updateNZ(cpu.Reg.X)
// }

// Increment Y register
// func (cpu *CPU) iny(inst *Instruction, operand []byte) {
// 	cpu.Reg.Y++
// 	cpu.updateNZ(cpu.Reg.Y)
// }

// Jump to memory address (NMOS 6502)
// func (cpu *CPU) jmpn(inst *Instruction, operand []byte) {
// 	cpu.Reg.PC = cpu.loadAddress(inst.Mode, operand)
// }

// Jump to memory address (CMOS 65c02)
// func (cpu *CPU) jmpc(inst *Instruction, operand []byte) {
// 	if inst.Mode == IND && operand[0] == 0xff {
// 		// Fix bug in NMOS 6502 address loading. In NMOS 6502, a JMP ($12FF)
// 		// would load LSB of jmp target from $12FF and MSB from $1200.
// 		// In CMOS, it loads the MSB from $1300.
// 		addr0 := uint16(operand[1])<<8 | 0xff
// 		addr1 := addr0 + 1
// 		lo := cpu.Mem.LoadByte(addr0)
// 		hi := cpu.Mem.LoadByte(addr1)
// 		cpu.Reg.PC = uint16(lo) | uint16(hi)<<8
// 		cpu.deltaCycles++
// 		return
// 	}

// 	cpu.Reg.PC = cpu.loadAddress(inst.Mode, operand)
// }

// Jump to subroutine
// func (cpu *CPU) jsr(inst *Instruction, operand []byte) {
// 	addr := cpu.loadAddress(inst.Mode, operand)
// 	cpu.pushAddress(cpu.Reg.PC - 1)
// 	cpu.Reg.PC = addr
// }

// load Accumulator
// func (cpu *CPU) lda(inst *Instruction, operand []byte) {
// 	cpu.Reg.A = cpu.load(inst.Mode, operand)
// 	cpu.updateNZ(cpu.Reg.A)
// }

// load the X register
// func (cpu *CPU) ldx(inst *Instruction, operand []byte) {
// 	cpu.Reg.X = cpu.load(inst.Mode, operand)
// 	cpu.updateNZ(cpu.Reg.X)
// }

// load the Y register
// func (cpu *CPU) ldy(inst *Instruction, operand []byte) {
// 	cpu.Reg.Y = cpu.load(inst.Mode, operand)
// 	cpu.updateNZ(cpu.Reg.Y)
// }

// Logical Shift Right
// func (cpu *CPU) lsr(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Carry = ((v & 1) == 1)
// 	v = v >> 1
// 	cpu.updateNZ(v)
// 	cpu.store(inst.Mode, operand, v)
// 	if cpu.Arch == CMOS && inst.Mode == ABX && !cpu.pageCrossed {
// 		cpu.deltaCycles--
// 	}
// }

// Boolean OR
// func (cpu *CPU) ora(inst *Instruction, operand []byte) {
// 	cpu.Reg.A |= cpu.load(inst.Mode, operand)
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Push Accumulator
// func (cpu *CPU) pha(inst *Instruction, operand []byte) {
// 	cpu.push(cpu.Reg.A)
// }

// Push Processor flags
// func (cpu *CPU) php(inst *Instruction, operand []byte) {
// 	cpu.push(cpu.Reg.SavePS(true))
// }

// Push X register (65c02 only)
// func (cpu *CPU) phx(inst *Instruction, operand []byte) {
// 	cpu.push(cpu.Reg.X)
// }

// Push Y register (65c02 only)
// func (cpu *CPU) phy(inst *Instruction, operand []byte) {
// 	cpu.push(cpu.Reg.Y)
// }

// Pull (pop) Accumulator
// func (cpu *CPU) pla(inst *Instruction, operand []byte) {
// 	cpu.Reg.A = cpu.pop()
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Pull (pop) Processor flags
// func (cpu *CPU) plp(inst *Instruction, operand []byte) {
// 	v := cpu.pop()
// 	cpu.Reg.RestorePS(v)
// }

// Pull (pop) X register (65c02 only)
// func (cpu *CPU) plx(inst *Instruction, operand []byte) {
// 	cpu.Reg.X = cpu.pop()
// 	cpu.updateNZ(cpu.Reg.X)
// }

// Pull (pop) Y register (65c02 only)
// func (cpu *CPU) ply(inst *Instruction, operand []byte) {
// 	cpu.Reg.Y = cpu.pop()
// 	cpu.updateNZ(cpu.Reg.Y)
// }

// Rotate Left
// func (cpu *CPU) rol(inst *Instruction, operand []byte) {
// 	tmp := cpu.load(inst.Mode, operand)
// 	v := (tmp << 1) | boolToByte(cpu.Reg.Carry)
// 	cpu.Reg.Carry = ((tmp & 0x80) != 0)
// 	cpu.updateNZ(v)
// 	cpu.store(inst.Mode, operand, v)
// 	if cpu.Arch == CMOS && inst.Mode == ABX && !cpu.pageCrossed {
// 		cpu.deltaCycles--
// 	}
// }

// Rotate Right
// func (cpu *CPU) ror(inst *Instruction, operand []byte) {
// 	tmp := cpu.load(inst.Mode, operand)
// 	v := (tmp >> 1) | (boolToByte(cpu.Reg.Carry) << 7)
// 	cpu.Reg.Carry = ((tmp & 1) != 0)
// 	cpu.updateNZ(v)
// 	cpu.store(inst.Mode, operand, v)
// 	if cpu.Arch == CMOS && inst.Mode == ABX && !cpu.pageCrossed {
// 		cpu.deltaCycles--
// 	}
// }

// Return from Interrupt
// func (cpu *CPU) rti(inst *Instruction, operand []byte) {
// 	v := cpu.pop()
// 	cpu.Reg.RestorePS(v)
// 	cpu.Reg.PC = cpu.popAddress()
// }

// Return from Subroutine
// func (cpu *CPU) rts(inst *Instruction, operand []byte) {
// 	addr := cpu.popAddress()
// 	cpu.Reg.PC = addr + 1
// }

// Subtract with Carry (CMOS)
// func (cpu *CPU) sbcc(inst *Instruction, operand []byte) {
// 	acc := uint32(cpu.Reg.A)
// 	sub := uint32(cpu.load(inst.Mode, operand))
// 	carry := boolToUint32(cpu.Reg.Carry)
// 	cpu.Reg.Overflow = ((acc ^ sub) & 0x80) != 0
// 	var v uint32

// 	switch cpu.Reg.Decimal {
// 	case true:
// 		cpu.deltaCycles++

// 		lo := 0x0f + (acc & 0x0f) - (sub & 0x0f) + carry

// 		var carrylo uint32
// 		if lo < 0x10 {
// 			lo -= 0x06
// 			carrylo = 0
// 		} else {
// 			lo -= 0x10
// 			carrylo = 0x10
// 		}

// 		hi := 0xf0 + (acc & 0xf0) - (sub & 0xf0) + carrylo

// 		if hi < 0x100 {
// 			cpu.Reg.Carry = false
// 			if hi < 0x80 {
// 				cpu.Reg.Overflow = false
// 			}
// 			hi -= 0x60
// 		} else {
// 			cpu.Reg.Carry = true
// 			if hi >= 0x180 {
// 				cpu.Reg.Overflow = false
// 			}
// 			hi -= 0x100
// 		}

// 		v = hi | lo

// 	case false:
// 		v = 0xff + acc - sub + carry
// 		if v < 0x100 {
// 			cpu.Reg.Carry = false
// 			if v < 0x80 {
// 				cpu.Reg.Overflow = false
// 			}
// 		} else {
// 			cpu.Reg.Carry = true
// 			if v >= 0x180 {
// 				cpu.Reg.Overflow = false
// 			}
// 		}
// 	}

// 	cpu.Reg.A = byte(v)
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Subtract with Carry (NMOS)
// func (cpu *CPU) sbcn(inst *Instruction, operand []byte) {
// 	acc := uint32(cpu.Reg.A)
// 	sub := uint32(cpu.load(inst.Mode, operand))
// 	carry := boolToUint32(cpu.Reg.Carry)
// 	var v uint32

// 	switch cpu.Reg.Decimal {
// 	case true:
// 		lo := 0x0f + (acc & 0x0f) - (sub & 0x0f) + carry

// 		var carrylo uint32
// 		if lo < 0x10 {
// 			lo -= 0x06
// 			carrylo = 0
// 		} else {
// 			lo -= 0x10
// 			carrylo = 0x10
// 		}

// 		hi := 0xf0 + (acc & 0xf0) - (sub & 0xf0) + carrylo

// 		if hi < 0x100 {
// 			cpu.Reg.Carry = false
// 			hi -= 0x60
// 		} else {
// 			cpu.Reg.Carry = true
// 			hi -= 0x100
// 		}

// 		v = hi | lo

// 		cpu.Reg.Overflow = ((acc^v)&0x80) != 0 && ((acc^sub)&0x80) != 0

// 	case false:
// 		v = 0xff + acc - sub + carry
// 		cpu.Reg.Carry = (v >= 0x100)
// 		cpu.Reg.Overflow = (((acc & 0x80) != (sub & 0x80)) && ((acc & 0x80) != (v & 0x80)))
// 	}

// 	cpu.Reg.A = byte(v)
// 	cpu.updateNZ(byte(v))
// }

// Set Carry flag
// func (cpu *CPU) sec(inst *Instruction, operand []byte) {
// 	cpu.Reg.Carry = true
// }

// Set Decimal flag
// func (cpu *CPU) sed(inst *Instruction, operand []byte) {
// 	cpu.Reg.Decimal = true
// }

// Set InterruptDisable flag
// func (cpu *CPU) sei(inst *Instruction, operand []byte) {
// 	cpu.Reg.InterruptDisable = true
// }

// Store Accumulator
// func (cpu *CPU) sta(inst *Instruction, operand []byte) {
// 	cpu.store(inst.Mode, operand, cpu.Reg.A)
// }

// Store X register
// func (cpu *CPU) stx(inst *Instruction, operand []byte) {
// 	cpu.store(inst.Mode, operand, cpu.Reg.X)
// }

// Store Y register
// func (cpu *CPU) sty(inst *Instruction, operand []byte) {
// 	cpu.store(inst.Mode, operand, cpu.Reg.Y)
// }

// Store Zero (65c02 only)
// func (cpu *CPU) stz(inst *Instruction, operand []byte) {
// 	cpu.store(inst.Mode, operand, 0)
// }

// Transfer Accumulator to X register
// func (cpu *CPU) tax(inst *Instruction, operand []byte) {
// 	cpu.Reg.X = cpu.Reg.A
// 	cpu.updateNZ(cpu.Reg.X)
// }

// Transfer Accumulator to Y register
// func (cpu *CPU) tay(inst *Instruction, operand []byte) {
// 	cpu.Reg.Y = cpu.Reg.A
// 	cpu.updateNZ(cpu.Reg.Y)
// }

// Test and Reset Bits (65c02 only)
// func (cpu *CPU) trb(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Zero = ((v & cpu.Reg.A) == 0)
// 	nv := (v & (cpu.Reg.A ^ 0xff))
// 	cpu.store(inst.Mode, operand, nv)
// }

// Test and Set Bits (65c02 only)
// func (cpu *CPU) tsb(inst *Instruction, operand []byte) {
// 	v := cpu.load(inst.Mode, operand)
// 	cpu.Reg.Zero = ((v & cpu.Reg.A) == 0)
// 	nv := (v | cpu.Reg.A)
// 	cpu.store(inst.Mode, operand, nv)
// }

// Transfer stack pointer to X register
// func (cpu *CPU) tsx(inst *Instruction, operand []byte) {
// 	cpu.Reg.X = cpu.Reg.SP
// 	cpu.updateNZ(cpu.Reg.X)
// }

// Transfer X register to Accumulator
// func (cpu *CPU) txa(inst *Instruction, operand []byte) {
// 	cpu.Reg.A = cpu.Reg.X
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Transfer X register to the stack pointer
// func (cpu *CPU) txs(inst *Instruction, operand []byte) {
// 	cpu.Reg.SP = cpu.Reg.X
// }

// Transfer Y register to the Accumulator
// func (cpu *CPU) tya(inst *Instruction, operand []byte) {
// 	cpu.Reg.A = cpu.Reg.Y
// 	cpu.updateNZ(cpu.Reg.A)
// }

// Unused instruction (6502)
func (cpu *CPU) unusedn(inst *Instruction, operand []byte) {
	// Do nothing
}

// Unused instruction (65c02)
func (cpu *CPU) unusedc(inst *Instruction, operand []byte) {
	// Do nothing
}

//=================== Added, Chris Riddick, 2024 ====================

// GetRegisters returns a formatted string of register values
func (cpu *CPU) GetRegisters() string {
	var s string
	s = s + fmt.Sprintf("R0: x%02x\n", cpu.Reg.R[0])
	s = s + fmt.Sprintf("R1: x%02x\n", cpu.Reg.R[1])
	s = s + fmt.Sprintf("R2: x%02x\n", cpu.Reg.R[2])
	s = s + fmt.Sprintf("R3: x%02x\n", cpu.Reg.R[3])
	s = s + fmt.Sprintf("R4: x%02x\n", cpu.Reg.R[4])
	s = s + fmt.Sprintf("R5: x%02x\n", cpu.Reg.R[5])
	s = s + fmt.Sprintf("R6: x%02x\n", cpu.Reg.R[6])
	s = s + fmt.Sprintf("R7: x%02x\n", cpu.Reg.R[7])
	s = s + fmt.Sprintf("SP: x%02x\n", cpu.Reg.SP)
	s = s + fmt.Sprintf("PC: x%04x\n", cpu.Reg.PC)
	s = s + fmt.Sprintf("Carry: %t\n", cpu.Reg.Carry)
	s = s + fmt.Sprintf("Zero: %t\n", cpu.Reg.Zero)
	s = s + fmt.Sprintf("InterruptDisable: %t\n", cpu.Reg.InterruptDisable)
	s = s + fmt.Sprintf("Decimal: %t\n", cpu.Reg.Decimal)
	s = s + fmt.Sprintf("Overflow: %t\n", cpu.Reg.Overflow)
	s = s + fmt.Sprintf("Sign: %t\n", cpu.Reg.Sign)
	return s
}

// GetStack returns a formatted string of bytes beginning at SP down to to of stack
// 6502 stack grows from $01FF down to $0000
func (cpu *CPU) GetStack() string {
	var s string
	stackbottom := uint16(0x01ff)
	for i := uint16(cpu.Reg.SP) + 0x0100; i < stackbottom; i++ {
		s = s + fmt.Sprintf("%04x: x%02x\n", i, cpu.Mem.LoadByte(i))
	}
	return s
}

// GetAllMemory returns a 16 byte formatted string starting at 0000
func (cpu *CPU) GetAllMemory(addr uint16) string {
	/* var line string
	var buf [256]byte
	var num uint16 = uint16(len(buf) - 1)
	var j uint16 = 0
	cpu.Mem.LoadBytes(addr, buf[0:]) // Copy len(buf) bytes from addr into buf[]
	blocks := num / 16
	remainder := num % 16
	// Send header line with memory locations
	line = "       00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f\n"
	k := addr
	for j = 0; j < blocks; j++ {
		line = line + fmt.Sprintf("%04x:  ", k)
		for i := k; i < k+16; i++ {
			line = line + fmt.Sprintf("%02x ", buf[i])
		}
		line = line + "\n"
		k = k + 16
	}
	if k >= num {
		return line
	}
	endBlock := blocks * 16
	line = line + fmt.Sprintf("%04x:  ", k)
	for i := endBlock; i < endBlock+remainder; i++ {
		line = line + fmt.Sprintf("%02x ", buf[i])
	}
	line = line + "\n" */
	line := "Memory placeholder"
	return line
}

//
//========================== New Opcodes =============================
//

// ADI - Add Immediate, 2's Complement with carry. Register # is last three bits of opcode
// Since this is 2's Complement, we need to do a few more things to be sure we don't go over
// -127 or +127
func (cpu *CPU) adi(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	r := cpu.getReg(inst.Opcode)      // Get reg # from instruction opcode
	//rv := uint32(cpu.Reg.R[r])        // Get value from register r
	//sum := uint32(v) + rv             // Add data value to content of register r
	sum := cpu.twosCompAdd(cpu.Reg.R[r], v) // Add and set flags
	cpu.Reg.R[r] = sum
}

// ADIC is redundant and not needed
// func (c *CPU) adic(inst *Instruction, operand []byte) {
// 	// TBD
// }

// ADM - Add (2's comp w/carry) contents at memory location specified by operand to the register
// from the op code
func (cpu *CPU) adm(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	// addr := operandToAddress(operand) // Get address from operand
	mv := cpu.load(inst.Mode, operand) // Get byte from memory
	cv := cpu.Reg.R[r]                 // retrieve current value from register
	sum := cpu.twosCompAdd(mv, cv)     // internal routine sets the PSR flags
	cpu.Reg.R[r] = sum
}

// ADMC is redundant and not needed
// func (c *CPU) admc(inst *Instruction, operand []byte) {
// 	// TBD
// }

func (cpu *CPU) adr(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand)
	x, y := cpu.getRegXY(v)
	cpu.Reg.R[x] = cpu.Reg.R[x] + cpu.Reg.R[y]
	cpu.updateNZ(cpu.Reg.R[x])
}

// ADRC is redundant and not needed
// func (c *CPU) adrc(inst *Instruction, operand []byte) {
// 	// TBD
// }

// Bitwise AND Register X with Y, result to X, Set zero and neg flags
func (cpu *CPU) and(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand)
	x, y := cpu.getRegXY(v)
	cpu.Reg.R[x] = cpu.Reg.R[x] & cpu.Reg.R[y]
	cpu.updateNZ(cpu.Reg.R[x])
}

// Bitwise AND of Register r with operand. Set Zero and Neg flags.
func (cpu *CPU) ani(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	r := cpu.getReg(inst.Opcode)      // Get reg # from instruction opcode
	result := v & r
	cpu.Reg.R[r] = result
	cpu.updateNZ(cpu.Reg.R[r])
}

func (c *CPU) call(inst *Instruction, operand []byte) {
	// TBD
}

// Compare Registers, Sets Carry flag to true if matched
func (cpu *CPU) cmp(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand)
	x, y := cpu.getRegXY(v)
	cpu.Reg.Carry = (cpu.Reg.R[x] == cpu.Reg.R[y])
}

// Decrement Register by 1. Set N if bit 7 on.Set Z if result is 0. No carry involved.
func (cpu *CPU) dec(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	v := cpu.Reg.R[r]
	v = v - 1
	cpu.updateNZ(v)
	cpu.Reg.R[r] = v
}

// EX - Swap content of two registers
func (cpu *CPU) ex(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand)
	x, y := cpu.getRegXY(v)
	xtemp := cpu.Reg.R[x]
	ytemp := cpu.Reg.R[y]
	cpu.Reg.R[x] = ytemp
	cpu.Reg.R[y] = xtemp
}
func (c *CPU) halt(inst *Instruction, operand []byte) {
	// TBD
}

// Increment memory value
func (cpu *CPU) inc(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) + 1
	cpu.updateNZ(v)
	cpu.store(inst.Mode, operand, v)
}

// LBR - Long Branch to memory address
func (cpu *CPU) lbr(inst *Instruction, operand []byte) {
	addr := operandToAddress(operand)
	cpu.Reg.PC = addr
}

// LBRC - Long Branch w/carry to memory address
func (cpu *CPU) lbrc(inst *Instruction, operand []byte) {
	addr := operandToAddress(operand)
	cpu.Reg.PC = addr
}

func (c *CPU) lbrq(inst *Instruction, operand []byte) {
	// TBD
}

// LBRZ - Long Branch if zero flag
func (cpu *CPU) lbrz(inst *Instruction, operand []byte) {
	if cpu.Reg.Zero {
		addr := operandToAddress(operand)
		cpu.Reg.PC = addr
	}
}

// Load Register Immediate
func (cpu *CPU) ldi(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	r := cpu.getReg(inst.Opcode)      // Get reg # from instruction opcode
	cpu.Reg.R[r] = v                  // Store value in register
	fmt.Printf("Operand: %02x, Reg #: %02x, Reg Content: %02x\n", v, r, cpu.Reg.R[r])
}

func (c *CPU) ldm(inst *Instruction, operand []byte) {
	// TBD
}

// No-operation
func (cpu *CPU) nop(inst *Instruction, operand []byte) {
	// Do nothing
}

// OR registers sppecified by operand and store in R[x]
func (cpu *CPU) or(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	x, y := cpu.getRegXY(v)           // Get reg # from instruction opcode
	result := cpu.Reg.R[x] | cpu.Reg.R[y]
	cpu.Reg.R[x] = result
	cpu.updateNZ(cpu.Reg.R[x])
}

// OR selected reg with byte following opcode
func (cpu *CPU) ori(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	r := cpu.getReg(inst.Opcode)      // Get reg # from instruction opcode
	result := v | r
	cpu.Reg.R[r] = result
	cpu.updateNZ(cpu.Reg.R[r])
}

// Push register
func (cpu *CPU) pushr(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	cpu.push(cpu.Reg.R[r])
}

// Pop register
func (cpu *CPU) popr(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	cpu.Reg.R[r] = cpu.pop()
	cpu.updateNZ(cpu.Reg.R[r])
}

func (cpu *CPU) resetq(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode)       // Get reg # from instruction opcode
	cpu.Reg.Q = bitClear(cpu.Reg.Q, r) // Clear the r bit of Q byte
}

func (c *CPU) ret(inst *Instruction, operand []byte) {
	// TBD
}

func (cpu *CPU) setq(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode)     // Get reg # from instruction opcode
	cpu.Reg.Q = bitSet(cpu.Reg.Q, r) // Set the r bit of Q byte
}

// SHL - Shift content of Register left 1 bit
func (cpu *CPU) shl(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode)                    // Get reg # from instruction opcode
	cpu.Reg.Carry = ((cpu.Reg.R[r] & 0x80) == 0x80) // Set carry if left-most bit was 1
	cpu.Reg.R[r] = cpu.Reg.R[r] << 1
	cpu.updateNZ(cpu.Reg.R[r])
}

// SHLC - Shift content of Register left 1 bit and add Carry bit if set
func (cpu *CPU) shlc(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	cpu.Reg.R[r] = (cpu.Reg.R[r] << 1) + boolToByte(cpu.Reg.Carry)
	cpu.updateNZ(cpu.Reg.R[r])
}

// SHR - Shift content of Register right 1 bit.
func (cpu *CPU) shr(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode)              // Get reg # from instruction opcode
	cpu.Reg.Carry = ((cpu.Reg.R[r] & 1) == 1) // Set carry if right-most bit was 1
	cpu.Reg.R[r] = cpu.Reg.R[r] >> 1
	cpu.updateNZ(cpu.Reg.R[r])
}

// SHRC - Shift content of Register right w/carry.
func (cpu *CPU) shrc(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	cpu.Reg.R[r] = cpu.Reg.R[r] >> 1
	// Check carry flag and put is msb if set
	if cpu.Reg.Carry {
		cpu.Reg.R[r] = cpu.Reg.R[r] | 0x80
	}
	cpu.updateNZ(cpu.Reg.R[r])
}

// SPSR - Set Program Status Register bits
// Bit 0 - Carry
// Bit 1 - Zero
// Bit 2 - InterruptDisable
// Bit 3 - Decimal
// Bit 4 - Break
// Bit 5 - Reserved
// Bit 6 - Overflow
// Bit 7 - Sign
func (cpu *CPU) spsr(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get bits from operand
	switch v {
	case CarryBit:
		cpu.Reg.Carry = true
	case ZeroBit:
		cpu.Reg.Zero = true
	case SignBit:
		cpu.Reg.Sign = true
	case OverflowBit:
		cpu.Reg.Overflow = true
	case BreakBit:
		cpu.Reg.Decimal = true
	case InterruptDisableBit:
		cpu.Reg.InterruptDisable = true
	case DecimalBit:
		cpu.Reg.Decimal = true
	}
}

// CPSR - Clear specified bit is Processor Status Register
func (cpu *CPU) cpsr(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get bits from operand
	switch v {
	case CarryBit:
		cpu.Reg.Carry = false
	case ZeroBit:
		cpu.Reg.Zero = false
	case SignBit:
		cpu.Reg.Sign = false
	case OverflowBit:
		cpu.Reg.Overflow = false
	case BreakBit:
		cpu.Reg.Decimal = false
	case InterruptDisableBit:
		cpu.Reg.InterruptDisable = false
	case DecimalBit:
		cpu.Reg.Decimal = false
	}
}

// Store Register value from opcode into address specified by two-byte operand
func (cpu *CPU) sti(inst *Instruction, operand []byte) {
	r := cpu.getReg(inst.Opcode) // Get reg # from instruction opcode
	addr := operandToAddress(operand)
	cpu.Mem.StoreByte(addr, cpu.Reg.R[r])
	//fmt.Printf("Address to store at: %04x, Reg #: %02x, Reg Content: %02x\n", addr, r, cpu.Reg.R[r])
}

func (c *CPU) sub(inst *Instruction, operand []byte) {
	// TBD
	/*
		The SBC (subtraction with carry) instruction is actually a sub‐ traction with BORROW,
		if we use mathematically correct terminology. The symbolic operation for SBC is
		A*M*_G-*A
		This notation says that the value fetched from memory (M) and the complement of the
		carry flag (G) is subtracted from the contents of the accumulator, and the result is
		stored in the accumulator. Note that the carry flag will be set (HIGH) if a result is
		equal to or greater than zero, and reset (LOW) if the results are less than zero, i.e., negative.
		The SBC instruction has available all 8 Group-I addressing modes, aswas also true of ADC.
		The SBC instruction affects the following PSR flags: negative (N), zero (Z), Carry (C), and
		overflow (V). The N-flag indicates a negative result and will be HIGH; the Z-flag is H I G H
		if the result of the SBC instruction is zero and LOW otherwise; the overflow flag (V) is HIGH
		when the result exceeds the values 7FH (+12710) and 80H with C = 1 (i.e., ‐ 12810).
		The 6502 manufacturer recommends for single-precision (8-bit) subtracts that the programmer
		ensure that the carry flag is set prior to the SBC operation to be sure that true two’s complement
		arithmetic takes place. We can set the carry flag by executing the SEC (set carry flag) instruction.
	*/
}

// SUBC is redundant and not needed
// func (c *CPU) subc(inst *Instruction, operand []byte) {
// 	// TBD
// }

func (c *CPU) subi(inst *Instruction, operand []byte) {
	// TBD
}

// SUBIC is redundant and not needed
// func (c *CPU) subic(inst *Instruction, operand []byte) {
// 	// TBD
// }

func (c *CPU) subm(inst *Instruction, operand []byte) {
	// TBD
}

// SUBMC is redundant and not needed
// func (c *CPU) submc(inst *Instruction, operand []byte) {
// 	// TBD
// }

// XOR registers sppecified by operand and store in R[x]
func (cpu *CPU) xor(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	x, y := cpu.getRegXY(v)           // Get reg # from instruction opcode
	result := cpu.Reg.R[x] ^ cpu.Reg.R[y]
	cpu.Reg.R[x] = result
	cpu.updateNZ(cpu.Reg.R[x])
}

// XOR selected reg with byte following opcode
func (cpu *CPU) xri(inst *Instruction, operand []byte) {
	v := cpu.load(inst.Mode, operand) // Get value from operand
	r := cpu.getReg(inst.Opcode)      // Get reg # from instruction opcode
	result := v ^ r
	cpu.Reg.R[r] = result
	cpu.updateNZ(cpu.Reg.R[r])
}
