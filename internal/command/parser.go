package command

import (
	"fmt"
	"strings"

	"github.com/mmnalaka/medis/internal/resp"
)

type Command struct {
	Name string
	Args [][]byte
}

// ParseCommand converts a RESP array into a Command structure
func ParseCommand(data resp.RESPData) (*Command, error) {
	array, ok := data.(*resp.Array)
	if !ok {
		return nil, fmt.Errorf("command must be a RESP array")
	}

	if len(array.Data) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// First element should be the command name as a bulk string
	cmdNameBulk, ok := array.Data[0].(*resp.BulkString)
	if !ok {
		return nil, fmt.Errorf("command name must be a bulk string")
	}

	cmd := &Command{
		Name: strings.ToUpper(string(cmdNameBulk.Data)), // Commands are case-insensitive
		Args: make([][]byte, len(array.Data)-1),
	}

	// Extract arguments
	for i := 1; i < len(array.Data); i++ {
		arg, ok := array.Data[i].(*resp.BulkString)
		if !ok {
			return nil, fmt.Errorf("command argument must be a bulk string")
		}
		cmd.Args[i-1] = arg.Data
	}

	return cmd, nil
}

// String returns a human-readable representation of the command
func (c *Command) String() string {
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		args[i] = string(arg)
	}
	return fmt.Sprintf("%s %s", c.Name, strings.Join(args, " "))
}
