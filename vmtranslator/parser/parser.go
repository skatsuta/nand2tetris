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
	// vmScanMode is a mode for scanning VM code.
	vmScanMode = scanner.ScanIdents | scanner.ScanInts | scanner.SkipComments
)

var (
	// ErrInvalidCommand is an error represenitng a command is invalid.
	ErrInvalidCommand = errors.New("invalid command")
	// ErrNoTokens is an error representing no tokens.
	ErrNoTokens = errors.New("no tokens")
)

// CommandType represents a type of VM command.
type CommandType string

// A list of command types.
const (
	Unknown    CommandType = "unknown"
	Arithmetic CommandType = "arithmetic"
	Push       CommandType = "push"
	Pop        CommandType = "pop"
	Label      CommandType = "label"
	Goto       CommandType = "goto"
	If         CommandType = "if-goto"
	Function   CommandType = "function"
	Return     CommandType = "return"
	Call       CommandType = "call"
)

// Command represents a single virtual machine instruction.
type Command struct {
	Type CommandType
	Arg1 string
	Arg2 uint
}

func (c Command) String() string {
	switch c.Type {
	case Arithmetic:
		return c.Arg1
	case Return:
		return string(c.Type)
	case Label, Goto, If:
		return fmt.Sprintf("%s %s", c.Type, c.Arg1)
	default:
		return fmt.Sprintf("%s %s %d", c.Type, c.Arg1, c.Arg2)
	}
}

// isIdentRune is a predicate controlling the characters accepted as the ith rune in an identifier.
func isIdentRune(ch rune, i int) bool {
	// make Scanner recognize non-first '-' and '.' as additional identifiers
	// cf. text/scanner.Scanner.isIdentRune()
	return ch == '_' || unicode.IsLetter(ch) ||
		i > 0 && (unicode.IsDigit(ch) || ch == '-' || ch == '.')
}

// Parser is a parser for VM code.
// Parser is not thread safe, so it should NOT be used in multiple goroutines.
type Parser struct {
	src    *bufio.Scanner
	sc     *scanner.Scanner
	line   string
	tokens []string
	cmd    Command
}

// New creates a new parser object that reads and parses src.
func New(src io.Reader) *Parser {
	return &Parser{
		src: bufio.NewScanner(src),
		sc: &scanner.Scanner{
			Mode:        vmScanMode,
			IsIdentRune: isIdentRune,
		},
	}
}

// Command returns the current command p has read.
func (p *Parser) Command() Command {
	return p.cmd
}

// HasMoreCommands reports whether there exist more commands in input.
func (p *Parser) HasMoreCommands() bool {
	// if Scan() == true && Text() is not a comment, return true
	// if Scan() == false, return false
	for p.src.Scan() {
		p.line = p.src.Text()
		p.tokens = p.tokens[:0]

		// prepare a Scanner
		p.sc.Init(strings.NewReader(p.line))

		// tokenize the current line
		var tok rune
		for tok != scanner.EOF {
			tok = p.sc.Scan()
			text := p.sc.TokenText()

			// ignore empty line
			if text != "" {
				p.tokens = append(p.tokens, text)
			}
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
		return fmt.Errorf("error parsing %q: %w", p.line, err)
	}
	p.cmd = cmd
	return nil
}

// parse parses tokens and returns a command object.
// If it fails to parse tokens, it returns an error.
func (p *Parser) parse(tokens []string) (Command, error) {
	// check the length of tokens: should be less than 4
	if len(tokens) == 0 {
		return Command{}, ErrNoTokens
	}

	// parse the first token as an opcode
	switch typ := p.dispatchCommand(tokens[0]); typ {
	case Arithmetic, Return:
		return p.parse1(typ, tokens)
	case Label, Goto, If:
		return p.parse2(typ, tokens)
	case Push, Pop, Function, Call:
		return p.parse3(typ, tokens)
	default:
		return Command{}, fmt.Errorf("unknown command: %s", tokens[0])
	}
}

// parse1 parses a command that should have one token.
func (p *Parser) parse1(typ CommandType, tokens []string) (Command, error) {
	if len(tokens) != 1 {
		return Command{}, ErrInvalidCommand
	}

	if typ == Arithmetic {
		return Command{Type: typ, Arg1: tokens[0]}, nil
	}
	return Command{Type: typ}, nil
}

// parse2 parses a command that should have two tokens.
func (p *Parser) parse2(typ CommandType, tokens []string) (Command, error) {
	if len(tokens) != 2 {
		return Command{}, ErrInvalidCommand
	}
	return Command{Type: typ, Arg1: tokens[1]}, nil
}

// parse3 parses a command that should have three tokens.
func (p *Parser) parse3(typ CommandType, tokens []string) (Command, error) {
	if len(tokens) != 3 {
		return Command{}, ErrInvalidCommand
	}

	arg1 := tokens[1]
	if typ == Push || typ == Pop {
		// check the validation of a segment
		if !segs.contains(arg1) {
			return Command{}, fmt.Errorf("unknown segment: %s", arg1)
		}
	}

	// parse the third token as an integer
	a := tokens[2]
	i, err := strconv.ParseUint(a, 10, 64)
	if i < 0 || err != nil {
		return Command{}, fmt.Errorf("not a positive integer: %s", a)
	}

	return Command{Type: typ, Arg1: arg1, Arg2: uint(i)}, nil
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
		return Unknown
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
