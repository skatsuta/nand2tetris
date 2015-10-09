package vmtranslator

import (
	"fmt"
	"io"

	"github.com/skatsuta/nand2tetris/projects/07/vmtranslator/codewriter"
	"github.com/skatsuta/nand2tetris/projects/07/vmtranslator/parser"
)

// VMTranslator is a translator that converts VM code to Hack assembly code.
type VMTranslator struct {
	p  *parser.Parser
	cw *codewriter.CodeWriter
}

// New creates a new VMTranslator that translates srces into one assembly code.
func New(out io.Writer) *VMTranslator {
	return &VMTranslator{
		cw: codewriter.New(out),
	}
}

// Run runs the translation from source VM files tr holds to out.
func (tr *VMTranslator) Run(filename string, src io.Reader) error {
	// write the file name as a comment
	if e := tr.cw.SetFileName(filename); e != nil {
		return fmt.Errorf("cannot write file name: %v", e)
	}

	var (
		err error
		p   = parser.New(src)
	)

	for p.HasMoreCommands() {
		if e := p.Advance(); e != nil {
			return fmt.Errorf("error parsing a command: %v", e)
		}

		switch p.CommandType() {
		case parser.Arithmetic:
			err = tr.cw.WriteArithmetic(p.Arg1())
		case parser.Push:
			err = tr.cw.WritePushPop("push", p.Arg1(), p.Arg2())
		}

		if err != nil {
			return fmt.Errorf("error writing a command: %v", err)
		}
	}

	// flush and close
	return tr.cw.Close()
}
