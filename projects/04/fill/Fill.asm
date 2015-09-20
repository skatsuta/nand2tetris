// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input. 
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel. When no key is pressed, the
// program clears the screen, i.e. writes "white" in every pixel.

// Put your code here.

// define max address in a screen
@24576
D=A
@MAXSCREEN
M=D

// define screen pointer
@SCREEN
D=A
@POINTER
M=D

// infinite loop
(LOOP)
    // jump to FILL if keyboard input
    @KBD
    D=M
    @FILL
    D;JGT

    // otherwise jump to UNFILL
    @UNFILL
    0;JMP

(UNFILL)
    // do nothing if POINTER == SCREEN
    @SCREEN
    D=A
    @POINTER
    D=D-M
    @LOOP
    D;JEQ

    // unfill the screen
    @POINTER
    A=M
    M=0

    // decrement POINTER
    @POINTER
    M=M-1

    // jump back to main loop
    @LOOP
    0;JMP

(FILL)
    // do nothing if the screen is full
    @MAXSCREEN
    D=M
    @POINTER
    D=D-M
    @LOOP
    D;JEQ

    // fill in the pixel that POINTER points to
    @POINTER
    A=M
    M=-1 // -1 == 0xFFFF (all 1)

    // iterate pointer by 1
    @POINTER
    M=M+1

    // jump back to main loop
    @LOOP
    0;JMP
