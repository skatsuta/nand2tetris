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
		os.Exit(1)
	}

	path := args[0]
	opath, err := convert(path)
	if err != nil {
		printErr(err.Error())
		os.Exit(255)
	}

	fmt.Println("Successfully converted", path, "to", opath)
}

// printErr prints an formatted error message in os.Stderr.
func printErr(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// convert converts files in path to one .asm file.
func convert(path string) (string, error) {
	// check whether the given path is valid
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("invalid path is given: %s", path)
	}

	// prepare an output .asm file
	opath := outpath(path, info.IsDir())
	out, err := os.Create(opath)
	if err != nil {
		return "", fmt.Errorf("cannot create %s", opath)
	}

	vmt := vmtranslator.New(out)
	if e := vmt.Init(); e != nil {
		return "", fmt.Errorf("error creating a translator object: %v", e)
	}
	defer func() {
		_ = vmt.Close()
	}()

	// walk throuth path and run conversion
	return opath, filepath.Walk(path, vmt.Run)
}

// outpath returns an output file path.
// This function expects the suffix of the path to be ".vm" if it is a file.
func outpath(path string, isDir bool) string {
	if isDir {
		filename := filepath.Base(path)
		return filepath.Join(path, filename+".asm")
	}

	// file.vm => file.asm
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + ".asm"
}
