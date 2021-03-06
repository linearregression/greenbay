package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/greenbay/check"
	"github.com/mongodb/greenbay/operations"
	"github.com/pkg/errors"
	"github.com/tychoish/grip"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func main() {
	// this is where the main action of the program starts. The
	// command line interface is managed by the cli package and
	// its objects/structures. This, plus the basic configuration
	// in buildApp(), is all that's necessary for bootstrapping the
	// environment.
	app := buildApp()
	err := app.Run(os.Args)
	grip.CatchEmergencyFatal(err)
}

////////////////////////////////////////////////////////////////////////
//
// Set up cli.App environment, configure logging and register sub-commands
//
////////////////////////////////////////////////////////////////////////

// we build the app outside of main so that we can test the operation
func buildApp() *cli.App {
	app := cli.NewApp()
	app.Name = "greenbay"
	app.Usage = "a system configuration integration test runner."
	app.Version = "0.0.1-pre"

	// Register sub-commands here.
	app.Commands = []cli.Command{
		list(),
		checks(),
	}

	// need to call a function in the check package so that the
	// init() methods fire.
	_ = check.NewBase("", -1)

	// These are global options. Use this to configure logging or
	// other options independent from specific sub commands.
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "level",
			Value: "info",
			Usage: "Specify lowest visible loglevel as string: 'emergency|alert|critical|error|warning|notice|info|debug'",
		},
	}

	app.Before = func(c *cli.Context) error {
		loggingSetup(app.Name, c.String("level"))
		return nil
	}

	return app
}

// logging setup is separate to make it unit testable
func loggingSetup(name, level string) {
	grip.SetName(name)
	grip.SetThreshold(level)
}

////////////////////////////////////////////////////////////////////////
//
// Define SubCommands
//
////////////////////////////////////////////////////////////////////////

func list() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "list all available checks",
		Action: func(c *cli.Context) error {
			var list []string
			for name := range registry.JobTypeNames() {
				list = append(list, name)
			}

			if len(list) == 0 {
				return errors.New("no jobs registered")
			}

			sort.Strings(list)
			fmt.Printf("Registered Greenbay Checks:\n\t%s\n",
				strings.Join(list, "\n\t"))

			grip.Infof("%d checks registered", len(list))
			return nil
		},
	}
}

func checks() cli.Command {
	defaultNumJobs := runtime.NumCPU()
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, "greenbay.yaml")

	return cli.Command{
		Name:  "run",
		Usage: "run greenbay suites",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name: "jobs",
				Usage: fmt.Sprintf("specify the number of parallel tests to run. (Default %s)",
					defaultNumJobs),
				Value: defaultNumJobs,
			},
			cli.StringFlag{
				Name: "conf",
				Usage: fmt.Sprintln("path to config file. '.json', '.yaml', and '.yml' extensions ",
					"supported.", "Default path:", configPath),
				Value: configPath,
			},
			cli.StringFlag{
				Name:  "output",
				Usage: "path of file to write output too. Defaults to *not* writing output to a file",
				Value: "",
			},
			cli.BoolFlag{
				Name:  "quiet",
				Usage: "specify to disable printed (standard output) results",
			},
			cli.StringFlag{
				Name: "format",
				Usage: fmt.Sprintln("Selects the output format, defaults to a format that mirrors gotest,",
					"but also supports evergreen's results format.",
					"Use 'gotest' (default), 'result', or 'log'."),
				Value: "gotest",
			},
			cli.StringSliceFlag{
				Name:  "test",
				Usage: "specify a check, by name. may specify multiple times",
			},
			cli.StringSliceFlag{
				Name:  "suite",
				Usage: "specify a suite or suites, by name. if not specified, runs the 'all' suite",
			},
		},
		Action: func(c *cli.Context) error {
			// Note: in the future in may make sense to
			// use this context to timeout the work of the
			// underlying processes.
			ctx := context.Background()

			suites := c.StringSlice("suite")
			tests := c.StringSlice("test")
			if len(suites) == 0 && len(tests) == 0 {
				suites = append(suites, "all")
			}

			app, err := operations.NewApp(
				c.String("conf"),
				c.String("output"),
				c.String("format"),
				c.Bool("quiet"),
				c.Int("jobs"),
				suites,
				tests)

			if err != nil {
				return errors.Wrap(err, "problem prepping to run tests")
			}

			return errors.Wrap(app.Run(ctx), "problem running tests")
		},
	}
}
