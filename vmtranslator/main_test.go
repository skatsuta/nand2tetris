package main

import "testing"

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
