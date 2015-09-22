package main

import "testing"

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
