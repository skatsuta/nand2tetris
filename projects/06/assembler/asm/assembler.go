package asm

import (
	"fmt"
	"io"
	"strconv"

	"github.com/skatsuta/nand2tetris/projects/06/assembler/code"
	"github.com/skatsuta/nand2tetris/projects/06/assembler/parser"
)

const (
	bitLen = 16
	binExt = ".hack"
)

// Asm is an Hack assembler.
type Asm struct {
	p *parser.Parser
	c *code.Code
}

// New creates a new Asm object that converts `in` to a Hack binary code.
func New(in io.Reader) *Asm {
	return &Asm{
		p: parser.NewParser(in),
		c: &code.Code{},
	}
}

// Run converts a Hack assembly code that `a` holds to a Hack binary code
// and write it into out.
func (a *Asm) Run(out io.Writer) error {
	for a.p.HasMoreCommands() {
		if e := a.p.Advance(); e != nil {
			return fmt.Errorf("error occurred while parsing: %s", e.Error())
		}

		switch a.p.CommandType() {
		case parser.ACommand:
			// if symbol is an integer
			if i, e := strconv.Atoi(a.p.Symbol()); e == nil {
				fmt.Fprintf(out, "%0"+strconv.Itoa(bitLen)+"b", i)
			}
		}
	}
	return nil
}
