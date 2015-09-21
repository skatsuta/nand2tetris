package parser

import (
	"io"
	"os"
	"strings"
	"testing"
)

var testAsm = `

// This is a comment.
@16

// This is also a comment.
D=M

(LOOP)
    @17 // indent spaces
	D=A // indent tab

	@LOOP
	0;JMP
`

func TestNewParser(t *testing.T) {
	filename := "../../add/Add.asm"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open %s: %s", filename, err.Error())
	}

	newParserTests := []struct {
		r io.Reader
	}{
		{strings.NewReader(testAsm)},
		{file},
	}

	for _, tt := range newParserTests {
		got := newParser(tt.r)
		if got.in == nil {
			t.Errorf("input is nil")
		}
	}
}

func TestHasMoreCommands(t *testing.T) {
	a := newParser(strings.NewReader(testAsm))
	numOfCmdsInTestAsm := 7
	var cnt int

	for a.hasMoreCommands() {
		cnt++
	}

	if cnt != numOfCmdsInTestAsm {
		t.Errorf("# of commands in testAsm should be %d, but got %d", numOfCmdsInTestAsm, cnt)
	}
}
