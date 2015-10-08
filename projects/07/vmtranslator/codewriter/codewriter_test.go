package codewriter

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSetFileName(t *testing.T) {
	testCases := []struct {
		filename string
		want     string
	}{
		{"foo.txt", "\n// foo.txt\n"},
	}

	var buf bytes.Buffer
	cw := New(&buf)
	for _, tt := range testCases {
		if e := cw.SetFileName(tt.filename); e != nil {
			t.Fatalf("SetFileName failed: %v", e)
		}
		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %v", e)
		}

		got := buf.String()
		if got != tt.want {
			t.Errorf("got = %q; want = %q", got, tt.want)
		}

		buf.Reset()
	}
}

func TestWriteArithmetic(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"add", asmBinary("M=D+M")},
		{"sub", asmBinary("M=M-D")},
		{"and", asmBinary("M=D&M")},
		{"or", asmBinary("M=D|M")},
		{"neg", asmUnary("-")},
		{"not", asmUnary("!")},
		{"eq", asmCompare("JEQ", "LABEL0", "LABEL1")},
		{"gt", asmCompare("JGT", "LABEL0", "LABEL1")},
		{"lt", asmCompare("JLT", "LABEL0", "LABEL1")},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		if e := cw.WriteArithmetic(tt.cmd); e != nil {
			t.Fatalf("WriteArithmetic failed: %s", e.Error())
		}
		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %s", e.Error())
		}

		got := buf.String()
		if got != tt.want {
			t.Errorf("src = %s\ngot =\n%s\nwant =\n%s", tt.cmd, got, tt.want)
		}
	}
}

func TestWriteArithmeticError(t *testing.T) {
	testCases := []struct {
		cmd string
	}{
		{"foo"},
		{"bar"},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		if e := cw.WriteArithmetic(tt.cmd); e == nil {
			t.Fatalf("WriteArithmetic should return error: cmd = %s", tt.cmd)
		}
	}
}

func TestWritePushPop(t *testing.T) {
	testCases := []struct {
		cmd  string
		seg  string
		idx  uint
		want string
	}{
		{"push", "constant", 0x0000, asmPushConst(0x0000)},
		{"push", "constant", 0xFFFF, asmPushConst(0xFFFF)},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		if e := cw.WritePushPop(tt.cmd, tt.seg, tt.idx); e != nil {
			t.Fatalf("WritePushPop failed: %s", e.Error())
		}
		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %s", e.Error())
		}

		got := buf.String()
		if got != tt.want {
			t.Errorf("src = %s %s %d\ngot =\n%s\nwant =\n%s", tt.cmd, tt.seg, tt.idx, got, tt.want)
		}

		buf.Reset()
		cw.err = nil
	}
}

func TestPushStack(t *testing.T) {
	testCases := []struct {
		v    uint
		want string
	}{
		{bitFalse, asmPushConst(bitFalse)},
		{1, asmPushConst(1)},
		{2, asmPushConst(2)},
		{bitTrue, asmPushConst(bitTrue)},
	}

	var buf bytes.Buffer
	cw := New(&buf)
	for _, tt := range testCases {
		if cw.pushStack(tt.v); cw.err != nil {
			t.Fatalf("pushStack failed: %v", cw.err)
		}
		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %s", e.Error())
		}

		got := buf.String()
		if got != tt.want {
			t.Errorf("v = %d\ngot =\n%s\nwant =\n%s", tt.v, got, tt.want)
		}

		buf.Reset()
		cw.err = nil
	}
}

func asmPushConst(v uint) string {
	tpl := `@%d
D=A
@SP
A=M
M=D
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, v)
}

func asmUnary(op string) string {
	tpl := `@SP
AM=M-1
M=%sM
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, op)
}

func TestUnary(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"neg", asmUnary("-")},
		{"not", asmUnary("!")},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.unary(tt.cmd)
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("cmd = %s\ngot =\n%s\nwant =\n%s", tt.cmd, got, tt.want)
		}
	}
}

func asmBinary(op string) string {
	tpl := `@SP
AM=M-1
D=M
@SP
AM=M-1
%s
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, op)
}

func TestBinary(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"add", asmBinary("M=D+M")},
		{"sub", asmBinary("M=M-D")},
		{"and", asmBinary("M=D&M")},
		{"or", asmBinary("M=D|M")},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.binary(tt.cmd)
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("cmd = %s\ngot =\n%s\nwant =\n%s", tt.cmd, got, tt.want)
		}
	}
}

func asmCompare(op, labelJmp, labelEnd string) string {
	tpl := `@SP
AM=M-1
D=M
@SP
AM=M-1
D=M-D
@%s
D;%s
@0
D=A
@SP
A=M
M=D
@%s
0;JMP
(%s)
@65535
D=A
@SP
A=M
M=D
(%s)
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, labelJmp, op, labelEnd, labelJmp, labelEnd)
}

func TestCompare(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"eq", asmCompare("JEQ", "LABEL0", "LABEL1")},
		{"gt", asmCompare("JGT", "LABEL0", "LABEL1")},
		{"lt", asmCompare("JLT", "LABEL0", "LABEL1")},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.compare(tt.cmd)
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("cmd = %s\ngot =\n%s\nwant =\n%s", tt.cmd, got, tt.want)
		}
	}
}

func TestAcmd(t *testing.T) {
	testCases := []struct {
		addr string
		want string
	}{
		{"16", "@16\n"},
		{"i", "@i\n"},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.acmd(tt.addr)
		if cw.err != nil {
			t.Fatalf("error writing aCommand: %s", cw.err)
		}
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("got = %s; want = %s", got, tt.want)
		}
	}
}

func TestCcmdj(t *testing.T) {
	testCases := []struct {
		dest, comp, jump string
		want             string
	}{
		{"M", "M+D", "", "M=M+D\n"},
		{"", "D", "JMP", "D;JMP\n"},
		{"AMD", "D|M", "JEQ", "AMD=D|M;JEQ\n"},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.ccmdj(tt.dest, tt.comp, tt.jump)
		if cw.err != nil {
			t.Fatalf("error writing cCommand: %s", cw.err)
		}
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("got = %s; want = %s", got, tt.want)
		}
	}
}

func TestLcmd(t *testing.T) {
	testCases := []struct {
		label string
		want  string
	}{
		{"LABEL", "(LABEL)\n"},
	}

	for _, tt := range testCases {
		var buf bytes.Buffer
		cw := New(&buf)

		cw.lcmd(tt.label)
		if cw.err != nil {
			t.Fatalf("error writing lCommand: %s", cw.err)
		}
		_ = cw.Close()

		got := buf.String()
		if got != tt.want {
			t.Errorf("got = %s; want = %s", got, tt.want)
		}
	}
}
