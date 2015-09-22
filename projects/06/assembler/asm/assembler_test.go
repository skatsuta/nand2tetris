package asm

import (
	"bytes"
	"strings"
	"testing"
)

// test assembly code
var testAsm = `

// This is a comment.
@16  \t

// This is also a comment.
D=M     
MD=0   
AD=D|M;JGE\t
AMD=D+1;JLT   \t
M+1;JLT
-1;JGT


(LOOP)     
	@17 // indent spaces
	D=A // indent tab

	@LOOP\t\t    
	D;JEQ

(END)
	0;JMP
`

// expected binary code converted from testAsm
var wantHack = `0000000000010000
1111110000010000
1110101010011000
1111010101110011
1110011111111100
1111110111000100
1110111010000001
0000000000010001
1110110000010000
1110101010000111
`

func TestRun(t *testing.T) {
	runTests := []struct {
		src  string
		want string
	}{
		{"@1", "0000000000000001"},
		{"@256", "0000000100000000"},
		// {testAsm, wantHack},
	}

	for _, tt := range runTests {
		asmblr := New(strings.NewReader(tt.src))
		var out bytes.Buffer

		if e := asmblr.Run(&out); e != nil {
			t.Fatalf("Run failed: %s", e.Error())
		}

		got := out.String()
		if got != tt.want {
			t.Errorf("src: %s\ngot:\n%s\nwant:\n%s", tt.src, got, tt.want)
		}
	}
}
