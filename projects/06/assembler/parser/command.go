package parser

// command is an interface that represents a Hack command.
type command interface {
	// commandType returns a type of the command.
	commandType() commandType
}

// newCommand returns a command object corresponding to cmd.
// This function does not validate cmd string.
func newCommand(cmd string) command {
	switch cmd[0] {
	case '@':
		return newACmd(cmd)
	case '(':
		return newLCmd(cmd)
	default:
		return newCCmd(cmd)
	}
}

// baseCmd is a struct that holds common fields and methods in aCmd, cCmd and lCmd.
// This struct implements command interface.
type baseCmd struct {
	// cmd is a string of a Hack command.
	cmd string
	// typ is a type of a Hack command.
	typ commandType
}

// commandType returns a type of the command.
func (c baseCmd) commandType() commandType {
	return c.typ
}

// aCmd represents an assginment command.
type aCmd struct {
	baseCmd
}

// newACmd creates a new aCmd object.
func newACmd(cmd string) aCmd {
	return aCmd{
		baseCmd{
			cmd: cmd,
			typ: aCommand,
		},
	}
}

// aCmd represents a computation command.
type cCmd struct {
	baseCmd
}

// newCCmd creates a new cCmd object.
func newCCmd(cmd string) cCmd {
	return cCmd{
		baseCmd{
			cmd: cmd,
			typ: cCommand,
		},
	}
}

// aCmd represents a label pseudo-command.
type lCmd struct {
	baseCmd
}

// newLCmd creates a new lCmd object.
func newLCmd(cmd string) lCmd {
	return lCmd{
		baseCmd{
			cmd: cmd,
			typ: lCommand,
		},
	}
}
