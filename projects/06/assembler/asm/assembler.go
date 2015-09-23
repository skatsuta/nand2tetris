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
	binExt = ".hack"
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
			// if symbol is an integer
			if b, err = strconv.Atoi(a.p.Symbol()); err != nil {
				// TODO implement @SYMBOL pattern
				continue
			}
		case parser.CCommand:
			b, err = a.formatCInst(a.p.Dest(), a.p.Comp(), a.p.Jump())
		}

		if err != nil {
			return fmt.Errorf("asm: %s", err.Error())
		}

		if e := a.write(out, b); e != nil {
			return fmt.Errorf("failed to write output: %s", e.Error())
		}
	}
	return nil
}

// formatCInst formats dest, comp and jump mneumonics into one machine code.
// If the arguments contain an invalid mneumonic, it returns an error.
func (a *Asm) formatCInst(dest, comp, jump string) (int, error) {
	dbyt, err := a.c.Dest(dest)
	a.setErr(err)
	cbyt, err := a.c.Comp(comp)
	a.setErr(err)
	jbyt, err := a.c.Jump(jump)
	a.setErr(err)

	if a.err != nil {
		return 0, fmt.Errorf("formatBit: %s", a.err.Error())
	}

	// C instruction: 111 comp[0000000] dest[000] jump[000]
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
