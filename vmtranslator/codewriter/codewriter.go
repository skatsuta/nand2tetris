package codewriter

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/skatsuta/nand2tetris/vmtranslator/parser"
)

// binary representation of logical values
const (
	bitTrue  = -1
	bitFalse = 0
)

// baseLabel is a base name of labels.
const baseLabel = "LABEL"

// Mneumonics of arithmetic operations.
const (
	opAdd = "add"
	opSub = "sub"
	opNeg = "neg"
	opEq  = "eq"
	opGt  = "gt"
	opLt  = "lt"
	opAnd = "and"
	opOr  = "or"
	opNot = "not"
)

// Memory segments.
const (
	segArgument = "argument"
	segLocal    = "local"
	segStatic   = "static"
	segConstant = "constant"
	segThis     = "this"
	segThat     = "that"
	segPointer  = "pointer"
	segTemp     = "temp"
)

// Registers and its aliases.
const (
	regSP   = "SP"
	regLCL  = "LCL"
	regARG  = "ARG"
	regTHIS = "THIS"
	regTHAT = "THAT"
	regR3   = "R3"
	regR5   = "R5"
	regR13  = "R13"
	regR14  = "R14"
	regR15  = "R15"
)

// CodeWriter converts VM commands to Hack assembly codes and write them out to a destination.
type CodeWriter struct {
	err      error
	dest     io.Writer
	buf      *bufio.Writer
	filename string
	fnbase   string

	mu  sync.Mutex
	cnt int
}

// New creates a new CodeWriter that writes converted codes to dest.
func New(dest io.Writer) *CodeWriter {
	return &CodeWriter{
		dest: dest,
		buf:  bufio.NewWriter(dest),
	}
}

// SetFileName sets an input VM file name and writes it to the output file as comment.
func (cw *CodeWriter) SetFileName(filename string) error {
	cw.filename = filename
	cw.fnbase = cw.fileNameBase(filename)

	// TODO print an absolute path or just a base file name,
	// or a file name if it's a file and dir/file if it's a dir.
	_, err := cw.buf.WriteString(fmt.Sprintln("//", filename))
	return err
}

// fileNameBase return a base name of a file.
// For example, if a filename is "foo.txt", it returns "foo".
func (cw *CodeWriter) fileNameBase(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return base[:len(base)-len(ext)]
}

// WriteArithmetic converts the given arithmetic command to assembly code and writes it out.
func (cw *CodeWriter) WriteArithmetic(cmd string) error {
	switch cmd {
	case opNeg, opNot:
		cw.unary(cmd)
	case opAdd, opSub, opAnd, opOr:
		cw.binary(cmd)
	case opEq, opGt, opLt:
		cw.compare(cmd)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
	return cw.err
}

// WritePushPop converts the given push or pop command to assembly code and writes it out.
func (cw *CodeWriter) WritePushPop(cmd parser.CommandType, seg string, idx uint) error {
	switch cmd {
	case parser.Push:
		return cw.push(seg, idx)
	case parser.Pop:
		return cw.pop(seg, idx)
	default:
		return fmt.Errorf("unknown command: %v", cmd)
	}
}

// WriteLabel converts the given label command to assembly code and writes it out.
func (cw *CodeWriter) WriteLabel(label string) error {
	cw.lcmd(label)
	return cw.err
}

// WriteGoto converts the given goto command to assembly code and writes it out.
func (cw *CodeWriter) WriteGoto(label string) error {
	cw.acmd(label)
	cw.ccmdj("", "0", "JMP")
	return cw.err
}

// WriteIf converts the given if-goto command to assembly code and writes it out.
func (cw *CodeWriter) WriteIf(label string) error {
	cw.decrSP()
	cw.ccmd("D", "M")
	cw.acmd(label)
	cw.ccmdj("", "D", "JNE")
	return cw.err
}

// WriteFunction converts the given function command to assembly code and writes it out.
func (cw *CodeWriter) WriteFunction(funcName string, numLocals uint) error {
	cw.lcmd(funcName)

	// initialize a variable pointed by symb + idx to 0.
	for i := 0; i < int(numLocals); i++ {
		cw.loadSeg(regLCL, i, false)
		cw.ccmd("M", "0")
	}

	return cw.err
}

// Close flushes bufferred data to the destination and closes it.
// Note that no data is written to the destination until Close is called.
func (cw *CodeWriter) Close() error {
	defer func() {
		if cl, ok := cw.dest.(io.Closer); ok {
			_ = cl.Close()
		}
	}()

	// write the end infinite loop
	if e := cw.end(); e != nil {
		return fmt.Errorf("error writing the end infinite loop: %v", e)
	}

	if e := cw.buf.Flush(); e != nil {
		return fmt.Errorf("error flushing bufferred data: %s", e)
	}
	return nil
}

// end writes the end infinite loop.
func (cw *CodeWriter) end() error {
	cw.lcmd("END")
	cw.acmd("END")
	cw.ccmdj("", "0", "JMP")
	return cw.err
}

// push converts the given push command to assembly and writes it out.
func (cw *CodeWriter) push(seg string, idx uint) error {
	switch seg {
	case segConstant:
		cw.pushVal(idx)
	case segLocal:
		cw.pushMem(regLCL, idx)
	case segArgument:
		cw.pushMem(regARG, idx)
	case segThis:
		cw.pushMem(regTHIS, idx)
	case segThat:
		cw.pushMem(regTHAT, idx)
	case segTemp:
		// temp: R5 ~ R12
		cw.pushReg(regR5, idx)
		// pointer: R3 ~ R4
	case segPointer:
		cw.pushReg(regR3, idx)
	case segStatic:
		cw.pushStatic(idx)
	default:
		return fmt.Errorf("unknown segment: %s", seg)
	}

	return cw.err
}

// pushVal pushes v to the top of the stack. Internally,
// it assgins v to *SP and increments SP.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) pushVal(v uint) {
	cw.loadVal(int(v))
	cw.incrSP()
}

