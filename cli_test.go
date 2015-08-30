package cli_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phylake/go-cli"
)

//
// testCmd implement cli.Command
//

type testCmd struct {
	commandName string
	shortHelp   string
	longHelp    string
	execute     bool
	subCommands []cli.Command

	executeCalled bool
	executeArgs   []string
	executeStdin  *os.File
}

func (c *testCmd) Name() string {
	return c.commandName
}

func (c *testCmd) ShortHelp() string {
	return c.shortHelp
}

func (c *testCmd) LongHelp() string {
	return c.longHelp
}

func (c *testCmd) Execute(args []string, stdin *os.File) bool {
	c.executeCalled = true
	c.executeArgs = args
	c.executeStdin = stdin
	return c.execute
}

func (c *testCmd) SubCommands() []cli.Command {
	return c.subCommands
}

func newDriver(args []string) (*cli.Driver, *bytes.Buffer) {
	var stdout bytes.Buffer

	stdin, err := ioutil.TempFile("", "go-cli")
	Expect(err).To(BeNil())

	// program name is ARGV[0]
	args = append([]string{"go-cli"}, args...)

	d := cli.NewWithEnv(args, stdin, &stdout)
	return d, &stdout
}

//
// BEGIN tests
//

var _ = Describe("CLI", func() {

	It("executes a non-root command", func() {
		var err error

		d, stdout := newDriver([]string{"foo"})

		cmd := &testCmd{
			commandName: "foo",
		}
		err = d.RegisterRoot(&testCmd{
			subCommands: []cli.Command{cmd},
		})
		Expect(err).To(BeNil())

		err = d.ParseInput()
		Expect(err).To(BeNil())

		fmt.Fprintln(GinkgoWriter, stdout.String())
		Expect(cmd.executeCalled).To(BeTrue())
	})

	It("passes remaining args to Execute", func() {
		var err error

		d, _ := newDriver([]string{"foo", "arg1", "arg2"})

		cmd := &testCmd{
			commandName: "foo",
		}
		err = d.RegisterRoot(&testCmd{
			subCommands: []cli.Command{cmd},
		})
		Expect(err).To(BeNil())

		err = d.ParseInput()
		Expect(err).To(BeNil())

		Expect(cmd.executeCalled).To(BeTrue())
		Expect(cmd.executeArgs).To(Equal([]string{"arg1", "arg2"}))
	})

	It("scopes name collision to subcommands", func() {
		var err error

		goodTree := &testCmd{
			subCommands: []cli.Command{
				&testCmd{
					commandName: "foo",
					subCommands: []cli.Command{
						&testCmd{
							commandName: "bar",
						},
					},
				},
				&testCmd{
					commandName: "bar",
				},
			},
		}

		badTree := &testCmd{
			subCommands: []cli.Command{
				&testCmd{
					commandName: "foo",
				},
				&testCmd{
					commandName: "foo",
				},
			},
		}

		d, _ := newDriver(nil)

		err = d.RegisterRoot(goodTree)
		Expect(err).To(BeNil())

		err = d.ParseInput()
		Expect(err).To(BeNil())

		d, _ = newDriver(nil)

		err = d.RegisterRoot(badTree)
		Expect(err).ToNot(BeNil())
	})

	Context("formatting", func() {

		It("trims newlines out of ShortHelp()", func() {

			d, stdout := newDriver(nil)

			d.RegisterRoot(&testCmd{
				longHelp: "program description",
				subCommands: []cli.Command{
					&testCmd{
						commandName: "foo",
						shortHelp:   "short\n help",
					},
				},
			})

			d.ParseInput()

			// be careful of whitespace in this string
			expected := `program description

Commands:
    foo - short help
`
			Expect(stdout.String()).To(Equal(expected))
		})

		It("pads the longest command name", func() {

			d, stdout := newDriver(nil)

			d.RegisterRoot(&testCmd{
				longHelp: "program description",
				subCommands: []cli.Command{
					&testCmd{
						commandName: "foo",
						shortHelp:   "short help",
					},

					&testCmd{
						commandName: "longerFoo",
						shortHelp:   "short help",
					},
				},
			})

			d.ParseInput()

			// be careful of whitespace in this string
			expected := `program description

Commands:
    foo       - short help
    longerFoo - short help
`
			Expect(stdout.String()).To(Equal(expected))
		})

		It("pads the longest command name per parent node", func() {

			root := &testCmd{
				longHelp: "program description",
				subCommands: []cli.Command{
					&testCmd{
						commandName: "foo",
						shortHelp:   "short help",
					},

					&testCmd{
						commandName: "longerFoo",
						shortHelp:   "short help",
						subCommands: []cli.Command{
							&testCmd{
								commandName: "evenLongerCommandName",
								shortHelp:   "evenLongerCommandName short help",
							},
						},
					},
				},
			}

			d, stdout := newDriver(nil)
			d.RegisterRoot(root)
			d.ParseInput()

			// be careful of whitespace in this string
			expected := `program description

Commands:
    foo       - short help
    longerFoo - short help
`
			Expect(stdout.String()).To(Equal(expected))

			d, stdout = newDriver([]string{"longerFoo"})
			d.RegisterRoot(root)
			d.ParseInput()

			// be careful of whitespace in this string
			expected = `

Commands:
    evenLongerCommandName - evenLongerCommandName short help
`
			Expect(stdout.String()).To(Equal(expected))
		})
	})
})
