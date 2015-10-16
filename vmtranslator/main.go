package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skatsuta/nand2tetris/vmtranslator/vmtranslator"
)

var (
	appName = "vmtranslator"
	usage   = "Usage: %s [-h | --help] path"
)

func init() {
	flag.Usage = func() {
		printErr(usage, appName)
	}
}

func main() {
	flag.Parse()

	// check whether one argument is passed
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return
	}

	path := args[0]

	if e := convert(path); e != nil {
		printErr("%v", e)
		return
	}
}

// printErr prints an formatted error message in os.Stderr.
func printErr(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// convert converts files in path to one .asm file.
func convert(path string) error {
	// check whether the given path is valid
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("invalid path is given: %s", path)
	}

	// prepare an output .asm file
	opath := outpath(path, info.IsDir())
	out, err := os.Create(opath)
	if err != nil {
		return fmt.Errorf("cannot create %s", opath)
	}

	vmt := vmtranslator.New(out)
	defer func() {
		_ = vmt.Close()
	}()

	// walk throuth path and run conversion
	if e := filepath.Walk(path, vmt.Run); e != nil {
		return fmt.Errorf("failed to convert: %v", e)
	}

	return nil
}

// outpath returns an output file path.
// This function expects the suffix of the path to be ".vm" if it is a file.
func outpath(path string, isDir bool) string {
	if !isDir {
		// file.vm => file.asm
		return path[:len(path)-2] + "asm"
	}

	filename := filepath.Base(path)
	return filepath.Join(path, filename+".asm")
}
