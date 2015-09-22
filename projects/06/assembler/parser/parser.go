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

// commandType represents a type of a Hack command.
type commandType int

const (
	// aCommand means @Xxx command.
	aCommand commandType = iota
	// cCommand means dest=comp;jump command.
	cCommand
	// lCommand means (Xxx) pseudo command.
	lCommand
)

// parser is a parser for Hack assembly language.
type parser struct {
	in   *bufio.Scanner
	line string
	cmd  command
	err  error
}

// newParser creates a new parser object that reads and parses r.
func newParser(r io.Reader) *parser {
	return &parser{
		in: bufio.NewScanner(r),
	}
}

// hasMoreCommands reports whether there exist more commands in input.
func (p *parser) hasMoreCommands() bool {
	if p.err != nil {
		return false
	}

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

// advance reads next command from input and set the command to current one.
// If the next command is invalid, it returns an error.
// This method should be called only if hasMoreCommands() returns true.
func (p *parser) advance() error {
	if p.err != nil {
		return p.err
	}

	return nil
}

// trimComment trims off an inline comment. If the line has no comment, it does nothing.
func (p *parser) trimComment(line string) string {
	idx := strings.Index(line, prefixComment)

	// If the line has no comment, do nothing.
	if idx < 0 {
		return line
	}

	return strings.TrimSpace(line[:idx])
}

// splitCmd splits cmd into up to two elements by sep.
func (p *parser) splitCmd(cmd string, sep string) []string {
	return strings.SplitN(cmd, sep, 2)
}
