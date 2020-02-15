package vmtranslator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/skatsuta/nand2tetris/vmtranslator/codewriter"
	"github.com/skatsuta/nand2tetris/vmtranslator/parser"
)

// VMTranslator is a translator that converts VM code to Hack assembly code.
type VMTranslator struct {
	cw      *codewriter.CodeWriter
	verbose bool
}

// New creates a new VMTranslator that translates virtual machine code into assembly code.
func New(out io.Writer) *VMTranslator {
	return &VMTranslator{cw: codewriter.New(out)}
}

// Verbose controls verbose mode of tr. If verbose mode is enabled, the VMTranslator also
// outputs each virtual machine code as a comment corresponding to each assembly code block.
func (tr *VMTranslator) Verbose(verbose bool) *VMTranslator {
	tr.verbose = verbose
	tr.cw.Verbose(verbose)
	return tr
}

// Init initializes the output assembly file.
// This method should be called immediately after New().
func (tr *VMTranslator) Init() error {
	return tr.cw.WriteInit()
}

// run runs the translation from source VM files tr holds to out.
func (tr *VMTranslator) run(filename string, src io.Reader) (err error) {
	// write the file name as a comment
	if e := tr.cw.SetFileName(filename); e != nil {
		return fmt.Errorf("cannot write file name: %v", e)
	}

	p := parser.New(src)
	for p.HasMoreCommands() {
		if e := p.Advance(); e != nil {
			return fmt.Errorf("error parsing a command: %v", e)
		}

		// Current VM instruction
		cmd := p.Command()

		// Write the current VM instruction as a comment for debugging
		if tr.verbose {
			if e := tr.cw.WriteComment(cmd.String()); e != nil {
				return fmt.Errorf("error writing a comment: %v", e)
			}
		}

		switch cmd.Type {
		case parser.Arithmetic:
			err = tr.cw.WriteArithmetic(cmd.Arg1)
		case parser.Push, parser.Pop:
			err = tr.cw.WritePushPop(cmd.Type, cmd.Arg1, cmd.Arg2)
		case parser.Label:
			err = tr.cw.WriteLabel(cmd.Arg1)
		case parser.Goto:
			err = tr.cw.WriteGoto(cmd.Arg1)
		case parser.If:
			err = tr.cw.WriteIf(cmd.Arg1)
		case parser.Function:
			err = tr.cw.WriteFunction(cmd.Arg1, cmd.Arg2)
		case parser.Return:
			err = tr.cw.WriteReturn()
		case parser.Call:
			err = tr.cw.WriteCall(cmd.Arg1, cmd.Arg2)
		default:
			err = fmt.Errorf("unknown command: %s %s %d", cmd.Type, cmd.Arg1, cmd.Arg2)
		}

		if err != nil {
			return fmt.Errorf("error writing a command: %v", err)
		}
	}

	return nil
}

// Run is a callback function called when a file is found.
// It implements filepath.WalkFunc.
func (tr *VMTranslator) Run(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// skip if path is a directory or not a ".vm" file
	if info.IsDir() || filepath.Ext(path) != ".vm" {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	return tr.run(path, f)
}

// Close flush the bufferred output into the output file and closes it.
func (tr *VMTranslator) Close() error {
	return tr.cw.Close()
}
