package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/skatsuta/nand2tetris/projects/06/assembler/asm"
)

const (
	binExt = "hack"
)

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
	in, err := os.Open(path)
	if err != nil {
		return err.Error()
	}

	outName := outPath(path, binExt)
	out, err := os.Create(outName)
	if err != nil {
		return err.Error()
	}

	// convert in to out
	asmblr, err := asm.New(in)
	if err != nil {
		return err.Error()
	}
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
