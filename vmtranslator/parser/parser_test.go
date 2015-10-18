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

label LABEL0
  goto 		 END  	
if-goto SYMBOL

function func1 0
call func2 1

return
`

func TestHasMoreCommands(t *testing.T) {
	p := New(strings.NewReader(testVM))

	testCases := []struct {
		next   bool
		tokens []string
	}{
		{true, []string{"push", "constant", "2"}},
		{true, []string{"push", "constant", "3"}},
		{true, []string{"add"}},
		{true, []string{"push", "constant", "1"}},
		{true, []string{"sub"}},
		{true, []string{"label", "LABEL0"}},
		{true, []string{"goto", "END"}},
		{true, []string{"if-goto", "SYMBOL"}},
		{true, []string{"function", "func1", "0"}},
		{true, []string{"call", "func2", "1"}},
		{true, []string{"return"}},
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
		{"push local 3", command{Push, "local", 3}},
		{"pop   argument		4", command{Pop, "argument", 4}},
		{"label LABEL0", command{Label, "LABEL0", 0}},
		{"goto END", command{Goto, "END", 0}},
		{"if-goto SYMBOL", command{If, "SYMBOL", 0}},
		{"function func1 0", command{Function, "func1", 0}},
		{"call func2 1", command{Call, "func2", 1}},
		{"return", command{Return, "", 0}},
	}

	for _, tt := range testCases {
		p := New(strings.NewReader(tt.src))
		if p.HasMoreCommands() {
			if e := p.Advance(); e != nil {
				t.Errorf("advance failed: %s", e.Error())
			}
			if !reflect.DeepEqual(p.cmd, tt.want) {
				t.Errorf("src = %q; got: %+v; want: %+v", tt.src, p.cmd, tt.want)
			}
		}
	}
}

func TestAdvanceError(t *testing.T) {
	testCases := []struct {
		src string
	}{
		{""},
		{"foo"},
		{"push"},
		{"label"},
		{"goto"},
		{"if-goto"},
		{"function"},
		{"call"},
		{"add sub"},
		{"pop temp"},
		{"function func1"},
		{"call func2"},
		{"return 0"},
		{"posh constant 1"},
		{"pop argment 0"},
		{"push local a"},
		{"label L 0"},
		{"goto G 1"},
		{"if-goto I 2"},
		{"function func1 -1"},
		{"call func1 a"},
		{"push constant 1 2"},
		{"function func1 2 3"},
	}

	for _, tt := range testCases {
		p := New(strings.NewReader(tt.src))
		if p.HasMoreCommands() {
			if e := p.Advance(); e == nil {
				t.Errorf("src = %q: expected error but got <nil>", tt.src)
			}
		}
	}
}
