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

	wg.Add(len(args))
	for _, path := range args {
		go func(path string) {
			defer wg.Done()
			convert(msg, path)
		}(path)
	}

	// clone channel if all goroutines finish
	go func() {
		wg.Wait()
		close(msg)
	}()

	for m := range msg {
		fmt.Fprintln(os.Stdout, m)
	}
}

// convert converts `in` source assembly code to machine code and write it to `out`.
func convert(ch chan string, path string) {
	in, err := os.Open(path)
	if err != nil {
		ch <- err.Error()
		return
	}

	outName := outPath(path, binExt)
	out, err := os.Create(outName)
	if err != nil {
		ch <- err.Error()
		return
	}

	// convert in to out
	if e := asm.New(in).Run(out); e != nil {
		ch <- e.Error()
		return
	}
	ch <- fmt.Sprintf("Successfully converted %s to %s", path, outName)
}

// outPath returns a new output file name path with the given new extension name.
// For example, if path is "/foo/bar/baz.old" and newExt is "new", it returns "/foo/bar/baz.new".
func outPath(path string, newExt string) string {
	oldExt := filepath.Ext(path)
	return path[:len(path)-len(oldExt)] + "." + newExt
}
