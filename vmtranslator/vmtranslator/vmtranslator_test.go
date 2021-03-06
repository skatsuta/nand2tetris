package vmtranslator

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	vmtransl := New(&bytes.Buffer{})

	if vmtransl.cw == nil {
		t.Errorf("VMTranslator.cw is nil")
	}
}

var (
	wantPushConst0 = `
@0
D=A
@SP
A=M
M=D
@SP
AM=M+1
`

	wantAdd = `
@1
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

	wantEq = `
@1
D=A
@SP
A=M
M=D
@SP
AM=M+1
@1
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
D=M-D
@LABEL0
D;JEQ
@0
D=A
@SP
A=M
M=D
@LABEL1
0;JMP
(LABEL0)
@SP
A=M
M=-1
(LABEL1)
@SP
AM=M+1
`

	wantPushPop = `
@0
D=A
@SP
A=M
M=D
@SP
AM=M+1
@0
D=A
@LCL
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

	wantLabelIfGoto = `
(LABEL0)
@LABEL1
0;JMP
@SP
AM=M-1
D=M
@LABEL2
D;JNE
`

	end = `(END)
@END
0;JMP
`
)

func TestRun(t *testing.T) {
	testCases := []struct {
		filename string
		src      string
		want     string
	}{
		{"push_const_0.vm", "// push_const_0.vm\npush constant 0", "// push_const_0.vm" + wantPushConst0 + end},
		{"add.vm", "// add.vm\npush constant 1\npush constant 2\nadd", "// add.vm" + wantAdd + end},
		{"eq.vm", "// eq.vm\npush constant 1\npush constant 1\neq", "// eq.vm" + wantEq + end},
		{"push_pop.vm", "// push_pop.vm\npush constant 0\npop local 0", "// push_pop.vm" + wantPushPop + end},
		{"label_if_goto.vm", "// label_if_goto.vm\nlabel LABEL0\ngoto LABEL1\nif-goto LABEL2",
			"// label_if_goto.vm" + wantLabelIfGoto + end},
	}

	var (
		buf      bytes.Buffer
		vmtransl *VMTranslator
	)
	for _, tt := range testCases {
		vmtransl = New(&buf)
		if e := vmtransl.run(tt.filename, strings.NewReader(tt.src)); e != nil {
			t.Fatalf("Run failed: %v", e)
		}
		if e := vmtransl.Close(); e != nil {
			t.Fatalf("Close failed: %v", e)
		}

		got := strings.Split(buf.String(), "\n")
		want := strings.Split(tt.want, "\n")
		if len(got) != len(want) {
			t.Fatalf("%s: the number of lines should be %d, but got %d", tt.filename, len(want), len(got))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("in %s\ngot = %q; want = %q", tt.filename, got[i], want[i])
			}
		}

		buf.Reset()
	}
}

func TestRunErr(t *testing.T) {
	testCases := []struct {
		filename string
		src      string
	}{
		{"4cmd.vm", "// 4cmd.vm\npop local -1"}, // split into [pop, local, -, 1]
		{"unknown_command.vm", "// unknown_command.vm\nfoo"},
		{"unknown_segment.vm", "// unknown_segment.vm\npush foo 1"},
		{"not_integer.vm", "// not_integer.vm\npop local a"},
	}

	var (
		buf      bytes.Buffer
		vmtransl *VMTranslator
	)
	for _, tt := range testCases {
		vmtransl = New(&buf)
		if e := vmtransl.run(tt.filename, strings.NewReader(tt.src)); e == nil {
			t.Errorf("filename = %s\nsrc = %q\nerror should occur but got <nil>", tt.filename, tt.src)
		}

		buf.Reset()
	}
}
