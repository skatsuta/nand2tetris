package parser

import (
	"io"
	"os"
	"reflect"
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
	p := newParser(strings.NewReader(testAsm))
	numOfCmdsInTestAsm := 7
	var cnt int

	for p.hasMoreCommands() {
		cnt++
	}

	if cnt != numOfCmdsInTestAsm {
		t.Errorf("# of commands in testAsm should be %d, but got %d", numOfCmdsInTestAsm, cnt)
	}
}

func TestAdvance(t *testing.T) {
	advanceTests := []command{
		nil,
		aCmd{baseCmd{cmd: "@16", typ: aCommand}},
		cCmd{baseCmd{cmd: "D=M", typ: cCommand}},
		lCmd{baseCmd{cmd: "(LOOP)", typ: lCommand}},
		aCmd{baseCmd{cmd: "@17", typ: aCommand}},
		cCmd{baseCmd{cmd: "D=A", typ: cCommand}},
		aCmd{baseCmd{cmd: "@LOOP", typ: aCommand}},
		cCmd{baseCmd{cmd: "0;JMP", typ: cCommand}},
	}

	p := newParser(strings.NewReader(testAsm))
	for _, want := range advanceTests {
		if e := p.advance(); e != nil {
			t.Errorf("advance failed: %s", e.Error())
		}
		if !reflect.DeepEqual(p.cmd, want) {
			t.Errorf("got: %+v; want: %+v", p.cmd, want)
		}
	}
}

func TestTrimComment(t *testing.T) {
	trimCommentTests := []struct {
		line string
		want string
	}{
		{"  D=A  // comment", "D=A"},
		{"@10", "@10"},
	}

	var p parser
	for _, tt := range trimCommentTests {
		got := p.trimComment(tt.line)
		if got != tt.want {
			t.Errorf(`got: "%s"; want: "%s"`, got, tt.want)
		}
	}
}

func TestSplitCmd(t *testing.T) {
	splitCmdTests := []struct {
		cmd  string
		sep  string
		want []string
	}{
		{"@10", "=", []string{"@10"}},
		{"M=D", "=", []string{"M", "D"}},
		{"MD=0", "=", []string{"MD", "0"}},
		{"AMD=M+1", "=", []string{"AMD", "M+1"}},
		{"0;JMP", ";", []string{"0", "JMP"}},
		{"D;JEQ", ";", []string{"D", "JEQ"}},
		{"M=D;JGT", "=", []string{"M", "D;JGT"}},
		{"M=D;JGT", ";", []string{"M=D", "JGT"}},
	}

	var p parser
	for _, tt := range splitCmdTests {
		got := p.splitCmd(tt.cmd, tt.sep)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got: %v; want: %v", got, tt.want)
		}
	}
}
