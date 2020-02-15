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

const entrypoint = "Sys.init"

// CodeWriter converts VM commands to Hack assembly codes and write them out to a destination.
type CodeWriter struct {
	verbose bool

	err      error
	dest     io.Writer
	buf      *bufio.Writer
	filename string
	fnbase   string

	funcName string

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

// Verbose controls verbose mode of tr. If verbose mode is enabled, the CodeWriter also
// outputs each method call as a comment corresponding to each assembly code block.
func (cw *CodeWriter) Verbose(verbose bool) *CodeWriter {
	cw.verbose = verbose
	return cw
}

// SetFileName sets an input VM file name and writes it to the output file as comment.
func (cw *CodeWriter) SetFileName(filename string) error {
	cw.filename = filename
	cw.fnbase = cw.fileNameBase(filename)

	// TODO print an absolute path or just a base file name,
	// or a file name if it's a file and dir/file if it's a dir.
	return cw.WriteComment(filename)
}

// fileNameBase return a base name of a file.
// For example, if a filename is "foo.txt", it returns "foo".
func (cw *CodeWriter) fileNameBase(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return base[:len(base)-len(ext)]
}

// WriteComment writes a comment.
func (cw *CodeWriter) WriteComment(comment string) error {
	_, err := cw.buf.WriteString("// " + comment + "\n")
	return err
}

// debug prints debugging message. Args are formatted with printf verbs (such as %v) in msg.
func (cw *CodeWriter) debug(msg string, a ...interface{}) {
	if !cw.verbose {
		return
	}

	err := cw.WriteComment(fmt.Sprintf("[DEBUG] CodeWriter#"+msg, a...))
	if err != nil && cw.err == nil {
		cw.err = err
	}
}

// WriteInit writes out bootstrap code.
func (cw *CodeWriter) WriteInit() error {
	cw.debug("WriteInit()")

	cw.loadVal(256, false)
	cw.WriteGoto(entrypoint)
	return cw.err
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
	cw.debug("WritePushPop(cmd=%q, seg=%q, idx=%d)", cmd, seg, idx)

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
	if cw.funcName != "" {
		label = cw.funcName + "$" + label
	}

	cw.lcmd(label)
	return cw.err
}

// WriteGoto converts the given goto command to assembly code and writes it out.
func (cw *CodeWriter) WriteGoto(label string) error {
	cw.debug("WriteGoto(label=%q)", label)

	cw.acmd(label)
	cw.jump("0", "JMP")
	return cw.err
}

// WriteIf converts the given if-goto command to assembly code and writes it out.
func (cw *CodeWriter) WriteIf(label string) error {
	cw.decrSP()
	cw.store("D", "M")
	cw.acmd(label)
	cw.jump("D", "JNE")
	return cw.err
}

// WriteFunction converts the given function command to assembly code and writes it out.
func (cw *CodeWriter) WriteFunction(funcName string, numLocals uint) error {
	cw.debug("WriteFunction(funcName=%q, numLocals=%d)", funcName, numLocals)

	// Save current function name
	cw.funcName = funcName

	cw.lcmd(funcName)

	// initialize a variable pointed by symb + idx to 0.
	for i := 0; i < int(numLocals); i++ {
		cw.pushVal(0)
	}

	return cw.err
}

// WriteReturn writes out the assembly code of return statement.
func (cw *CodeWriter) WriteReturn() error {
	cw.loadSeg(regLCL, 0, true)
	cw.saveTo(regR14, false)
	cw.loadSeg(regR14, -5, true)
	cw.store("D", "M")
	cw.saveTo(regR15, false)
	cw.popStack()
	cw.saveTo(regARG, true)
	cw.loadSeg(regARG, 1, true)
	cw.saveTo(regSP, false)
	for i, reg := range []string{regTHAT, regTHIS, regARG, regLCL} {
		cw.loadSeg(regR14, -i-1, true)
		cw.store("D", "M")
		cw.saveTo(reg, false)
	}
	cw.acmd(regR15)
	cw.store("A", "M")
	cw.jump("0", "JMP")
	return cw.err
}

// WriteCall converts the given function call command to assembly code and writes it out.
func (cw *CodeWriter) WriteCall(funcName string, numArgs uint) error {
	cw.debug("WriteCall(funcName=%q, numArgs=%d)", funcName, numArgs)

	// Push return address onto the stack
	retAddrLbl := cw.uniqueLabel(funcName + "_RET_ADDR")
	cw.loadSymb(retAddrLbl, false) // Load the address at the label to D register
	cw.pushStack()                 // Push the address onto the stack for function return

	// Push local/argument/this/that segment base addresses onto the stack
	for _, symb := range []string{regLCL, regARG, regTHIS, regTHAT} {
		cw.loadSymb(symb, true) // Load the value the address of the symbol points to to D register
		cw.pushStack()          // Push the address onto the stack for function return
	}

	// Update argument segment for the callee function
	cw.loadSeg(regSP, -int(numArgs+5), true)
	cw.saveTo(regARG, false)

	// Update local segment for the callee function
	cw.loadSeg(regSP, 0, true)
	cw.saveTo(regLCL, false)

	// Go to the callee function
	cw.WriteGoto(funcName)

	// Mark the return address
	cw.lcmd(retAddrLbl)

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
	cw.jump("0", "JMP")
	return cw.err
}

// uniqueLabel generates a temporary label from a given label. It returns different labels even
// if the same label is given multiple times.
func (cw *CodeWriter) uniqueLabel(label string) string {
	defer cw.countUp()
	return fmt.Sprintf("%s_%d", label, cw.cnt)
}

// countUp counts up an internal counter.
func (cw *CodeWriter) countUp() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.cnt++
}

