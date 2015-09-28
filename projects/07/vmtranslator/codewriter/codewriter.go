package codewriter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// baseLabel is a base name of labels.
const baseLabel = "LABEL"

// CodeWriter converts VM commands to Hack assembly codes and write them out to a destination.
type CodeWriter struct {
	err      error
	dest     io.Writer
	buf      *bufio.Writer
	filename string

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

// SetFileName sets an input VM file name to filename.
func (cw *CodeWriter) SetFileName(filename string) {
	cw.filename = filename
}

// WriteArithmetic converts the given arithmetic command to assembly code and writes it out.
func (cw *CodeWriter) WriteArithmetic(cmd string) error {
	switch cmd {
	case "neg", "not":
		cw.unary(cmd)
	case "add", "sub", "and", "or":
		cw.binary(cmd)
	case "eq", "gt", "lt":
		cw.compare(cmd)
	default:
		cw.err = fmt.Errorf("unknown command: %s", cmd)
	}

	if cw.err != nil {
		return fmt.Errorf("error writing code: %s", cw.err.Error())
	}
	return nil
}

// Close flushes bufferred data to the destination and closes it.
// Note that no data is written to the destination until Close is called.
func (cw *CodeWriter) Close() error {
	defer func() {
		if cl, ok := cw.dest.(io.Closer); ok {
			_ = cl.Close()
		}
	}()

	if e := cw.buf.Flush(); e != nil {
		return fmt.Errorf("error flushing bufferred data: %s", e)
	}
	return nil
}

// unary writes a unary operation for a value at the top of the stack.
// cmd must be one of the following:
//   - "neg"
//   - "not"
func (cw *CodeWriter) unary(cmd string) {
	var op string
	switch cmd {
	case "neg":
		op = "-"
	case "not":
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
	case "add":
		op = "M+D"
	case "sub":
		op = "M-D"
	case "and":
		op = "D&M"
	case "or":
		op = "D|M"
	}

	cw.pop()
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

	cw.pop()
	cw.decrSP()
	cw.ccmd("D", "M-D")
	cw.acmd(label1)
	cw.ccmdj("", "D", op)
	cw.loadToSP("0")
	cw.acmd(label2)
	cw.ccmdj("", "0", "JMP")
	cw.lcmd(label1)
	cw.loadToSP("-1")
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

// push pushes v to the top of the stack. Internally,
// it assgins v to *SP and increments SP.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) push(v string) {
	cw.loadToSP(v)
	cw.incrSP()
}

// loadToSP loads v to *SP.
func (cw *CodeWriter) loadToSP(v string) {
	cw.acmd("SP")
	cw.ccmd("A", "M")
	cw.ccmd("M", v)
}

// pop pops a value at the top of the stack. Internally,
// it decrements SP and assigns a value pointed by SP to D.
// If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) pop() {
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
	cw.acmd("SP")
	cw.ccmd("AM", "M"+op+"1")
}

// acmd writes @ command. If an error occurs and cw.err is nil, it is set at cw.err.
func (cw *CodeWriter) acmd(addr string) {
	if _, e := cw.buf.WriteString("@" + addr + "\n"); cw.err == nil {
		cw.err = e
	}
}

// ccmd writes C command with no jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) ccmd(dest, comp string) {
	cw.ccmdj(dest, comp, "")
}

// ccmdj writes C command with jump. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) ccmdj(dest, comp, jump string) {
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

	if _, e := cw.buf.Write(opc); cw.err == nil {
		cw.err = e
	}
}

// lcmd writes label command. If an error occurs, it is set at cw.err.
func (cw *CodeWriter) lcmd(label string) {
	if _, e := cw.buf.WriteString("(" + label + ")\n"); cw.err == nil {
		cw.err = e
	}
}
