package vmtranslator

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		src []io.Reader
	}{
		{[]io.Reader{strings.NewReader("foo")}},
		{[]io.Reader{strings.NewReader("foo"), strings.NewReader("bar")}},
	}

	var (
		vmtransl *VMTranslator
		out      bytes.Buffer
	)
	for _, tt := range testCases {
		vmtransl = New(tt.src, &out)

		lv, ls := len(vmtransl.parsers), len(tt.src)
		if lv != ls {
			t.Errorf("length of parsers: got = %d, but want = %d", lv, ls)
		}
		if vmtransl.cw == nil {
			t.Errorf("VMTranslator.cw is nil")
		}
	}
}

var (
	wantPushConst0 = `@0
D=A
@SP
A=M
M=D
@SP
AM=M+1
`

	wantAdd = `@1
D=A
@SP
A=M
M=D
@SP
AM=M+1
@2
D=A
@SP
A=M
M=D
@SP
AM=M+1
@SP
AM=M-1
D=M
@SP
AM=M-1
M=D+M
@SP
AM=M+1
`
)

func TestRun(t *testing.T) {
	testCases := []struct {
		src  string
		want string
	}{
		{"push constant 0", wantPushConst0},
		{"push constant 1\npush constant 2\nadd", wantAdd},
	}

	var (
		buf      bytes.Buffer
		vmtransl *VMTranslator
	)
	for _, tt := range testCases {
		vmtransl = New([]io.Reader{strings.NewReader(tt.src)}, &buf)
		if e := vmtransl.Run(); e != nil {
			t.Fatalf("Run failed: %v", e)
		}

		got := buf.String()
		if got != tt.want {
			t.Errorf("src = %s\n\ngot =\n%s\nwant =\n%s", tt.src, got, tt.want)
		}

		buf.Reset()
	}
}
