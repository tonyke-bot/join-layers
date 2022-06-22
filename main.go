package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"join-layers/command/generate"
	"join-layers/command/initialize"
	"os"
)

func main() {
	parser := argparse.NewParser("join-layers", "")

	commandSet := []*struct {
		setupFunc func(*argparse.Parser) *argparse.Command
		execFunc  func()
		cmd       *argparse.Command
	}{
		{setupFunc: initialize.Setup, execFunc: initialize.Exec},
		{setupFunc: generate.Setup, execFunc: generate.Exec},
	}

	for _, command := range commandSet {
		command.cmd = command.setupFunc(parser)
	}

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
	}

	for _, command := range commandSet {
		if command.cmd.Happened() {
			command.execFunc()
		}
	}
}
