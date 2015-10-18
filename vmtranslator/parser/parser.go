package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

const (
	// prefixComment is a prefix of a comment.
	prefixComment = "//"

	// vmScanMode is a mode for scanning VM code.
	vmScanMode = scanner.ScanIdents | scanner.ScanInts | scanner.SkipComments
)

// ErrInvalidCommand is an error represenitng a command is invalid.
var ErrInvalidCommand = errors.New("invalid command")

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
	sc := scanner.Scanner{
		Mode: vmScanMode,
		IsIdentRune: func(ch rune, i int) bool {
			// make Scanner recognize '-' as an identifier too
			// cf. text/scanner.Scanner.isIdentRune()
			return ch == '_' || ch == '-' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
		},
	}

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
		if err == ErrInvalidCommand {
			return fmt.Errorf("%v: %s", ErrInvalidCommand, p.line)
		}
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
	case Arithmetic, Return:
		return p.parse1(typ, tokens)
	case Label, Goto, If:
		return p.parse2(typ, tokens)
	case Push, Pop, Function, Call:
		return p.parse3(typ, tokens)
	default:
		return command{}, fmt.Errorf("unknown command: %s", cmd)
	}
}

// parse1 parses a command that should have one token.
func (p *Parser) parse1(typ CommandType, tokens []string) (command, error) {
	if len(tokens) != 1 {
		return command{}, ErrInvalidCommand
	}

	var arg1 string
	if typ == Arithmetic {
		arg1 = tokens[0]
	}

	return command{typ: typ, arg1: arg1}, nil
}

// parse2 parses a command that should have two tokens.
func (p *Parser) parse2(typ CommandType, tokens []string) (command, error) {
	if len(tokens) != 2 {
		return command{}, ErrInvalidCommand
	}
	return command{typ: typ, arg1: tokens[1]}, nil
}

// parse3 parses a command that should have three tokens.
func (p *Parser) parse3(typ CommandType, tokens []string) (command, error) {
	if len(tokens) != 3 {
		return command{}, ErrInvalidCommand
	}

	arg1 := tokens[1]
	if typ == Push || typ == Pop {
		// check the validation of a segment
		if !segs.contains(arg1) {
			return command{}, fmt.Errorf("unknown segment: %s", arg1)
		}
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
// In order to check an existence of a segment by O(1) segments is a map.
type segments map[string]struct{}

// dummy is a dummy empty struct for segments.
var dummy = struct{}{}

// segs is a collection of all the segments on VM.
var segs = segments{
	"constant": dummy,
	"local":    dummy,
	"argument": dummy,
	"this":     dummy,
	"that":     dummy,
	"temp":     dummy,
	"pointer":  dummy,
	"static":   dummy,
}

// contains reports whether text is contained in s.
func (s segments) contains(text string) bool {
	_, exists := segs[text]
	return exists
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
