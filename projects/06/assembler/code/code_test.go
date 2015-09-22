package code

import "testing"

func TestDest(t *testing.T) {
	destTests := []struct {
		mneum string
		want  byte
	}{
		{"", 0x0},
		{"M", 0x1},
		{"D", 0x2},
		{"MD", 0x3},
		{"A", 0x4},
		{"AM", 0x5},
		{"AD", 0x6},
		{"AMD", 0x7},
	}

	var c Code
	for _, tt := range destTests {
		got, err := c.dest(tt.mneum)
		if err != nil {
			t.Fatalf("dest error: %s", err.Error())
		}

		if got != tt.want {
			t.Errorf("got = %X; want = %X", got, tt.want)
		}
	}
}

func TestDestError(t *testing.T) {
	destTests := []struct {
		mneum string
	}{
		{"B"},
		{"2"},
	}

	var c Code
	for _, tt := range destTests {
		found := c.IsValidDest(tt.mneum)
		_, err := c.dest(tt.mneum)
		if found || err == nil {
			t.Errorf("%s should not be a valid dest mneumonic", tt.mneum)
		}
	}
}

func TestComp(t *testing.T) {
	compTests := []struct {
		mneum string
		want  byte
	}{
		{"0", 0x2A},
		{"1", 0x3F},
		{"-1", 0x3A},
		{"D", 0xC},
		{"A", 0x30},
		{"M", 0x70},
		{"!D", 0xD},
		{"!A", 0x31},
		{"!M", 0x71},
		{"-D", 0xF},
		{"-A", 0x33},
		{"-M", 0x73},
		{"D+1", 0x1F},
		{"A+1", 0x37},
		{"M+1", 0x77},
		{"D-1", 0xE},
		{"A-1", 0x32},
		{"M-1", 0x72},
		{"D+A", 0x2},
		{"D-A", 0x13},
		{"D+M", 0x42},
		{"D-M", 0x53},
		{"A-D", 0x7},
		{"M-D", 0x47},
		{"D&A", 0x0},
		{"D|A", 0x15},
		{"D&M", 0x40},
		{"D|M", 0x55},
	}

	var c Code
	for _, tt := range compTests {
		got, err := c.comp(tt.mneum)
		if err != nil {
			t.Fatalf("comp error: %s", err.Error())
		}

		if got != tt.want {
			t.Errorf("got = %X; want = %X", got, tt.want)
		}
	}
}

func TestCompError(t *testing.T) {
	compTests := []struct {
		mneum string
	}{
		{""},
		{"2"},
		{"B"},
		{"0;JMP"},
		{"M=D"},
	}

	var c Code
	for _, tt := range compTests {
		found := c.IsValidComp(tt.mneum)
		_, err := c.comp(tt.mneum)
		if found || err == nil {
			t.Errorf("%s should not be a valid comp mneumonic", tt.mneum)
		}
	}
}

func TestJump(t *testing.T) {
	jumpTests := []struct {
		mneum string
		want  byte
	}{
		{"", 0x0},
		{"JGT", 0x1},
		{"JEQ", 0x2},
		{"JGE", 0x3},
		{"JLT", 0x4},
		{"JNE", 0x5},
		{"JLE", 0x6},
		{"JMP", 0x7},
	}

	var c Code
	for _, tt := range jumpTests {
		got, err := c.jump(tt.mneum)
		if err != nil {
			t.Fatalf("jump error: %s", err.Error())
		}

		if got != tt.want {
			t.Errorf("got = %X; want = %X", got, tt.want)
		}
	}
}

func TestJumpError(t *testing.T) {
	jumpTests := []struct {
		mneum string
	}{
		{"JJ"},
		{"0"},
	}

	var c Code
	for _, tt := range jumpTests {
		found := c.IsValidJump(tt.mneum)
		_, err := c.jump(tt.mneum)
		if found || err == nil {
			t.Errorf("%s should not be a valid jump mneumonic", tt.mneum)
		}
	}
}
