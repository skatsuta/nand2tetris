Hack Assembler
====

Hack assembler written in Go.

## Description

This subdirectory contains Hack assembler written in Go. The program reads `.asm` files in Hack assembly language, converts it to Hack machine codes and write it out as `.hack` files.

This assembler can treat multiple files at once and process them in parallel.

`asm`, `code`, `parser` and `symbtbl` packages can also be used as libraries.

## Requirement

- Go 1.5+

## Installation

```sh
$ go get github.com/skatsuta/nand2tetris/projects/06/assembler
```

## Usage

To run the assembler, 

```sh
$ assembler file.asm [files...]
```

It generates Hack machine code file named `file.hack`.

For example, if your input file `file.asm` is

```asm
   @R0
   D=M              // D = first number
   @R1
   D=D-M            // D = first number - second number
   @OUTPUT_FIRST
   D;JGT            // if D>0 (first is greater) goto output_first
   @R1
   D=M              // D = second number
   @OUTPUT_D
   0;JMP            // goto output_d
(OUTPUT_FIRST)
   @R0
   D=M              // D = first number
(OUTPUT_D)
   @R2
   M=D              // M[2] = D (greatest number)
(INFINITE_LOOP)
   @INFINITE_LOOP
   0;JMP            // infinite loop
```

then the assembler generates `file.hack` such as

```bin
0000000000000000
1111110000010000
0000000000000001
1111010011010000
0000000000001010
1110001100000001
0000000000000001
1111110000010000
0000000000001100
1110101010000111
0000000000000000
1111110000010000
0000000000000010
1110001100001000
0000000000001110
1110101010000111
```

## Licence

[MIT](https://github.com/skatsuta/nand2tetris/blob/master/LICENCE)

## Author

[Soshi Katsuta (skatsuta)](https://github.com/skatsuta)