# cpu1-simulator
My first custom CPU simulator. CPU1 is an imaginary 8-bit processor.

# Introduction

This CPU simulator implements an imaginary 8-bit processor and instruction set. Much of the code is derived from work by Brett Vickers and posted on github [here](https://github.com/beevik/go6502).

## Architecture description

This cpu, called CPU1, has an 8-bit data bus and a 16-bit address bus. It has an accumulator, flag register, and eight 8-bit general purpose registers. There is a 16-bit program counter and an 8-bit stack pointer. The stack is limited in size and always grows downward from $01FF. Programs execute from any address beginning at $0200. Memory addresses are stored Big Endian, with most significant byte at lower address in memory.

## Instruction Set
In the description below the following symbolic bits are used:
**V** - value (bits encoding literal value); 
**R** - register (registers are encoded according to the pattern: 0 - R0, 1 - R1, etc.); 
**X** - ignored; 
**M** - memory address; 
**D** - data bit;
**C** - carry bit; 

The table below describes the complete instruction set, together with bit patterns:

Nemonic|Opcode (hex)|Opcode (binary)|    Operand(s)  |    Description
-------|------------|---------------|----------------|--------------------------------------------
ADR|80|10000000|XRRRXRRR|RX <- RX + RY Add rgisters specified by operand lo and hi nibbles
ADI|88,89,8A,8B,8C,8D,8E,8F|10001RRR|DDDDDDDD|R <- R + (PC+1); Add immediate
ADM|90,91,92,93,94,95,96,97|10001RRR|MMMMMMMM MMMMMMMM|R <- R + (M); Add memory at address to R
ADRC|81|10000001|XRRRXRRR|RX <- RX + RY + C; Add registers with carry bit
ADIC|A0,A1,A2,A3,A4,A5,A6,A7|10100RRR|DDDDDDDD|R <- R + (PC+1) + C; Add w/carry immediate
ADMC|A8,A9,AA,AB,AC,AD,AE,AF|10101RRR|MMMMMMMM MMMMMMMM|R <- R + (M) + C; Add w/carry the byte at M
SUB|82|10000010|XRRRXRRR|RX <- RX - RY; Subtract RX from RY. Set carry and negative flags
SUBI|B8,B9,B1,BB,BC,BD,BE,BF|10111RRR|DDDDDDDD|R <- R - (PC+1); Subtract immediate. Set carry and neg flags
SUBM|C0,C1,C2,C3,C4,C5,C6,C7|11000RRR|MMMMMMMM MMMMMMMM|R <- R - (M); Subtract memory. Set carry and neg flags
SUBC|83|1000011|XRRRXRRR|RX <- RX - RY - C - (NOT C); Subtract register w/borrow from carry bit
SUBIC|D0,D1,D2,D3,D4,D5,D6,D7|11010RRR|DDDDDDDD|R <- R - (PC+1) - C - (NOT C); Subtract immediate w/borrow, flags
SUBMC|D8,D9,DA,DB,DC,DD,DE,DF|11011RRR|MMMMMMMM MMMMMMMM|R <- R - (M) - C - (NOT C); Sub immed w/borrow from carry
LDI|E0,E1,E2,E3,E4,E5,E6,E7|11100RRR|DDDDDDDD|R <- (PC+1); Load immediate into R
STI|E8,E9,EA,EB,EC,ED,EE,EF|11101RRR|MMMMMMMM MMMMMMMM|(M) <- R; Store immediate R at M
LDM|F0,F1,F2,F3,F4,F5,F6,F7|11110RRR|MMMMMMMM MMMMMMMM|R <- (M); Load from memory into R
EX|84|10000100|XRRRXRRR|RX <- RY; RY <- RX; Exchange registers
CMP|85|10000100|XRRRXRRR|IF RX=RY,CP=TRUE,ELSE CP=FALSE; Compare registers and set compare flag if equal
AND|86|10000110|XRRRXRRR|RX <-RX AND RY; AND: Logical AND of RX and RY. Result to RX
OR|87|10000111|XRRRXRRR|RX <- RX OR RY; OR: Logical OR of RX and RY. Result to RX
XOR|19|00011001|XRRRXRRR|RX <- RX XOR RY; XPR: Exclusive OR of RX and RY. Result to RX
ANI|50,51,52,53,54,55,56,57|01010RRR|DDDDDDDD|R <- R AND (PC+1); AND immediate. Result to R
ORI|58,59,5A,5B,5C,5D,5E,5F|01011RRR|DDDDDDDD|R <- R OR (PC+1); OR immediate. Result in R
XRI|60,61,62,63,64,65,66,67|01100RRR|DDDDDDDD|R <- R XOR (PC+1); XOR immediate. Result in R
SHR|68,69,6A,6B,6C,6D,6E,6F|01101RRR||R <- R>>1;Shift right reg R by one bit. Fill w/zero on left
SHRC|70,71,72,73,74,75,76,77|01110RRR||R <- R>>1:Shift right reg R by one. Fill left with carry bit
SHL|78,79,7A,7B,7C,7D,7E,7F|01111RRR||R <- R<<1; Shift left reg R one bit. Fill least sig with 0
SHLC|20,21,22,,23,24,25,26,27|00100RRR||R <- R<<1; Shift left reg R one bit, fill lsb with carry bit
INC|28,29,2A,2B,2C,2D,2E,2F|00101RRR||R <- R + 1; Increment reg R by 1
DEC|30,31,32,33,34,35,36,37|00110RRR||R <- R - 1; Decrement reg R by 1
NOP|00|00000000||PC <- PC + 1; Continue to next instruction
HALT|01|00000001||PC <- PC; Stop CPU clock and instruction execution at current PC
SETQ|38,39,3A,3B,3C,3D,3E,3F|00111QQQ||QN <- true(1); Sets specified I/O line to true(1)
RESETQ|10,11,12,13,14,15,16,17|00010QQQ||QN <- false(0); Sets specified I/O liine to false(0)
LBRC|18|00011000|MMMMMMMM MMMMMMMM|If CP=true, PC <- M, else PC <- PC+2;Long branch if compare flag true
LBRQ|08,09,0A,0B,0C,0D,0E,0F|00001QQQ|MMMMMMMM MMMMMMMM|IF QN, PC <- M, else PC <- PC + 2; Long branch if true
PUSH|40,41,42,43,44,45,46,47|01000RRR||SP <- SP-1; (SP) <- R; Push register onto stack
POP|48,49,4A,,4B,4C,4D,4E,4F|01001RRR||R <- (SP); SP <- SP + 1; Pop register from stack
CALL|02|00000010|MMMMMMMM MMMMMMMM|PC <- PC+3,SP <- SP-1;(SP) <- PC; Call subroutine, save PC on stack (Big Endian)
RET|03|00000011||PC.1 <- (SP),SP+1,PC.0 <- (SP), SP+1; Return from subroutine popping PC off stack (Big Endian)

## Assembler


## GUI Dashboard


Here is a sample of what you will see on the dashboard
![Dashboard](./dashboard.png)

## Building
First be sure the latest version of golang is installed.
```
$ sudo rm -rf /usr/local/go && curl -sSL "https://go.dev/dl/go1.21.6.linux-arm64.tar.gz" | sudo tar -xz -C /usr/local
$ echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
$ source $HOME/.profile
$ go version
go version go1.21.6 linux/arm64
```
Clone the github repo for go-cpu-simulator
```
$ cd go-cpu-simulator
$ go mod tidy
$ go run go-cpu-simulator
```


