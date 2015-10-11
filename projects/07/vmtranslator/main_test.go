package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	testCases := []struct {
		path string
		got  string
		want string
	}{
		{"../StackArithmetic/SimpleAdd/SimpleAdd.vm",
			"../StackArithmetic/SimpleAdd/SimpleAdd.asm", simpleAdd},
	}

	for _, tt := range testCases {
		if e := convert(tt.path); e != nil {
			t.Fatal(e)
		}

		gotb, _ := ioutil.ReadFile(tt.got)
		got := strings.Split(string(gotb), "\n")
		want := strings.Split(tt.want, "\n")

		if len(got) != len(want) {
			t.Fatalf("the number of lines: got = %d, want = %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("line %2d:\n got = %q\nwant = %q", i+1, got[i], want[i])
			}
		}
	}
}

func TestOutpath(t *testing.T) {
	testCases := []struct {
		path  string
		isDir bool
		want  string
	}{
		{"foo.vm", false, "foo.asm"},
		{"foo", true, "foo/foo.asm"},
	}

	for _, tt := range testCases {
		got := outpath(tt.path, tt.isDir)
		if got != tt.want {
			t.Errorf("got = %s; want = %s", got, tt.want)
		}
	}
}

var simpleAdd = `// ../StackArithmetic/SimpleAdd/SimpleAdd.vm
@7
D=A
@SP
A=M
M=D
@SP
AM=M+1
@8
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