// push converts the given push command to assembly and writes it out.
func (cw *CodeWriter) push(seg string, idx uint) error {
	cw.debug("push(seg=%q, idx=%d)", seg, idx)

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
	cw.loadVal(int(v), true)
	cw.incrSP()
}

// pushMem pushes a value pointed by an address in seg to the stack.
func (cw *CodeWriter) pushMem(seg string, idx uint) {
	cw.debug("pushMem(seg=%q, idx=%d)", seg, idx)

	cw.push0(seg, idx, true)
}

// pushReg pushes a value in reg to the stack.
func (cw *CodeWriter) pushReg(reg string, idx uint) {
	cw.push0(reg, idx, false)
}

// pushStatic loads a value of the static segment to *SP.
func (cw *CodeWriter) pushStatic(idx uint) {
	// direct is ignored so meaningless
	cw.push0("STATIC", idx, true)
}

// push0 pushes a value of symb to the top of the stack.
// If symb is "STATIC", it pushes idx-th static variable.
// If indirect is true a value pointed by an address in symb indirectly,
// otherwise a value in symb is pushed directly.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) push0(symb string, idx uint, indirect bool) {
	if symb == "STATIC" {
		cw.acmd(fmt.Sprintf("%s.%d", cw.fnbase, idx))
	} else {
		cw.loadSeg(symb, int(idx), indirect)
	}
	cw.store("D", "M")
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
	cw.pop0(seg, idx, true)
}

// popReg pops a value from the stack and stores it to reg directly.
func (cw *CodeWriter) popReg(reg string, idx uint) {
	cw.pop0(reg, idx, false)
}

// popStatic pops a value from the stack and stores it to the static segment.
func (cw *CodeWriter) popStatic(idx uint) {
	cw.decrSP()
	cw.store("D", "M")
	cw.acmd(fmt.Sprintf("%s.%d", cw.fnbase, idx))
	cw.store("M", "D")
}

