package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tucats/gopackages/app-cli/app"
	"github.com/tucats/gopackages/app-cli/cli"
)

// BuildVersion is the incremental build version that is
// injected into the version number string by the build
// script
var BuildVersion = "0"

// EgoGrammar handles the command line options
var EgoGrammar = []cli.Option{
	cli.Option{
		LongName:           "path",
		Description:        "Print the default ego path",
		OptionType:         cli.Subcommand,
		Action:             PathAction,
		ParametersExpected: 0,
	},
	cli.Option{
		LongName:             "run",
		Description:          "Run an existing program",
		OptionType:           cli.Subcommand,
		Action:               RunAction,
		Value:                RunGrammar,
		ParametersExpected:   -99,
		ParameterDescription: "file-name",
	},
	cli.Option{
		LongName:    "server",
		Description: "Accept REST calls",
		OptionType:  cli.Subcommand,
		Action:      Server,
		Value:       ServerGrammar,
	},
	cli.Option{
		LongName:             "test",
		Description:          "Run a test suite",
		OptionType:           cli.Subcommand,
		Action:               TestAction,
		ParametersExpected:   -99,
		ParameterDescription: "file or path",
	},
}

// ServerGrammar handles command line options for the server subcommand
var ServerGrammar = []cli.Option{
	cli.Option{
		LongName:    "port",
		ShortName:   "p",
		OptionType:  cli.IntType,
		Description: "Specify port number to listen on",
	},
	cli.Option{
		LongName:    "not-secure",
		ShortName:   "k",
		OptionType:  cli.BooleanType,
		Description: "If set, use HTTP instead of HTTPS",
	},
	cli.Option{
		LongName:    "trace",
		ShortName:   "t",
		Description: "Display trace of bytecode execution",
		OptionType:  cli.BooleanType,
	},
}

// RunGrammar handles the command line options
var RunGrammar = []cli.Option{
	cli.Option{
		LongName:    "disassemble",
		ShortName:   "d",
		Description: "Display a disassembly of the bytecode before execution",
		OptionType:  cli.BooleanType,
	},
	cli.Option{
		LongName:    "trace",
		ShortName:   "t",
		Description: "Display trace of bytecode execution",
		OptionType:  cli.BooleanType,
	},
	cli.Option{
		LongName:    "symbols",
		ShortName:   "s",
		Description: "Display symbol table",
		OptionType:  cli.BooleanType,
	},
	cli.Option{
		LongName:    "auto-import",
		Description: "Override auto-import profile setting",
		OptionType:  cli.BooleanValueType,
	},
	cli.Option{
		LongName:    "environment",
		ShortName:   "e",
		Description: "Automatically add environment vars as symbols",
		OptionType:  cli.BooleanType,
	},
	cli.Option{
		LongName:    "source-tracing",
		ShortName:   "x",
		Description: "Print source lines as they are executed",
		OptionType:  cli.BooleanType,
	},
}

func main() {
	app := app.New("ego: execute code in the ego language")

	// Use the build number from the externally-generated build processor.
	buildVer, _ := strconv.Atoi(BuildVersion)
	app.SetVersion(1, 0, buildVer)
	app.SetCopyright("(C) Copyright Tom Cole 2020")

	// fF there aren't any arguments, default to "run"
	args := os.Args
	if len(args) == 1 {
		args = append(args, "run")
	}
	err := app.Run(EgoGrammar, args)

	// If something went wrong, report it to the user and force an exit
	// status from the error, else a default General error.
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		if e2, ok := err.(cli.ExitError); ok {
			os.Exit(e2.ExitStatus)
		}
		os.Exit(1)
	}
}
