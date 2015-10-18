package parser

import (
	"bufio"
	"errors"
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
	sc := scanner.Scanner{Mode: vmScanMode}

	// if Scan() == true && Text() is not a comment, return true
	// if Scan() == false, return false
	for p.src.Scan() {
		p.line = p.src.Text()
		p.tokens = p.tokens[:0]

		// prepare a Scanner
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
	cmd, err := p.parse(p.tokens)
	if err != nil {
		return err
	}

	p.cmd = cmd
	return nil
}

// parse parses tokens and returns a command object.
// If it fails to parse tokens, it returns an error.
func (p *Parser) parse(tokens []string) (command, error) {
	// check the length of tokens: should be less than 4
	if len(tokens) == 0 {
		return command{}, errors.New("empty tokens")
	}

	// parse the first token as an opcode
	cmd := tokens[0]
	typ := p.dispatchCommand(cmd)

	switch typ {
	case Arithmetic:
		return p.parseArithmetic(tokens)
	case Push, Pop:
		return p.parsePushPop(typ, tokens)
	case Label:
		return p.parseLabel(tokens)
	default:
		return command{}, fmt.Errorf("unknown command: %s", cmd)
	}
}

// parseArithmetic parses an arithmetic command.
func (p *Parser) parseArithmetic(tokens []string) (command, error) {
	if len(tokens) != 1 {
		return command{}, fmt.Errorf("invalid arithmetic command: %s", p.line)
	}
	return command{typ: Arithmetic, arg1: tokens[0]}, nil
}

// parseLabel parses a label command.
func (p *Parser) parseLabel(tokens []string) (command, error) {
	if len(tokens) != 2 {
		return command{}, fmt.Errorf("invalid label command: %s", p.line)
	}
	return command{typ: Label, arg1: tokens[1]}, nil
}

// parsePushPop parses a push/pop command.
func (p *Parser) parsePushPop(typ CommandType, tokens []string) (command, error) {
	if len(tokens) != 3 {
		return command{}, fmt.Errorf("invalid push/pop command: %s", p.line)
	}

	arg1 := tokens[1]
	if !segs.contains(arg1) {
		return command{}, fmt.Errorf("unknown segment: %s", arg1)
	}

	// parse the third token as an integer
	a := tokens[2]
	i, err := strconv.Atoi(a)
	if i < 0 || err != nil {
		return command{}, fmt.Errorf("not a positive integer: %s", a)
	}

	return command{typ: typ, arg1: arg1, arg2: uint(i)}, nil
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
