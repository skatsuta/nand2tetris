// R16 = 2 + 3

// push constant 2
@2
D=A
@16
M=D

// push constant 3
@3
D=A
@17
M=D

// SP = 18
@18
D=A
@SP
M=D

//--- add ---
// pop R17
@SP
AM=M-1
D=M
// pop R16
@SP
AM=M-1
M=D+M
// add
@SP
M=M+1
