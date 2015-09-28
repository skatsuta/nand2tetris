package codewriter

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriteArithmetic(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"add", asmBinary("M=M+D")},
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

func asmUnary(op string) string {
	tpl := `@SP
AM=M-1
M=%sM
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, op)
}

func Test_unary(t *testing.T) {
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

func Test_binary(t *testing.T) {
	testCases := []struct {
		cmd  string
		want string
	}{
		{"add", asmBinary("M=M+D")},
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
@SP
A=M
M=0
@%s
0;JMP
(%s)
@SP
A=M
M=-1
(%s)
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, labelJmp, op, labelEnd, labelJmp, labelEnd)
}

func Test_compare(t *testing.T) {
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

func Test_acmd(t *testing.T) {
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

func Test_ccmdj(t *testing.T) {
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

func Test_lcmd(t *testing.T) {
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
