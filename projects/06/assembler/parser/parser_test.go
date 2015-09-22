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
		got := NewParser(tt.r)
		if got.in == nil {
			t.Errorf("input is nil")
		}
	}
}

func TestHasMoreCommands(t *testing.T) {
	p := NewParser(strings.NewReader(testAsm))
	numOfCmdsInTestAsm := 7
	var cnt int

	for p.HasMoreCommands() {
		cnt++
	}

	if cnt != numOfCmdsInTestAsm {
		t.Errorf("# of commands in testAsm should be %d, but got %d", numOfCmdsInTestAsm, cnt)
	}
}

func TestAdvance(t *testing.T) {
	advanceTests := []command{
		{cmd: "@16", typ: aCommand, symb: "16"},
		{cmd: "D=M", typ: cCommand, dest: "D", comp: "M"},
		{cmd: "(LOOP)", typ: lCommand, symb: "LOOP"},
		{cmd: "@17", typ: aCommand, symb: "17"},
		{cmd: "D=A", typ: cCommand, dest: "D", comp: "A"},
		{cmd: "@LOOP", typ: aCommand, symb: "LOOP"},
		{cmd: "0;JMP", typ: cCommand, comp: "0", jump: "JMP"},
	}

	p := NewParser(strings.NewReader(testAsm))
	for _, want := range advanceTests {
		if p.HasMoreCommands() {
			if e := p.Advance(); e != nil {
				t.Errorf("advance failed: %s", e.Error())
			}
			if !reflect.DeepEqual(p.command, want) {
				t.Errorf("got: %+v; want: %+v", p.command, want)
			}
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

	var p Parser
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

	var p Parser
	for _, tt := range splitCmdTests {
		got := p.splitCmd(tt.cmd, tt.sep)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got: %v; want: %v", got, tt.want)
		}
	}
}
