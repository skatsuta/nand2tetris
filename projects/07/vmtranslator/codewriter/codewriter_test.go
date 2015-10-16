package codewriter

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestSetFileName(t *testing.T) {
	testCases := []struct {
		filename string
		want     string
	}{
		{"", "// \n" + asmEnd},
		{"foo.txt", "// foo.txt\n" + asmEnd},
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
		{"add", asmBinary("M=D+M") + asmEnd},
		{"sub", asmBinary("M=M-D") + asmEnd},
		{"and", asmBinary("M=D&M") + asmEnd},
		{"or", asmBinary("M=D|M") + asmEnd},
		{"neg", asmUnary("-") + asmEnd},
		{"not", asmUnary("!") + asmEnd},
		{"eq", asmCompare("JEQ", "LABEL0", "LABEL1") + asmEnd},
		{"gt", asmCompare("JGT", "LABEL0", "LABEL1") + asmEnd},
		{"lt", asmCompare("JLT", "LABEL0", "LABEL1") + asmEnd},
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
		{"push", "constant", 0, asmPushConst(0) + asmEnd},
		{"push", "constant", 1, asmPushConst(1) + asmEnd},
		{"push", "local", 0, asmPushMem("LCL", 0) + asmEnd},
		{"push", "argument", 0, asmPushMem("ARG", 0) + asmEnd},
		{"push", "this", 0, asmPushMem("THIS", 0) + asmEnd},
		{"push", "that", 0, asmPushMem("THAT", 0) + asmEnd},
		{"push", "temp", 0, asmPushReg("R5", 0) + asmEnd},
		{"push", "temp", 7, asmPushReg("R5", 7) + asmEnd},
		{"push", "pointer", 0, asmPushReg("R3", 0) + asmEnd},
		{"push", "pointer", 1, asmPushReg("R3", 1) + asmEnd},
		{"pop", "local", 0, asmPopMem("LCL", 0) + asmEnd},
		{"pop", "argument", 2, asmPopMem("ARG", 2) + asmEnd},
		{"pop", "this", 3, asmPopMem("THIS", 3) + asmEnd},
		{"pop", "that", 4, asmPopMem("THAT", 4) + asmEnd},
		{"pop", "temp", 0, asmPopReg("R5", 0) + asmEnd},
		{"pop", "temp", 7, asmPopReg("R5", 7) + asmEnd},
		{"pop", "pointer", 0, asmPopReg("R3", 0) + asmEnd},
		{"pop", "pointer", 1, asmPopReg("R3", 1) + asmEnd},
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

		got := strings.Split(buf.String(), "\n")
		want := strings.Split(tt.want, "\n")
		if !reflect.DeepEqual(got, want) {
			var b bytes.Buffer
			_, _ = b.WriteString("<got>\t\t\t<want>\n-------\t\t\t-------\n")
			minlen := int(math.Min(float64(len(got)), float64(len(want))))
			for i := 0; i < minlen; i++ {
				line := fmt.Sprintf("%s\t\t\t%s\n", got[i], want[i])
				_, _ = b.WriteString(line)
			}

			t.Errorf("src = \"%s %s %d\"\n%s", tt.cmd, tt.seg, tt.idx, b.String())
		}

		buf.Reset()
		cw.err = nil
	}
}

