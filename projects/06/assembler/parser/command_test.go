package parser

import (
	"reflect"
	"testing"
)

func TestNewCommand(t *testing.T) {
	newCommandTests := []struct {
		cmd  string
		want command
	}{
		{"@1", aCmd{baseCmd{"@1", aCommand}}},
		{"@a", aCmd{baseCmd{"@a", aCommand}}},
		{"@END", aCmd{baseCmd{"@END", aCommand}}},
		{"D=M", cCmd{baseCmd{"D=M", cCommand}}},
		{"D=1", cCmd{baseCmd{"D=1", cCommand}}},
		{"M,D=A", cCmd{baseCmd{"M,D=A", cCommand}}},
		{"0;JMP", cCmd{baseCmd{"0;JMP", cCommand}}},
		{"D;JEQ", cCmd{baseCmd{"D;JEQ", cCommand}}},
		{"(LOOP)", lCmd{baseCmd{"(LOOP)", lCommand}}},
	}

	for _, tt := range newCommandTests {
		got := newCommand(tt.cmd)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got: %+v; want: %+v", got, tt.want)
		}
	}
}
