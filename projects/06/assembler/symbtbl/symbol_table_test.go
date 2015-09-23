package symbtbl

import (
	"reflect"
	"testing"
)

func TestAddEntry(t *testing.T) {
	addEntryTests := []struct {
		symb string
		addr uintptr
	}{
		{"foo", 0x1},
		{"bar", 0x123},
		{"LOOP", 0xAB},
	}

	st := NewSymbolTable()
	for _, tt := range addEntryTests {
		st.AddEntry(tt.symb, tt.addr)

		if !st.Contains(tt.symb) {
			t.Errorf("key %q should be contained", tt.symb)
		}
	}
}

func TestAddEntries(t *testing.T) {
	tbl := map[string]uintptr{
		"foo":  0x1,
		"bar":  0x123,
		"LOOP": 0xAB,
	}

	st := NewSymbolTable()
	st.AddEntries(tbl)

	if !reflect.DeepEqual(st.m, tbl) {
		t.Errorf("got: %v; want: %v", st.m, tbl)
	}
}

func TestGetAddress(t *testing.T) {
	tbl := map[string]uintptr{
		"foo":  0x1,
		"bar":  0x123,
		"LOOP": 0xAB,
	}

	getAddressTests := []struct {
		symb string
		want uintptr
	}{
		{"foo", 0x1},
		{"bar", 0x123},
		{"LOOP", 0xAB},
		{"", 0x0},
		{"no", 0x0},
	}

	st := NewSymbolTable()
	st.AddEntries(tbl)

	for _, tt := range getAddressTests {
		got := st.GetAddress(tt.symb)
		if got != tt.want {
			t.Errorf("the value of %s should be %d, but got %d", tt.symb, tt.want, got)
		}
	}
}
