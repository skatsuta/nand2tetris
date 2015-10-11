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
)

func TestRun(t *testing.T) {
	testCases := []struct {
		filename string
		src      string
		want     string
	}{
		{"foo.vm", "// foo.vm\npush constant 0", "// foo.vm" + wantPushConst0},
		{"bar.vm", "// bar.vm\npush constant 1\npush constant 2\nadd", "// bar.vm" + wantAdd},
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

		got := buf.String()
		if got != tt.want {
			t.Errorf("src = %s\n\ngot =\n%s\nwant =\n%s", tt.src, got, tt.want)
		}

		buf.Reset()
	}
}
