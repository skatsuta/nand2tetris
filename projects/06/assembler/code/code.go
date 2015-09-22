package code

import "fmt"

// instSet is a map of an opcode and a binary opcode.
type instSet map[string]byte

// contains reports whether mneum is contained in instSet.
func (is instSet) isValid(mneum string) bool {
	_, found := is[mneum]
	return found
}

var (
	// destInstSet is a map of dest mneumonics and its binary opcodes.
	destInstSet = instSet{
		"":    0x0,
		"M":   0x1,
		"D":   0x2,
		"MD":  0x3,
		"A":   0x4,
		"AM":  0x5,
		"AD":  0x6,
		"AMD": 0x7,
	}

	// compInstSet is a map of comp mneumonics and its binary opcodes.
	compInstSet = instSet{
		// a = 0
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
		// a = 1
		"M":   0x70,
		"!M":  0x71,
		"-M":  0x73,
		"M+1": 0x77,
		"M-1": 0x72,
		"D+M": 0x42,
		"D-M": 0x53,
		"M-D": 0x47,
		"D&M": 0x40,
		"D|M": 0x55,
	}

	// jumpInstSet is a map of jump mneumonics and its binary opcodes.
	jumpInstSet instSet = map[string]byte{
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

// Code is a converter from mneumonics to binary codes.
type Code struct {
}

// dest returns 3 bit binary opcode corresponding to the dest mneumonic.
func (c *Code) dest(mneum string) (byte, error) {
	if !destInstSet.isValid(mneum) {
		return 0, fmt.Errorf("invalid dest mneumonic: %s", mneum)
	}
	return destInstSet[mneum], nil
}

// IsValidDest reports whether mneum is a valid dest mneumonic.
func (c *Code) IsValidDest(mneum string) bool {
	return destInstSet.isValid(mneum)
}

// comp returns 7 bit binary opcode corresponding to the comp mneumonic.
func (c *Code) comp(mneum string) (byte, error) {
	if !compInstSet.isValid(mneum) {
		return 0, fmt.Errorf("invalid comp mneumonic: %s", mneum)
	}
	return compInstSet[mneum], nil
}

// IsValidComp reports whether mneum is a valid comp mneumonic.
func (c *Code) IsValidComp(mneum string) bool {
	return compInstSet.isValid(mneum)
}

// jump returns 3 bit binary code corresponding to the jump mneumonic.
func (c *Code) jump(mneum string) (byte, error) {
	if !jumpInstSet.isValid(mneum) {
		return 0, fmt.Errorf("invalid jump mneumonic: %s", mneum)
	}
	return jumpInstSet[mneum], nil
}

// IsValidJump reports whether mneum is a valid jump mneumonic.
func (c *Code) IsValidJump(mneum string) bool {
	return jumpInstSet.isValid(mneum)
}
