package asm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/skatsuta/nand2tetris/projects/06/assembler/code"
	"github.com/skatsuta/nand2tetris/projects/06/assembler/parser"
	"github.com/skatsuta/nand2tetris/projects/06/assembler/symbtbl"
)

const (
	bitLen = 16
)

// Asm is an Hack assembler.
type Asm struct {
	err  error
	data []byte
	p    *parser.Parser
	c    *code.Code
	st   *symbtbl.SymbolTable
}

// New creates a new Asm object that converts `in` to a Hack binary code.
func New(in io.Reader) (*Asm, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("asm.New: %s", err.Error())
	}

	a := &Asm{
		data: data,
		p:    parser.NewParser(bytes.NewBuffer(data)),
		c:    &code.Code{},
		st:   symbtbl.NewSymbolTable(),
	}
	return a, nil
}

// DefineSymbols adds pre-defined symbols into the assembler.
func (a *Asm) DefineSymbols(sym map[string]uintptr) {
	a.st.AddEntries(sym)
}

// Run converts a Hack assembly code that `a` holds to a Hack binary code
// and write it into out.
func (a *Asm) Run(out io.Writer) error {
	//=== first loop: only creating a symbol table ===//
	for a.p.HasMoreCommands() {
		if e := a.p.Advance(); e != nil {
			return fmt.Errorf("asm: %s", e.Error())
		}

		// first loop focuses on label commands, so skip the others
		if a.p.CommandType() != parser.LCommand {
			continue
		}

		// add label symbol and next ROM address
		a.st.AddEntry(a.p.Symbol(), a.p.ROMAddr()+1)
	}

	//=== second loop: parsing entire code ===//
	a.p = parser.NewParser(bytes.NewBuffer(a.data))
	for a.p.HasMoreCommands() {
		if e := a.p.Advance(); e != nil {
			return fmt.Errorf("asm: %s", e.Error())
		}

		var (
			b   int
			err error
		)
		switch a.p.CommandType() {
		case parser.LCommand:
			// skip a label command
			continue
		case parser.ACommand:
			symb := a.p.Symbol()
			if b, err = strconv.Atoi(symb); err != nil {
				// add the symbol only if it is not an integer and is not contained yet in symbol table
				if !a.st.Contains(symb) {
					a.st.AddVar(symb)
				}
				// if symbol is not an integer, get its address from symbol table
				b = int(a.st.GetAddress(symb))
			}
		case parser.CCommand:
			if b, err = a.formatCCmd(a.p.Dest(), a.p.Comp(), a.p.Jump()); err != nil {
				return fmt.Errorf("failed to parse command: %s", err.Error())
			}
		}

		if e := a.write(out, b); e != nil {
			return fmt.Errorf("failed to write output: %s", e.Error())
		}
	}
	return nil
}

// formatCCmd formats dest, comp and jump mneumonics into one machine code.
// If the arguments contain an invalid mneumonic, it returns an error.
func (a *Asm) formatCCmd(dest, comp, jump string) (int, error) {
	dbyt, err := a.c.Dest(dest)
	a.setErr(err)
	cbyt, err := a.c.Comp(comp)
	a.setErr(err)
	jbyt, err := a.c.Jump(jump)
	a.setErr(err)

	if a.err != nil {
		return 0, fmt.Errorf("formatCCmd: %s", a.err.Error())
	}

	// C instruction: 111 0000000(comp) 000(dest) 000(jump)
	return int(0x7)<<13 | int(cbyt)<<6 | int(dbyt)<<3 | int(jbyt), nil
}

// setErr sets err only if a.err is nil. This method is used for holding the first error.
func (a *Asm) setErr(err error) {
	if a.err != nil {
		return
	}
	a.err = err
}

// write writes binary format of i into out.
func (a *Asm) write(out io.Writer, i int) error {
	_, e := fmt.Fprintf(out, "%0"+strconv.Itoa(bitLen)+"b\n", i)
	return e

}
