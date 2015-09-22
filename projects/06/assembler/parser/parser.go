package parser

import (
	"bufio"
	"fmt"
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

type command struct {
	cmd  string
	typ  commandType
	symb string
	dest string
	comp string
	jump string
}

// parser is a parser for Hack assembly language.
type parser struct {
	in      *bufio.Scanner
	err     error
	line    string
	command command
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

	// trim a comment and get a pure command string
	cmd := p.trimComment(p.line)
	var typ commandType
	var symb, dest, comp, jump string

	switch cmd[0] {
	// assginment command
	case '@':
		typ = aCommand
		symb = cmd[1:]
	// lobal command
	case '(':
		lastc := cmd[len(cmd)-1]
		if lastc != ')' {
			p.err = fmt.Errorf("label command should be closed with ')', but got %s", string(lastc))
			return p.err
		}
		typ = lCommand
		symb = cmd[1 : len(cmd)-1]
	// computation command
	default:
		s1 := p.splitCmd(cmd, "=")
		// next parse target command
		next := s1[0]
		if len(s1) == 2 {
			// check whether dest command is valid
			if !destBit.contains(s1[0]) {
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
		if !compBit0.contains(s2[0]) && !compBit1.contains(s2[0]) {
			p.err = fmt.Errorf("invalid comp command: \"%s\"", s2[0])
			return p.err
		}
		comp = s2[0]
		if len(s2) == 2 {
			// check whether jump command is valid
			if !jumpBit.contains(s2[1]) {
				p.err = fmt.Errorf("invalid jump command: %s", s2[1])
				return p.err
			}
			jump = s2[1]
		}
		typ = cCommand
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

// commandType returns a type of a current command.
func (p *parser) commandType() commandType {
	return p.command.typ
}

// symbol returns a symbol of a current command. This method should be called
// only if commandType() returns aCommand or lCommand.
func (p *parser) symbol() string {
	return p.command.symb
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

// opcBit is a map of an opcode and a binary instruction.
type opcBit map[string]byte

// contains reports whether opc is contained in ob.
func (ob opcBit) contains(opc string) bool {
	_, found := ob[opc]
	return found
}

var (
	// destBit is a map of a dest command and a binary instruction.
	destBit opcBit = map[string]byte{
		"":    0x0,
		"M":   0x1,
		"D":   0x2,
		"MD":  0x3,
		"A":   0x4,
		"AM":  0x5,
		"AD":  0x6,
		"AMD": 0x7,
	}

	// compBit0 is a map of a comp command and a binary instruction in the case a = 0
	compBit0 opcBit = map[string]byte{
		"0":   0x2A,
		"1":   0x3F,
		"-1":  0x3A,
		"D":   0xC,
		"A":   0x30,
		"!D":  0xD,
		"!A":  0x31,
		"-D":  0xF,
		"-A":  0x33,
		"D+1": 0x1F,
		"A+1": 0x37,
		"D-1": 0xE,
		"A-1": 0x32,
		"D+A": 0x2,
		"D-A": 0x13,
		"A-D": 0x7,
		"D&A": 0x0,
		"D|A": 0x15,
	}

	// compBit1 is a map of a comp command and a binary instruction in the case a = 1
	compBit1 opcBit = map[string]byte{
		"M":   0x30,
		"!M":  0x31,
		"-M":  0x33,
		"M+1": 0x37,
		"M-1": 0x32,
		"D+M": 0x2,
		"D-M": 0x13,
		"M-D": 0x7,
		"D&M": 0x0,
		"D|M": 0x15,
	}

	// jumpBit is a map of a jump command and a binary instruction.
	jumpBit opcBit = map[string]byte{
		"":    0x0,
		"JGT": 0x1,
		"JEQ": 0x2,
		"JGE": 0x3,
		"JLT": 0x4,
		"JNE": 0x5,
		"JLE": 0x6,
		"JMP": 0x7,
	}
)