// pushMem pushes a value pointed by an address in seg to the stack.
func (cw *CodeWriter) pushMem(seg string, idx uint) {
	cw.push0(seg, idx, false)
}

// pushReg pushes a value in reg to the stack.
func (cw *CodeWriter) pushReg(reg string, idx uint) {
	cw.push0(reg, idx, true)
}

// pushStatic loads a value of the static segment to *SP.
func (cw *CodeWriter) pushStatic(idx uint) {
	// direct is ignored so meaningless
	cw.push0("STATIC", idx, false)
}

// push0 pushes a value of symb to the top of the stack.
// If symb is "STATIC", it pushes idx-th static variable.
// If direct is true a value in symb is pushed directly,
// otherwise a value pointed by an address in symb indirectly.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) push0(symb string, idx uint, direct bool) {
	if symb == "STATIC" {
		cw.acmd(fmt.Sprintf("%s.%d", cw.fnbase, idx))
	} else {
		cw.loadSeg(symb, int(idx), direct)
	}
	cw.ccmd("D", "M")
	cw.saveTo(regSP, true)
	cw.incrSP()
}

// pop converts the given pop command to assembly and writes it out.
func (cw *CodeWriter) pop(seg string, idx uint) error {
	switch seg {
	case segLocal:
		cw.popMem(regLCL, idx)
	case segArgument:
		cw.popMem(regARG, idx)
	case segThis:
		cw.popMem(regTHIS, idx)
	case segThat:
		cw.popMem(regTHAT, idx)
	case segTemp:
		// temp: R5 ~ R12
		cw.popReg(regR5, idx)
	case segPointer:
		// pointer R3 ~ R4
		cw.popReg(regR3, idx)
	case segStatic:
		cw.popStatic(idx)
	default:
		return fmt.Errorf("unknown segment: %s", seg)
	}

	return cw.err
}

// popMem pops a value from the stack and stores it to an address seg points to.
func (cw *CodeWriter) popMem(seg string, idx uint) {
	cw.pop0(seg, idx, false)
}

// popReg pops a value from the stack and stores it to reg directly.
func (cw *CodeWriter) popReg(reg string, idx uint) {
	cw.pop0(reg, idx, true)
}

// popStatic pops a value from the stack and stores it to the static segment.
func (cw *CodeWriter) popStatic(idx uint) {
	cw.decrSP()
	cw.ccmd("D", "M")
	cw.acmd(fmt.Sprintf("%s.%d", cw.fnbase, idx))
	cw.ccmd("M", "D")
}

// pop0 pops a value from the top of the stack to symb.
// If symb is "STATIC", it pops idx-th static variable.
// If direct is true a value in symb is popped directly,
// otherwise a value pointed by an address in symb indirectly.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) pop0(symb string, idx uint, direct bool) {
	tmpreg := regR13

	cw.loadSeg(symb, int(idx), direct)
	cw.acmd(tmpreg)
	cw.ccmd("M", "D")
	cw.popStack()
	cw.saveTo(tmpreg, true)
}

// unary writes a unary operation for a value at the top of the stack.
// cmd must be one of the following:
//   - "neg"
//   - "not"
func (cw *CodeWriter) unary(cmd string) {
	var op string
	switch cmd {
	case opNeg:
		op = "-"
	case opNot:
		op = "!"
	}

	cw.decrSP()
	cw.ccmd("M", op+"M")
	cw.incrSP()
}

