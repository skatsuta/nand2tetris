// if R16 == R17 {
//   true
// } else {
//   false
// }

// push constant 2
@3
D=A
@16
M=D

// push constant 2
@2
D=A
@17
M=D

// SP = 18
@18
D=A
@SP
M=D

// pop R17
@SP
AM=M-1
D=M

// pop R16
@SP
AM=M-1

// R16 == R17?
D=M-D
@L
D;JEQ

// false
@SP
A=M
M=0
@FIN
0;JMP

// true
(L)
@SP
A=M
M=-1

// inc SP
(FIN)
@SP
M=M+1

