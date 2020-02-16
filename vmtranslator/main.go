package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skatsuta/nand2tetris/vmtranslator/vmtranslator"
)

const (
	appName = "vmtranslator"
	usage   = "Usage: %s [-h | --help] path"
)

func init() {
	flag.Usage = func() {
		printErr(usage, appName)
	}
}

func main() {
	// Define and parse flags
	var bootstrap, verbose bool
	flag.BoolVar(&bootstrap, "bootstrap", true, "Emit bootstrap code")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode")
	flag.Parse()

	// Check if only one argument is passed
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	path := flag.Arg(0)
	opath, err := compile(path, bootstrap, verbose)
	if err != nil {
		printErr(err.Error())
		os.Exit(255)
	}

	fmt.Println("Successfully compiled", path, "to", opath)
}

// printErr prints an formatted error message in os.Stderr.
func printErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// compile compiles files in path to one .asm file. If bootstrap is true, it also emits
// bootstrap code at the beginning of the output file. If verbose is true, it also emits
// virtual machine instuctions as comments.
func compile(path string, bootstrap, verbose bool) (string, error) {
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
	defer out.Close()

	vmt := vmtranslator.New(out).Verbose(verbose)

	if bootstrap {
		if e := vmt.Init(); e != nil {
			return "", fmt.Errorf("error creating a translator object: %w", e)
		}
	}

	// callback is a callback function called when a file is found.
	// It implements filepath.WalkFunc.
	callback := func(path string, info os.FileInfo, err error) error {
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

		return vmt.Run(path, f)
	}

	// walk throuth path and compile each .vm file
	return opath, filepath.Walk(path, callback)
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
