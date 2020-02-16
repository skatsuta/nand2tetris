package vmtranslator

import (
	"strings"
	"testing"
)

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
@LABEL_0
D;JEQ
@0
D=A
@SP
A=M
M=D
@LABEL_1
0;JMP
(LABEL_0)
@SP
A=M
M=-1
(LABEL_1)
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
(LABEL_0)
@LABEL_1
0;JMP
@SP
AM=M-1
D=M
@LABEL_2
D;JNE
`

	wantFunction = `
(Class.method)
@0
D=A
@SP
A=M
M=D
@SP
AM=M+1
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
D=M
@SP
A=M
M=D
@SP
AM=M+1
@1
D=A
@LCL
AD=D+M
D=M
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
@0
D=A
@LCL
AD=D+M
@R14
M=D
@5
D=A
@R14
AD=M-D
D=M
@R15
M=D
@SP
AM=M-1
D=M
@ARG
A=M
M=D
@1
D=A
@ARG
AD=D+M
@SP
M=D
@1
D=A
@R14
AD=M-D
D=M
@THAT
M=D
@2
D=A
@R14
AD=M-D
D=M
@THIS
M=D
@3
D=A
@R14
AD=M-D
D=M
@ARG
M=D
@4
D=A
@R14
AD=M-D
D=M
@LCL
M=D
@R15
A=M
0;JMP
`

	wantCall = `
(Sys.init)
@Class.method_RET_ADDR_0
D=A
@SP
A=M
M=D
@SP
AM=M+1
@LCL
D=M
@SP
A=M
M=D
@SP
AM=M+1
@ARG
D=M
@SP
A=M
M=D
@SP
AM=M+1
@THIS
D=M
@SP
A=M
M=D
@SP
AM=M+1
@THAT
D=M
@SP
A=M
M=D
@SP
AM=M+1
@6
D=A
@SP
AD=M-D
@ARG
M=D
@0
D=A
@SP
AD=D+M
@LCL
M=D
@Class.method
0;JMP
(Class.method_RET_ADDR_0)
`

	wantInit = `@256
D=A
@SP
M=D
@Sys.init
0;JMP
`

	end = `(END)
@END
0;JMP
`
)

func TestInit(t *testing.T) {
	var buf strings.Builder
	vmt := New(&buf)

	if e := vmt.Init(); e != nil {
		t.Fatalf("Init failed: %v", e)
	}
	if e := vmt.cw.Close(); e != nil {
		t.Fatalf("CodeWriter#Close failed: %v", e)
	}

	got := strings.Split(buf.String(), "\n")
	want := strings.Split(wantInit+end, "\n")
	if len(got) != len(want) {
		t.Errorf("the number of lines should be %d, but got %d", len(want), len(got))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got = %q; want = %q", got[i], want[i])
		}
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		filename string
		src      string
		want     string
	}{
		{
			filename: "push_const_0.vm",
			src:      "// push_const_0.vm\npush constant 0",
			want:     "// push_const_0.vm" + wantPushConst0 + end,
		},
		{
			filename: "add.vm",
			src:      "// add.vm\npush constant 1\npush constant 2\nadd",
			want:     "// add.vm" + wantAdd + end,
		},
		{
			filename: "eq.vm",
			src:      "// eq.vm\npush constant 1\npush constant 1\neq",
			want:     "// eq.vm" + wantEq + end,
		},
		{
			filename: "push_pop.vm",
			src:      "// push_pop.vm\npush constant 0\npop local 0",
			want:     "// push_pop.vm" + wantPushPop + end,
		},
		{
			filename: "label_if_goto.vm",
			src:      "// label_if_goto.vm\nlabel LABEL_0\ngoto LABEL_1\nif-goto LABEL_2",
			want:     "// label_if_goto.vm" + wantLabelIfGoto + end,
		},
		{
			filename: "function.vm",
			src:      "// function.vm\nfunction Class.method 2\npush local 0\npush local 1\nadd\nreturn",
			want:     "// function.vm" + wantFunction + end,
		},
		{
			filename: "call.vm",
			src:      "// call.vm\nfunction Sys.init 0\ncall Class.method 1",
			want:     "// call.vm" + wantCall + end,
		},
	}

	var (
		buf strings.Builder
		vmt *VMTranslator
	)
	for _, tt := range testCases {
		vmt = New(&buf)
		if e := vmt.Run(tt.filename, strings.NewReader(tt.src)); e != nil {
			t.Fatalf("Run failed: %v", e)
		}

		got := strings.Split(buf.String(), "\n")
		want := strings.Split(tt.want, "\n")
		if len(got) != len(want) {
			t.Errorf(
				"%s: the number of lines should be %d, but got %d", tt.filename, len(want), len(got),
			)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("in %s\n%d: got = %q; want = %q", tt.filename, i, got[i], want[i])
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
		{"invalid_function.vm", "// invalid_function.vm\nfunction"},
	}

	var (
		buf strings.Builder
		vmt *VMTranslator
	)
	for _, tt := range testCases {
		vmt = New(&buf)
		if e := vmt.Run(tt.filename, strings.NewReader(tt.src)); e == nil {
			t.Errorf("filename = %s\nsrc = %q\nerror should occur but got <nil>", tt.filename, tt.src)
		}

		buf.Reset()
	}
}