// binary writes a binary operation for two values at the top of the stack.
// cmd must be one of the following:
//   - "add"
//   - "sub"
//   - "and"
//   - "or"
func (cw *CodeWriter) binary(cmd string) {
	var op string
	switch cmd {
	case opAdd:
		op = "D+M"
	case opSub:
		op = "M-D"
	case opAnd:
		op = "D&M"
	case opOr:
		op = "D|M"
	}

	cw.popStack()
	cw.decrSP()
	cw.ccmd("M", op)
	cw.incrSP()
}

// compare writes a comparison operation for two values at the top of the stack.
// cmd must be one of the following:
//   - "eq"
//   - "gt"
//   - "lt"
func (cw *CodeWriter) compare(cmd string) {
	// JEQ, JGT, JLT
	op := "J" + strings.ToUpper(cmd)
	label1, label2 := cw.label(), cw.label()

	cw.popStack()
	cw.decrSP()
	cw.ccmd("D", "M-D")
	cw.acmd(label1)
	cw.ccmdj("", "D", op)
	cw.loadVal(bitFalse)
	cw.acmd(label2)
	cw.ccmdj("", "0", "JMP")
	cw.lcmd(label1)
	cw.loadVal(bitTrue)
	cw.lcmd(label2)
	cw.incrSP()
}

// label returns a label.
func (cw *CodeWriter) label() string {
	defer cw.countUp()
	return baseLabel + strconv.Itoa(cw.cnt)
}

// countUp counts up an internal counter.
func (cw *CodeWriter) countUp() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.cnt++
}

// loadSeg loads a value of the symb segment shifted by idx to D.
// If direct is true a value in symb is loaded directly,
// otherwise a value pointed by an address in symb indirectly.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) loadSeg(symb string, idx int, direct bool) {
	m := "M"
	if direct {
		m = "A"
	}

	// get the absolute value of idx and its sign
	abs := idx
	rhs := "D+" + m
	if idx < 0 {
		abs = -idx
		rhs = m + "-D"
	}

	cw.acmd(abs)
	cw.ccmd("D", "A")
	cw.acmd(symb)
	cw.ccmd("AD", rhs)
}

// saveTo save the value of D to addr.
// If indirect is true it saves D to *addr instead of addr.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) saveTo(addr string, indirect bool) {
	cw.acmd(addr)
	if indirect {
		cw.ccmd("A", "M")
	}
	cw.ccmd("M", "D")
}

// loadVal loads v to *SP. v should be greater than or equal -1 (v >= -1).
func (cw *CodeWriter) loadVal(v int) {
	if v < 0 {
		cw.acmd(regSP)
		cw.ccmd("A", "M")
		cw.ccmd("M", strconv.Itoa(v))
		return
	}

	cw.acmd(v)
	cw.ccmd("D", "A")
	cw.saveTo(regSP, true)
}

// popStack pops a value at the top of the stack. Internally,
// it decrements SP and assigns a value pointed by SP to D.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) popStack() {
	cw.decrSP()
	cw.ccmd("D", "M")
}

// incrSP increments SP and sets the current address to it.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) incrSP() {
	cw.sp("+")
}

// decrSP increments SP and sets the current address to it.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) decrSP() {
	cw.sp("-")
}

// sp controls the position of SP and sets the current address to it.
// op must be one of the following:
//   "+": SP++
//   "-": SP--
func (cw *CodeWriter) sp(op string) {
	cw.acmd(regSP)
	cw.ccmd("AM", "M"+op+"1")
}

// acmd writes @ command. If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) acmd(addr interface{}) {
	if cw.err != nil {
		return
	}

	a := fmt.Sprintf("@%v\n", addr)
	_, cw.err = cw.buf.WriteString(a)
}

// ccmd writes C command with no jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) ccmd(dest, comp string) {
	cw.ccmdj(dest, comp, "")
}

// ccmdj writes C command with jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) ccmdj(dest, comp, jump string) {
	if cw.err != nil {
		return
	}

	// allocate a slice whose length is len(dest=comp;jump\n)
	opc := make([]byte, 0, len(dest)+1+len(comp)+1+len(jump)+1)

	// append `dest=`
	if dest != "" {
		opc = append(append(opc, dest...), '=')
	}

	// append comp
	opc = append(opc, comp...)

	// append `;jump`
	if jump != "" {
		opc = append(append(opc, ';'), jump...)
	}

	// append \n
	opc = append(opc, '\n')

	_, cw.err = cw.buf.Write(opc)
}

// lcmd writes label command. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) lcmd(label string) {
	if cw.err != nil {
		return
	}

	lbl := fmt.Sprintf("(%s)\n", label)
	_, cw.err = cw.buf.WriteString(lbl)
}
