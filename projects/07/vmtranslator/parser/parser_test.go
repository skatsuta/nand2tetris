package parser

import (
	"reflect"
	"strings"
	"testing"
)

var testVM = `
// comment
  push  constant	2
  // comment 2
push	 constant 3 	
	add 	// inline comment
  
push  constant	 1
sub	 //  inline comment 2
`

/*
func TestNewParser(t *testing.T) {
	filename := "../../StackArithmetic/SimpleAdd/SimpleAdd.vm"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open %s: %s", filename, err.Error())
	}

	testCases := []struct {
		r io.Reader
	}{
		{strings.NewReader(testVM)},
		{file},
	}

	for _, tt := range testCases {
		got := NewParser(tt.r)
		if got.in == nil {
			t.Error("input is nil")
		}
	}
}
*/

func TestHasMoreCommands(t *testing.T) {
	p := NewParser(strings.NewReader(testVM))

	testCases := []struct {
		next   bool
		tokens []string
	}{
		{true, []string{"push", "constant", "2"}},
		{true, []string{"push", "constant", "3"}},
		{true, []string{"add"}},
		{true, []string{"push", "constant", "1"}},
		{true, []string{"sub"}},
		{false, []string{}},
	}

	for _, tt := range testCases {
		if p.HasMoreCommands() != tt.next {
			t.Errorf("HasMoreCommands should return %t, but %t", tt.next, !tt.next)
		}
		if !reflect.DeepEqual(p.tokens, tt.tokens) {
			t.Errorf("got %#v; want %#v", p.tokens, tt.tokens)
		}
	}
}

func TestAdvance(t *testing.T) {
	testCases := []struct {
		src  string
		want command
	}{
		{"add", command{Arithmetic, "add", 0}},
		{"sub", command{Arithmetic, "sub", 0}},
		{"push constant 1", command{Push, "constant", 1}},
		{"pop constant 2", command{Pop, "constant", 2}},
	}

	for _, tt := range testCases {
		p := NewParser(strings.NewReader(tt.src))
		if p.HasMoreCommands() {
			if e := p.Advance(); e != nil {
				t.Errorf("advance failed: %s", e.Error())
			}
			if !reflect.DeepEqual(p.cmd, tt.want) {
				t.Errorf("got: %+v; want: %+v", p.cmd, tt.want)
			}
		}
	}
}

func TestAdvanceError(t *testing.T) {
	testCases := []struct {
		src string
	}{
		{"foo"},
		{"add sub"},
		{"push constant 1 2"},
		{"posh constant 1"},
		{"pop argment 0"},
		{"push local a"},
	}

	for _, tt := range testCases {
		p := NewParser(strings.NewReader(tt.src))
		if p.HasMoreCommands() {
			if e := p.Advance(); e == nil {
				t.Errorf("expected error but got <nil>")
			}
		}
	}
}
