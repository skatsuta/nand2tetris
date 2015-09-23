package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/skatsuta/nand2tetris/projects/06/assembler/code"
)

const (
	// prefixComment is a prefix of a comment.
	prefixComment = "//"
)

// CommandType represents a type of a Hack command.
type CommandType int

const (
	// ACommand means @Xxx command.
	ACommand CommandType = iota
	// CCommand means dest=comp;jump command.
	CCommand
	// LCommand means (Xxx) pseudo command.
	LCommand
)

type command struct {
	cmd  string
	typ  CommandType
	symb string
	dest string
	comp string
	jump string
}

// Parser is a parser for Hack assembly language.
// Parser is not thread safe, so it should not be used in multiple goroutines.
type Parser struct {
	in      *bufio.Scanner
	err     error
	line    string
	command command
	romaddr uintptr
}

// NewParser creates a new parser object that reads and parses r.
func NewParser(r io.Reader) *Parser {
	ptr := uintptr(0)
	return &Parser{
		in: bufio.NewScanner(r),
		// initialize to the max value of uintptr
		// in order to set romaddr as 0 in the first increment
		romaddr: ptr - 1,
	}
}

// HasMoreCommands reports whether there exist more commands in input.
func (p *Parser) HasMoreCommands() bool {
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
			p.romaddr++
			return true
		}
	}
	return false
}

// ROMAddr returns current ROM address.
func (p *Parser) ROMAddr() uintptr {
	return p.romaddr
}

// Advance reads next command from input and set the command to current one.
// If the next command is invalid, it returns an error.
// This method should be called only if hasMoreCommands() returns true.
func (p *Parser) Advance() error {
	if p.err != nil {
		return p.err
	}

	// trim a comment and get a pure command string
	cmd := p.trimComment(p.line)
	var typ CommandType
	var symb, dest, comp, jump string

	switch cmd[0] {
	// assginment command
	case '@':
		typ = ACommand
		symb = cmd[1:]
	// lobal command
	case '(':
		lastc := cmd[len(cmd)-1]
		if lastc != ')' {
			p.err = fmt.Errorf("label command should be closed with ')', but got %s", string(lastc))
			return p.err
		}
		typ = LCommand
		symb = cmd[1 : len(cmd)-1]
	// computation command
	default:
		var cod code.Code
		s1 := p.splitCmd(cmd, "=")
		// next parse target command
		next := s1[0]
		if len(s1) == 2 {
			// check whether dest command is valid
			if !cod.IsValidDest(s1[0]) {
				p.err = fmt.Errorf("invalid dest command: %s", s1[0])
				return p.err
			}
			dest = s1[0]
			// replace next parse target command
			next = s1[1]
		}
		// split next parse target command
		s2 := p.splitCmd(next, ";")
		// check whether comp command is valid
		if !cod.IsValidComp(s2[0]) {
			p.err = fmt.Errorf("invalid comp command: \"%s\"", s2[0])
			return p.err
		}
		comp = s2[0]
		if len(s2) == 2 {
			// check whether jump command is valid
			if !cod.IsValidJump(s2[1]) {
				p.err = fmt.Errorf("invalid jump command: %s", s2[1])
				return p.err
			}
			jump = s2[1]
		}
		typ = CCommand
	}

	// assgin into fields if no error occurs
	p.command = command{
		cmd:  cmd,
		typ:  typ,
		symb: symb,
		dest: dest,
		comp: comp,
		jump: jump,
	}

	return nil
}

// CommandType returns a type of a current command.
func (p *Parser) CommandType() CommandType {
	return p.command.typ
}

// Symbol returns a symbol in a current command. This method should be called
// only if CommandType() returns aCommand or lCommand.
func (p *Parser) Symbol() string {
	return p.command.symb
}

// Dest returns a destination in a current command. This method should be called
// only if CommandType() returns cCommand.
func (p *Parser) Dest() string {
	return p.command.dest
}

// Comp returns a comp section in a current command. This method should be called
// only if CommandType() returns cCommand.
func (p *Parser) Comp() string {
	return p.command.comp
}

// Jump returns a jump section in a current command. This method should be called
// only if CommandType() returns cCommand.
func (p *Parser) Jump() string {
	return p.command.jump
}

// trimComment trims off an inline comment. If the line has no comment, it does nothing.
func (p *Parser) trimComment(line string) string {
	idx := strings.Index(line, prefixComment)

	// If the line has no comment, do nothing.
	if idx < 0 {
		return line
	}

	return strings.TrimSpace(line[:idx])
}

// splitCmd splits cmd into up to two elements by sep.
func (p *Parser) splitCmd(cmd string, sep string) []string {
	return strings.SplitN(cmd, sep, 2)
}