// pop0 pops a value from the top of the stack to symb.
// If symb is "STATIC", it pops idx-th static variable.
// If indirect is true a value pointed by an address in symb indirectly,
// otherwise a value in symb is popped directly.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) pop0(symb string, idx uint, indirect bool) {
	tmpreg := regR13

	cw.loadSeg(symb, int(idx), indirect)
	cw.acmd(tmpreg)
	cw.store("M", "D")
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
	cw.store("M", op+"M")
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
	cw.store("M", op)
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
	label1 := cw.uniqueLabel(baseLabel)
	label2 := cw.uniqueLabel(baseLabel)

	cw.popStack()
	cw.decrSP()
	cw.store("D", "M-D")
	cw.acmd(label1)
	cw.jump("D", op)
	cw.loadVal(bitFalse, true)
	cw.acmd(label2)
	cw.jump("0", "JMP")
	cw.lcmd(label1)
	cw.loadVal(bitTrue, true)
	cw.lcmd(label2)
	cw.incrSP()
}

// loadSymb loads the address of a given symbol or label to D register.
// If indirect is true, it loads the value the address points to instead.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) loadSymb(symb string, indirect bool) {
	cw.debug("loadSymb(symb=%q, indirect=%t)", symb, indirect)

	cw.acmd(symb) // Load the address of the label or symbol to A register
	if indirect {
		cw.store("D", "M") // Store the value the address points to to D register
	} else {
		cw.store("D", "A") // Store the value of the address itself to D register
	}
}

// loadSeg loads a value of the symb segment shifted by idx to D register.
// If indirect is true, a value pointed by an address in symb is loaded instead.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) loadSeg(symb string, idx int, indirect bool) {
	cw.debug("loadSeg(symb=%q, idx=%d, indirect=%t)", symb, idx, indirect)

	m := "A"
	if indirect {
		m = "M"
	}

	// get the absolute value of idx and its sign
	abs := idx
	rhs := "D+" + m
	if idx < 0 {
		abs = -idx
		rhs = m + "-D"
	}

	cw.acmd(abs)
	cw.store("D", "A")
	cw.acmd(symb)
	cw.store("AD", rhs)
}

// loadVal loads v to *SP. v should be greater than or equal -1 (v >= -1).
func (cw *CodeWriter) loadVal(v int, indirect bool) {
	cw.debug("loadVal(v=%d, indirect=%t)", v, indirect)

	if v < 0 {
		cw.acmd(regSP)
		cw.store("A", "M")
		cw.store("M", strconv.Itoa(v))
		return
	}

	cw.acmd(v)
	cw.store("D", "A")
	cw.saveTo(regSP, indirect)
}

// pushStack pushs a value in D register onto the stack.
// If an error occurs and cw.err is nil, it is set at cw.err.
//
// Internally, it writes a value in D register to a memory address pointed by SP
// and increments SP.
func (cw *CodeWriter) pushStack() {
	cw.debug("pushStack()")

	cw.saveTo(regSP, true)
	cw.incrSP()
}

// popStack pops a value at the top of the stack. Internally,
// it decrements SP and assigns a value pointed by SP to D.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) popStack() {
	cw.decrSP()
	cw.store("D", "M")
}

// saveTo saves the value in D register to a memory address pointed by addr.
// If indirect is true, it saves D to *addr instead of addr.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) saveTo(addr string, indirect bool) {
	cw.debug("saveTo(addr=%q, indirect=%t)", addr, indirect)

	cw.acmd(addr)
	if indirect {
		cw.store("A", "M")
	}
	cw.store("M", "D")
}

// incrSP increments SP and sets the current address to it.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) incrSP() {
	cw.debug("incrSP()")

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
	cw.store("AM", "M"+op+"1")
}

// acmd writes @ command. If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) acmd(addr interface{}) {
	if cw.err != nil {
		return
	}

	a := fmt.Sprintf("@%v\n", addr)
	_, cw.err = cw.buf.WriteString(a)
}

// store writes C command with no jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) store(dest, comp string) {
	cw.ccmd(dest, comp, "")
}

// jump writes C command with jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) jump(comp, jump string) {
	cw.ccmd("", comp, jump)
}

// ccmd writes raw C command. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) ccmd(dest, comp, jump string) {
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
	cw.debug("lcmd(label=%q)", label)

	if cw.err != nil {
		return
	}

	lbl := fmt.Sprintf("(%s)\n", label)
	_, cw.err = cw.buf.WriteString(lbl)
}
