package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kyle221b/stc-consistent-hashing/pkg/ring"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	ctx := &Context{
		Handlers: NewHandlers(),
	}
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		cmd := parseCommand(line)
		ctx.Command = cmd
		if err := processCommand(cmd, ctx); err != nil {
			fmt.Println("Failed to process command", err)
			os.Exit(1)
		}
	}
}

func parseCommand(line string) Command {
	parts := strings.Fields(line)
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}
	return Command{
		Name: parts[0],
		Args: args,
	}
}

type Context struct {
	Handlers Handlers
	Command  Command
	Keys     []string
	Ring     *ring.Ring
}

type Command struct {
	Name string
	Args []string
}

type Handler func(ctx *Context, args []string) error

func processCommand(cmd Command, ctx *Context) error {
	switch cmd.Name {
	case "KEYS":
		return ctx.Handlers.Keys(ctx, cmd.Args)
	case "BEFORE":
		return ctx.Handlers.Before(ctx, cmd.Args)
	case "AFTER":
		return ctx.Handlers.After(ctx, cmd.Args)
	default:
		return fmt.Errorf("Unrecognized command: \"%s\"", cmd.Name)
	}
}

func NewHandlers() Handlers {
	return Handlers{
		Keys:   handleKeys,
		Before: handleBefore,
		After:  handleAfter,
	}
}

type Handlers struct {
	Keys   Handler
	Before Handler
	After  Handler
}

func NewInvalidCommandArgsError(cmd Command, msg string) InvalidCommandArgsError {
	return InvalidCommandArgsError{
		cmd:     cmd.Name,
		message: msg,
	}
}

type InvalidCommandArgsError struct {
	cmd     string
	message string
}

func (e InvalidCommandArgsError) Error() string {
	return fmt.Sprintf("Invalid command args for command '%s': %s", e.cmd, e.message)
}

func handleKeys(ctx *Context, args []string) error {
	ring := ring.NewRing(args, 1)
	ctx.Keys = args
	ctx.Ring = ring
	return nil
}

func handleBefore(ctx *Context, args []string) error {
	n, err := validateReshardArg(ctx, args)
	if err != nil {
		return err
	}
	ctx.Ring.Reshard(uint32(n))
	printMapping(ctx)
	return nil
}

func handleAfter(ctx *Context, args []string) error {
	n, err := validateReshardArg(ctx, args)
	if err != nil {
		return err
	}
	numChanged := ctx.Ring.Reshard(uint32(n))
	printMapping(ctx)
	fmt.Printf("moved=%d\n", numChanged)
	return nil
}

func validateReshardArg(ctx *Context, args []string) (uint32, error) {
	if ctx.Ring == nil {
		return 0, errors.New("Must establish keyset before performing a hashing command")
	}
	if len(args) != 1 {
		return 0, NewInvalidCommandArgsError(ctx.Command, "Only one arg allowed")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n <= 0 {
		return 0, NewInvalidCommandArgsError(ctx.Command, "Arg must be positive int value")
	}
	return uint32(n), nil
}

func printMapping(ctx *Context) {
	for _, key := range ctx.Keys {
		shard := ctx.Ring.GetShard(key)
		fmt.Printf("%s: %d\n", key, shard)
	}
}