func TestWritePushPopStatic(t *testing.T) {
	testCases := []struct {
		filename string
		cmd      string
		seg      string
		idx      uint
		want     string
	}{
		{"push0.vm", "push", "static", 0, asmPushStatic("push0.vm", 0) + asmEnd},
		//{"push", "constant", 1, asmPushConst(1) + asmEnd},
		//{"push", "local", 0, asmPushMem("LCL", 0) + asmEnd},
		//{"push", "argument", 0, asmPushMem("ARG", 0) + asmEnd},
		//{"push", "this", 0, asmPushMem("THIS", 0) + asmEnd},
		//{"push", "that", 0, asmPushMem("THAT", 0) + asmEnd},
		//{"push", "temp", 0, asmPushReg("R5", 0) + asmEnd},
		//{"push", "temp", 7, asmPushReg("R5", 7) + asmEnd},
		//{"push", "pointer", 0, asmPushReg("R3", 0) + asmEnd},
		//{"push", "pointer", 1, asmPushReg("R3", 1) + asmEnd},
		//{"pop", "local", 0, asmPopMem("LCL", 0) + asmEnd},
		//{"pop", "argument", 2, asmPopMem("ARG", 2) + asmEnd},
		//{"pop", "this", 3, asmPopMem("THIS", 3) + asmEnd},
		//{"pop", "that", 4, asmPopMem("THAT", 4) + asmEnd},
		//{"pop", "temp", 0, asmPopReg("R5", 0) + asmEnd},
		//{"pop", "temp", 7, asmPopReg("R5", 7) + asmEnd},
		//{"pop", "pointer", 0, asmPopReg("R3", 0) + asmEnd},
		//{"pop", "pointer", 1, asmPopReg("R3", 1) + asmEnd},
	}

	for _, tt := range testCases {
		var out bytes.Buffer
		cw := New(&out)
		if e := cw.SetFileName(tt.filename); e != nil {
			t.Fatalf("SetFileName failed: %v", e)
		}

		if e := cw.WritePushPop(tt.cmd, tt.seg, tt.idx); e != nil {
			t.Fatalf("WritePushPop failed: %s", e.Error())
		}
		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %s", e.Error())
		}

		got := strings.Split(out.String(), "\n")
		want := strings.Split(tt.want, "\n")
		if !reflect.DeepEqual(got, want) {
			var buf bytes.Buffer
			_, _ = buf.WriteString("<< got >>\t\t\t<< want >>\n-------\t\t\t-------\n")
			minlen := int(math.Min(float64(len(got)), float64(len(want))))
			for i := 0; i < minlen; i++ {
				line := fmt.Sprintf("%s\t\t\t%s\n", got[i], want[i])
				_, _ = buf.WriteString(line)
			}

			t.Errorf("src = \"%s %s %d\"\n%s", tt.cmd, tt.seg, tt.idx, buf.String())
		}

		out.Reset()
		cw.err = nil
	}
}

func TestPushVal(t *testing.T) {
	testCases := []struct {
		v    uint
		want string
	}{
		{bitFalse, asmPushConst(bitFalse) + asmEnd},
		{1, asmPushConst(1) + asmEnd},
		{2, asmPushConst(2) + asmEnd},
	}

	var buf bytes.Buffer
	cw := New(&buf)
	for _, tt := range testCases {
		if cw.pushVal(tt.v); cw.err != nil {
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

func asmPushMem(symb string, idx uint) string {
	tpl := `@%d
D=A
@%s
AD=D+M
D=M
@SP
A=M
M=D
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, idx, symb)
}

func asmPopMem(symb string, idx uint) string {
	tpl := `@%d
D=A
@%s
AD=D+M
@R13
M=D
@SP
AM=M-1
D=M
@R13
A=M
M=D
`
	return fmt.Sprintf(tpl, idx, symb)
}

func asmPushReg(symb string, idx uint) string {
	tpl := `@%d
D=A
@%s
AD=D+A
D=M
@SP
A=M
M=D
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, idx, symb)
}

func asmPopReg(symb string, idx uint) string {
	tpl := `@%d
D=A
@%s
AD=D+A
@R13
M=D
@SP
AM=M-1
D=M
@R13
A=M
M=D
`
	return fmt.Sprintf(tpl, idx, symb)
}

func asmPushStatic(filename string, idx uint) string {
	tpl := `// %s
@%s.%d
D=M
@SP
A=M
M=D
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, filename, filename, idx)
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
		{"neg", asmUnary("-") + asmEnd},
		{"not", asmUnary("!") + asmEnd},
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
		{"add", asmBinary("M=D+M") + asmEnd},
		{"sub", asmBinary("M=M-D") + asmEnd},
		{"and", asmBinary("M=D&M") + asmEnd},
		{"or", asmBinary("M=D|M") + asmEnd},
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
@SP
A=M
M=-1
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
		{"eq", asmCompare("JEQ", "LABEL0", "LABEL1") + asmEnd},
		{"gt", asmCompare("JGT", "LABEL0", "LABEL1") + asmEnd},
		{"lt", asmCompare("JLT", "LABEL0", "LABEL1") + asmEnd},
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
		{"16", "@16\n" + asmEnd},
		{"i", "@i\n" + asmEnd},
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
		{"M", "M+D", "", "M=M+D\n" + asmEnd},
		{"", "D", "JMP", "D;JMP\n" + asmEnd},
		{"AMD", "D|M", "JEQ", "AMD=D|M;JEQ\n" + asmEnd},
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
		{"LABEL", "(LABEL)\n" + asmEnd},
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

var asmEnd = `(END)
@END
0;JMP
`
