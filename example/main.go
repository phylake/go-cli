package main

import (
	"fmt"
	"os"

	"github.com/phylake/go-cli"
	"github.com/phylake/go-cli/cmd"
)

// Run `go run main.go`, `go run main.go punch`, etc.
func main() {
	driver := cli.New()

	rootCmd := &cmd.Root{
		Help: `Usage: ninja COMMAND [args]

A madeup CLI to demonstrate this framework`,
		SubCommandList: []cli.Command{
			&cmd.Default{
				NameStr:      "punch",
				ShortHelpStr: "punch your shell",
				LongHelpStr: `Punch your shell with the power of 1000 hurricanes.

Usage: punch [OPTIONS]

Options:
  --execute`,
				ExecuteFunc: func(args []string, stdin *os.File) bool {
					if len(args) == 1 && args[0] == "--execute" {
						fmt.Println("POW!")
						return true
					}
					return false
				},
			},
			&KickCmd{},
		},
	}

	if err := driver.RegisterRoot(rootCmd); err != nil {
		panic(err)
	}

	if err := driver.ParseInput(); err != nil {
		panic(err)
	}
}

// KickCmd implements cli.Command
type KickCmd struct{}

func (cmd *KickCmd) Name() string {
	return "kick"
}

func (cmd *KickCmd) ShortHelp() string {
	return "kick your shell"
}

func (cmd *KickCmd) LongHelp() string {
	return `kick your shell with the power of one supernova

Usage: kick [OPTIONS]

Options:
  --execute`
}

// Return false if this command wasn't correctly invoked and LongHelp() will be
// printed out
func (cmd *KickCmd) Execute(args []string, stdin *os.File) bool {
	if len(args) == 1 && args[0] == "--execute" {
		fmt.Println("BOOM!")
		return true
	}
	return false
}

func (cmd *KickCmd) SubCommands() []cli.Command {
	return nil
}
