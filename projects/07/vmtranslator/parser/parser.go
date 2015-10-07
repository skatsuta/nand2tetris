package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"
)

const (
	// prefixComment is a prefix of a comment.
	prefixComment = "//"

	// vmScanMode is a mode for scanning VM code.
	vmScanMode = scanner.ScanIdents | scanner.ScanInts | scanner.SkipComments
)

// CommandType represents a type of VM command.
type CommandType int

// A list of command types.
const (
	unknown CommandType = iota
	Arithmetic
	Push
	Pop
	Label
	Goto
	If
	Function
	Return
	Call
)

// command has command information.
type command struct {
	typ  CommandType
	arg1 string
	arg2 uint
}

// Parser is a parser for VM code.
// Parser is not thread safe, so it should NOT be used in multiple goroutines.
type Parser struct {
	src    *bufio.Scanner
	line   string
	tokens []string
	cmd    command
}

// New creates a new parser object that reads and parses r.
func New(src io.Reader) *Parser {
	return &Parser{
		src: bufio.NewScanner(src),
	}
}

// HasMoreCommands reports whether there exist more commands in input.
func (p *Parser) HasMoreCommands() bool {
	// if Scan() == true && Text() is not a comment, return true
	// if Scan() == false, return false
	for p.src.Scan() {
		p.line = p.src.Text()
		p.tokens = p.tokens[:0]

		// prepare a Scanner
		var sc scanner.Scanner
		sc.Mode = vmScanMode
		sc.Init(strings.NewReader(p.line))

		// tokenize the current line
		var tok rune
		for tok != scanner.EOF {
			tok = sc.Scan()
			text := sc.TokenText()

			// ignore empty string
			if text == "" {
				continue
			}

			p.tokens = append(p.tokens, text)
		}

		if len(p.tokens) > 0 {
			return true
		}
	}

	p.tokens = p.tokens[:0]
	return false
}

// Advance reads next command from source and set the command to current one.
// If the next command is invalid, it returns an error.
// This method should be called only if HasMoreCommands() returns true.
func (p *Parser) Advance() error {
	tokens := p.tokens

	// check the length of tokens: only 1 or 3 is valid
	switch len(tokens) {
	case 1, 3:
		// valid length; skip
	default:
		return fmt.Errorf("invalid command: %q", p.line)
	}

	// parse the first token as an opcode
	cmd := tokens[0]
	typ := p.dispatchCommand(cmd)
	if typ == unknown {
		return fmt.Errorf("unknown command: %s", cmd)
	}
	p.cmd.typ = typ
	if typ == Arithmetic {
		p.cmd.arg1 = cmd
		p.cmd.arg2 = 0
		return nil
	}

	// parse the second token as a segment
	seg := tokens[1]
	if !segs.contains(seg) {
		return fmt.Errorf("unknown segment: %s", seg)
	}
	p.cmd.arg1 = seg

	// parse the third token as an integer
	a := tokens[2]
	i, err := strconv.Atoi(a)
	if i < 0 || err != nil {
		return fmt.Errorf("not a positive integer: %s", a)
	}
	p.cmd.arg2 = uint(i)

	return nil
}

// dispatchCommand dispatches CommandType from cmd.
// If cmd is not a valid command string, it returns `unknown`.
func (*Parser) dispatchCommand(cmd string) CommandType {
	switch cmd {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return Arithmetic
	case "push":
		return Push
	case "pop":
		return Pop
	case "label":
		return Label
	case "goto":
		return Goto
	case "if-goto":
		return If
	case "function":
		return Function
	case "call":
		return Call
	case "return":
		return Return
	default:
		return unknown
	}
}

// segments is a collection of segments.
type segments []string

// segs is a collection of all the segments on VM.
// TODO use map[string]struct{}
var segs = segments{
	"argument",
	"local",
	"static",
	"constant",
	"this", "that",
	"pointer",
	"temp",
}

// contains reports whether text is contained in s.
func (s segments) contains(text string) bool {
	for _, seg := range s {
		if seg == text {
			return true
		}
	}
	return false
}

// CommandType returns a type of a current VM command. In all arithmetic commands
// it returns Arithmetic.
func (p *Parser) CommandType() CommandType {
	return p.cmd.typ
}

// Arg1 returns the first argument in a current command. If a type of the current command is
// Arithmetic, it returns the command itself. This method should NOT be called if CommandType()
// returns Return.
func (p *Parser) Arg1() string {
	return p.cmd.arg1
}

// Arg2 returns the second argument in a current command. This method should be called
// only if CommandType() returns Push, Pop, Function or Call.
func (p *Parser) Arg2() uint {
	return p.cmd.arg2
}
