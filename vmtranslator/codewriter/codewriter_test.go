package codewriter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/skatsuta/nand2tetris/vmtranslator/parser"
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

func TestFileNameBase(t *testing.T) {
	testCases := []struct {
		filename string
		want     string
	}{
		{"foo.txt", "foo"},
		{"a/foo.txt", "foo"},
		{"a/b/foo.txt", "foo"},
	}

	var buf bytes.Buffer
	cw := New(&buf)
	for _, tt := range testCases {
		got := cw.fileNameBase(tt.filename)
		if got != tt.want {
			t.Errorf("filename = %s; got %s; want %s", tt.filename, got, tt.want)
		}
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
			diff := diffTexts(got, tt.want)
			t.Errorf("src = %s\n%s", tt.cmd, diff)
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
		cmd  parser.CommandType
		seg  string
		idx  uint
		want string
	}{
		{parser.Push, "constant", 0, asmPushConst(0) + asmEnd},
		{parser.Push, "constant", 1, asmPushConst(1) + asmEnd},
		{parser.Push, "local", 0, asmPushMem("LCL", 0) + asmEnd},
		{parser.Push, "argument", 0, asmPushMem("ARG", 0) + asmEnd},
		{parser.Push, "this", 0, asmPushMem("THIS", 0) + asmEnd},
		{parser.Push, "that", 0, asmPushMem("THAT", 0) + asmEnd},
		{parser.Push, "temp", 0, asmPushReg("R5", 0) + asmEnd},
		{parser.Push, "temp", 7, asmPushReg("R5", 7) + asmEnd},
		{parser.Push, "pointer", 0, asmPushReg("R3", 0) + asmEnd},
		{parser.Push, "pointer", 1, asmPushReg("R3", 1) + asmEnd},
		{parser.Pop, "local", 0, asmPopMem("LCL", 0) + asmEnd},
		{parser.Pop, "argument", 2, asmPopMem("ARG", 2) + asmEnd},
		{parser.Pop, "this", 3, asmPopMem("THIS", 3) + asmEnd},
		{parser.Pop, "that", 4, asmPopMem("THAT", 4) + asmEnd},
		{parser.Pop, "temp", 0, asmPopReg("R5", 0) + asmEnd},
		{parser.Pop, "temp", 7, asmPopReg("R5", 7) + asmEnd},
		{parser.Pop, "pointer", 0, asmPopReg("R3", 0) + asmEnd},
		{parser.Pop, "pointer", 1, asmPopReg("R3", 1) + asmEnd},
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
			diff := diffTexts(got, tt.want)
			t.Errorf("src = \"%d %s %d\"\n%s", tt.cmd, tt.seg, tt.idx, diff)
		}

		buf.Reset()
		cw.err = nil
	}
}

func TestWritePushPopStatic(t *testing.T) {
	testCases := []struct {
		filename string
		cmd      parser.CommandType
		seg      string
		idx      uint
		want     string
	}{
		{"push0.vm", parser.Push, "static", 0, asmPushStatic("push0.vm", "push0", 0) + asmEnd},
		{"push5.vm", parser.Push, "static", 5, asmPushStatic("push5.vm", "push5", 5) + asmEnd},
		{"pop0.vm", parser.Pop, "static", 0, asmPopStatic("pop0.vm", "pop0", 0) + asmEnd},
		{"pop5.vm", parser.Pop, "static", 5, asmPopStatic("pop5.vm", "pop5", 5) + asmEnd},
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

		got := out.String()
		if got != tt.want {
			diff := diffTexts(got, tt.want)
			t.Errorf("src = \"%d %s %d\"\n%s", tt.cmd, tt.seg, tt.idx, diff)
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
			diff := diffTexts(got, tt.want)
			t.Errorf("v = %d\n%s", tt.v, diff)
		}

		buf.Reset()
		cw.err = nil
	}
}

func TestWriteLabelGoto(t *testing.T) {
	testCases := []struct {
		cmd   string
		label string
		want  string
	}{
		{"label", "LABEL", asmLabel("LABEL") + asmEnd},
		{"goto", "LABEL", asmGoto("LABEL") + asmEnd},
		{"if-goto", "LABEL", asmIf("LABEL") + asmEnd},
	}

	var (
		out bytes.Buffer
		cw  *CodeWriter
		err error
	)

	for _, tt := range testCases {
		cw = New(&out)

		switch tt.cmd {
		case "label":
			err = cw.WriteLabel(tt.label)
		case "goto":
			err = cw.WriteGoto(tt.label)
		case "if-goto":
			err = cw.WriteIf(tt.label)
		}
		if err != nil {
			t.Fatalf("WriteLabel failed: %v", err)
		}

		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %v", e)
		}

		got := out.String()
		if got != tt.want {
			diff := diffTexts(got, tt.want)
			t.Errorf("label = %q\n%s", tt.label, diff)
		}

		out.Reset()
	}
}

func TestWriteFunction(t *testing.T) {
	testCases := []struct {
		funcName  string
		numLocals uint
		want      string
	}{
		{"func0", 0, asmFunc("func0", 0) + asmEnd},
		{"func1", 1, asmFunc("func1", 1) + asmEnd},
		{"func2", 2, asmFunc("func2", 2) + asmEnd},
	}

	var (
		out bytes.Buffer
		cw  *CodeWriter
	)
	for _, tt := range testCases {
		cw = New(&out)
		if e := cw.WriteFunction(tt.funcName, tt.numLocals); e != nil {
			t.Fatalf("WriteFunction failed: %v", e)
		}

		if e := cw.Close(); e != nil {
			t.Fatalf("Close failed: %v", e)
		}

		got := out.String()
		if got != tt.want {
			diff := diffTexts(got, tt.want)
			t.Errorf("%s(%d):\n%s", tt.funcName, tt.numLocals, diff)
		}

		out.Reset()
	}
}

func asmFunc(name string, num int) string {
	tpl := "(%s)\n"
	ini := `@%d
D=A
@LCL
AD=D+M
M=0
`
	for i := 0; i < num; i++ {
		tpl += fmt.Sprintf(ini, i)
	}
	return fmt.Sprintf(tpl, name)
}

func asmIf(label string) string {
	tpl := `@SP
AM=M-1
D=M
@%s
D;JNE
`
	return fmt.Sprintf(tpl, label)
}

func asmGoto(label string) string {
	tpl := `@%s
0;JMP
`
	return fmt.Sprintf(tpl, label)
}

func asmLabel(label string) string {
	return fmt.Sprintf("(%s)\n", label)
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

func asmPushStatic(filename, base string, idx uint) string {
	tpl := `// %s
@%s.%d
D=M
@SP
A=M
M=D
@SP
AM=M+1
`
	return fmt.Sprintf(tpl, filename, base, idx)
}

func asmPopStatic(filename, base string, idx uint) string {
	tpl := `// %s
@SP
AM=M-1
D=M
@%s.%d
M=D
`
	return fmt.Sprintf(tpl, filename, base, idx)
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

// diffTexts returns a text representing a difference between text1 and text2.
func diffTexts(text1, text2 string) string {
	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(text1, text2)
	diffs := dmp.DiffMain(a, b, false)
	res := dmp.DiffCharsToLines(diffs, c)

	var buf bytes.Buffer
	for _, diff := range res {
		if diff.Type < 0 {
			_, _ = buf.WriteString("< " + diff.Text)
		} else if diff.Type > 0 {
			_, _ = buf.WriteString("\t\t> " + diff.Text)
		}
	}

	return buf.String()
}
