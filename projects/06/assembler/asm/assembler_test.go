package asm

import (
	"bytes"
	"strings"
	"testing"
)

// test assembly code
var testAsm = `

// This is a comment.
@256  	

// This is also a comment.
D=M     
@i
MD=0   
@j
AD=D|M;JGE	
@i
AMD=D+1;JLT   	
@j
M+1;JLT
@END
D;JLE

(LOOP)     
	@17 // indent spaces
	D=A // indent tab

	@LOOP		    
	D;JEQ

(END)
	@END
	0;JMP
`

// expected binary code converted from testAsm
var wantHack = `0000000100000000
1111110000010000
0000000000010000
1110101010011000
0000000000010001
1111010101110011
0000000000010000
1110011111111100
0000000000010001
1111110111000100
0000000000010001
1110001100000110
0000000000010001
1110110000010000
0000000000001100
1110001100000010
0000000000010001
1110101010000111
`

func TestDefineSymbols(t *testing.T) {
	symtbl := map[string]uintptr{
		"SP":     0x0,
		"LCL":    0x1,
		"ARG":    0x2,
		"THIS":   0x3,
		"THAT":   0x4,
		"R0":     0x0,
		"R4":     0x4,
		"R5":     0x5,
		"R15":    0xF,
		"SCREEN": 0x4000,
		"KBD":    0x6000,
	}

	asmblr, err := New(strings.NewReader(testAsm))
	if err != nil {
		t.Fatalf("New failed: %s", err.Error())
	}

	asmblr.DefineSymbols(symtbl)

	for symb, addr := range symtbl {
		got := asmblr.st.GetAddress(symb)
		if got != addr {
			t.Errorf("got = 0x%X; want = 0x%X", got, addr)
		}
	}
}

func TestRun(t *testing.T) {
	runTests := []struct {
		src  string
		want string
	}{
		{"@1", "0000000000000001\n"},
		{"@256", "0000000100000000\n"},
		{"(LOOP)", ""},
		{"(END)", ""},
		{"A=!A", "1110110001100000\n"},
		{"M=M-D", "1111000111001000\n"},
		{"1;JMP", "1110111111000111\n"},
		{"A-1;JNE", "1110110010000101\n"},
		{"AM=D&A;JLE", "1110000000101110\n"},
		{"@i\n@j", "0000000000010000\n0000000000010001\n"},
		{"(LOOP)\nD=0\n@LOOP", "1110101010010000\n0000000000000000\n"},
		{"@32\nM=1\n@a\nMD=-1",
			"0000000000100000\n1110111111001000\n0000000000010000\n1110111010011000\n"},
		{testAsm, wantHack},
	}

	for _, tt := range runTests {
		var out bytes.Buffer
		asmblr, err := New(strings.NewReader(tt.src))
		if err != nil {
			t.Fatalf("New failed: %s", err.Error())
		}

		if e := asmblr.Run(&out); e != nil {
			t.Fatalf("%s", e.Error())
		}

		got := strings.Split(out.String(), "\n")
		want := strings.Split(tt.want, "\n")
		if len(got) != len(want) {
			t.Fatalf("the number of lines should be %d, but got %d", len(want), len(got))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("\nsrc:\n%s\n\nline %d: got: %s != want: %s", tt.src, i+1, got[i], want[i])
			}
		}
	}
}
