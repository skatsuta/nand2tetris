package parser

import (
	"bufio"
	"io"
	"strings"
)

const (
	// prefixComment is a prefix of a comment.
	prefixComment = "//"
)

// parser is a parser for Hack assembly language.
type parser struct {
	in   *bufio.Scanner
	line string
}

// newParser creates a new parser object that reads and parses r.
func newParser(r io.Reader) *parser {
	return &parser{
		in: bufio.NewScanner(r),
	}
}

// hasMoreCommands reports whether there exist more commands in input.
func (p *parser) hasMoreCommands() bool {
	// if Scan() == true && Text() is a comment, continue (skip)
	// if Scan() == true && Text() is not a comment, return true
	// if Scan() == false, return false
	for p.in.Scan() {
		// trim all leading and trailing white spaces
		p.line = strings.TrimSpace(p.in.Text())

		// return true if the line is not empty and not a comment, that is, a command
		if p.line != "" && !strings.HasPrefix(p.line, prefixComment) {
			return true
		}
	}
	return false
}
