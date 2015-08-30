package cli

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
)

type (
	Command interface {
		// The name of the command
		Name() string

		// A one-line description of this command
		ShortHelp() string

		// A multi-line description of this command.
		//
		// Its subcommands' ShortHelp message will also be printed.
		LongHelp() string

		// Execute executes with the remaining passed in arguments and os.Stdin
		//
		// Return false if the command can't execute which will display the
		// command's LongHelp message
		Execute([]string, *os.File) bool

		// Any sub commands this command is capable of
		SubCommands() []Command
	}

	Driver struct {
		// os.Args
		args []string

		// stdin is passed unaltered to commands since we can't
		// make assumptions about the minimal interface
		stdin *os.File

		// to communicate out we only need a writer so there's no need to couple
		// simple communication with a *os.File
		stdout io.Writer

		/* command-related fields */
		root              Command
		registry          map[string]Command
		longestSubCommand map[string]float64
	}
)

var newlineRE = regexp.MustCompile(`\n`)

func New() *Driver {
	return NewWithEnv(nil, nil, nil)
}

// NewWithEnv inverts control of the outside world and enables testing
func NewWithEnv(args []string, stdin *os.File, stdout io.Writer) *Driver {
	if args == nil {
		args = os.Args
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	if stdout == nil {
		stdout = os.Stdout
	}

	return &Driver{
		args:   args,
		stdin:  stdin,
		stdout: stdout,
	}
}

func (d *Driver) ParseInput() error {
	if d.root == nil {
		return errors.New("root command doesn't exist. call RegisterRoot first")
	}

	cmd := d.root
	i := 1 // 0 is the program name (similar to ARGV)
	for ; i < len(d.args); i++ {

		// fmt.Fprintf(d.stdout, "arg %d %s\n", i, d.args[i])
		if subCmd, exists := d.registry[d.args[i]]; exists {
			cmd = subCmd
		} else {
			break
		}
	}

	if !cmd.Execute(d.args[i:], d.stdin) {

		fmt.Fprintln(d.stdout, cmd.LongHelp())
		fmt.Fprintln(d.stdout)

		padding, _ := d.longestSubCommand[cmd.Name()]

		subCmds := cmd.SubCommands()
		if len(subCmds) > 0 {

			fmt.Fprintln(d.stdout, "Commands:")

			for _, subCmd := range subCmds {
				cmdName := subCmd.Name()

				// create format string with correct padding to accommodate
				// the longest command name.
				//
				// e.g. "    %-42s - %s\n" if 42 is the longest
				fmtStr := fmt.Sprintf("    %%-%.fs - %%s\n", padding)
				shortHelp := newlineRE.ReplaceAllString(subCmd.ShortHelp(), "")
				fmt.Fprintf(d.stdout, fmtStr, cmdName, shortHelp)
			}
		}
	}

	return nil
}

func (d *Driver) RegisterRoot(newRoot Command) error {
	if d.root != nil {
		return errors.New("root command already registered")
	}

	if newRoot.Name() != "" {
		return errors.New("root command name must be \"\"")
	}

	d.registry = make(map[string]Command)
	d.longestSubCommand = make(map[string]float64)
	d.root = newRoot

	return d.registerCmd(d.root, nil)
}

func (d *Driver) registerCmd(cmd Command, maxLen *float64) error {
	if cmd == nil {
		return nil
	}

	cmdName := cmd.Name()

	if maxLen != nil {
		*maxLen = math.Max(*maxLen, float64(len(cmdName)))
	}

	if _, exists := d.registry[cmdName]; exists {
		return fmt.Errorf("command named %s already exists", cmdName)
	}

	d.registry[cmdName] = cmd

	subCmds := cmd.SubCommands()
	if subCmds != nil {

		longestSub := new(float64)

		for _, subCmd := range subCmds {

			err := d.registerCmd(subCmd, longestSub)
			if err != nil {
				return err
			}
		}

		d.longestSubCommand[cmdName] = *longestSub
	}

	return nil
}
