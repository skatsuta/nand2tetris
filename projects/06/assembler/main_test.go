package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	convert("../max/Max.asm")

	gotb, _ := ioutil.ReadFile("../max/Max.hack")
	wantb, _ := ioutil.ReadFile("../max/MaxL.hack")
	got := strings.Split(string(gotb), "\n")
	want := strings.Split(string(wantb), "\n")

	if len(got) != len(want) {
		t.Fatalf("the number of lines should be %d, but got %d", len(want), len(got))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("line %2d: got %s != want %s", i+1, got[i], want[i])
		}
	}
}

func TestFileName(t *testing.T) {
	fileNameTests := []struct {
		path   string
		newExt string
		want   string
	}{
		{"/usr/local/foo.old", "new", "/usr/local/foo.new"},
		{"./bar/foo.old", "new", "./bar/foo.new"},
		{"./foo.old", "new", "./foo.new"},
		{"~/foo.old", "new", "~/foo.new"},
		{"foo.old", "new", "foo.new"},
	}

	for _, tt := range fileNameTests {
		got := outPath(tt.path, tt.newExt)
		if got != tt.want {
			t.Errorf("got = %s, but want = %s", got, tt.want)
		}
	}
}
