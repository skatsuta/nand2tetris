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

	hmcTests := []struct {
		want string
	}{
		{"@16"},
		{"D=M"},
		{"(LOOP)"},
		{"@17 // indent spaces"},
		{"D=A // indent tab"},
		{"@LOOP"},
		{"0;JMP"},
	}

	for _, tt := range hmcTests {
		if !p.HasMoreCommands() {
			t.Errorf("HasMoreCommands should not return false: %s", tt.want)
		}
		if p.line != tt.want {
			t.Errorf("expected %q but got %q", tt.want, p.line)
		}
	}
}

func TestAdvance(t *testing.T) {
	advanceTests := []command{
		{cmd: "@16", typ: ACommand, symb: "16"},
		{cmd: "D=M", typ: CCommand, dest: "D", comp: "M"},
		{cmd: "(LOOP)", typ: LCommand, symb: "LOOP"},
		{cmd: "@17", typ: ACommand, symb: "17"},
		{cmd: "D=A", typ: CCommand, dest: "D", comp: "A"},
		{cmd: "@LOOP", typ: ACommand, symb: "LOOP"},
		{cmd: "0;JMP", typ: CCommand, comp: "0", jump: "JMP"},
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

func TestROMAddr(t *testing.T) {
	p := NewParser(strings.NewReader(testAsm))

	romAddrTests := []struct {
		want string
		addr uintptr
	}{
		{"@16", 0x0},
		{"D=M", 0x1},
		{"(LOOP)", 0x1},
		{"@17", 0x2},
		{"D=A", 0x3},
		{"@LOOP", 0x4},
		{"0;JMP", 0x5},
	}

	for _, tt := range romAddrTests {
		if p.HasMoreCommands() {
			if e := p.Advance(); e != nil {
				t.Fatalf("Advance failed: %s", e.Error())
			}
		}

		if p.command.cmd != tt.want {
			t.Errorf("command: got = %s but want = %s", p.command.cmd, tt.want)
		}
		if p.ROMAddr() != tt.addr {
			t.Errorf("ROM address: got = 0x%X but want = 0x%X", p.ROMAddr(), tt.addr)
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
