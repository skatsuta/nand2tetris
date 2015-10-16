package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/skatsuta/nand2tetris/projects/06/assembler/asm"
)

const (
	// extension name of binary file
	binExt = "hack"
)

// pre-defined symbols
var preDefSymb = map[string]uintptr{
	"SP":     0x0,
	"LCL":    0x1,
	"ARG":    0x2,
	"THIS":   0x3,
	"THAT":   0x4,
	"R0":     0x0,
	"R1":     0x1,
	"R2":     0x2,
	"R3":     0x3,
	"R4":     0x4,
	"R5":     0x5,
	"R6":     0x6,
	"R7":     0x7,
	"R8":     0x8,
	"R9":     0x9,
	"R10":    0xA,
	"R11":    0xB,
	"R12":    0xC,
	"R13":    0xD,
	"R14":    0xE,
	"R15":    0xF,
	"SCREEN": 0x4000,
	"KBD":    0x6000,
}

func main() {
	flag.Parse()
	args := flag.Args()

	msg := make(chan string, len(args))
	var wg sync.WaitGroup

	// convert files concurrently
	wg.Add(len(args))
	for _, path := range args {
		go func(path string) {
			defer wg.Done()
			msg <- convert(path)
		}(path)
	}

	// clone channel after all goroutines finish
	go func() {
		wg.Wait()
		close(msg)
	}()

	// print each message
	for m := range msg {
		fmt.Println(m)
	}
}

// convert converts `in` source assembly code to machine code and write it to `out`.
// It returns a result message if successful, otherwise an error message.
func convert(path string) string {
	// open source file
	in, err := os.Open(path)
	if err != nil {
		return err.Error()
	}
	defer close0(in)

	// create destination file
	outName := outPath(path, binExt)
	out, err := os.Create(outName)
	if err != nil {
		return err.Error()
	}
	defer close0(out)

	// create a new Asm object
	asmblr, err := asm.New(in)
	if err != nil {
		return err.Error()
	}

	// add pre-defined symbols
	asmblr.DefineSymbols(preDefSymb)

	// convert source file to binary file
	if e := asmblr.Run(out); e != nil {
		return e.Error()
	}
	return fmt.Sprintf("Successfully converted %s to %s", path, outName)
}

// outPath returns a new output file name path with the given new extension name.
// For example, if path is "/foo/bar/baz.old" and newExt is "new", it returns "/foo/bar/baz.new".
func outPath(path string, newExt string) string {
	oldExt := filepath.Ext(path)
	return path[:len(path)-len(oldExt)] + "." + newExt
}

// close0 closes cl. If an error occurs, it writes it out to os.Stderr.
func close0(cl io.Closer) {
	if e := cl.Close(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
	}
}
