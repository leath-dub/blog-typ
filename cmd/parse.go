package cmd

import (
	"errors"
	"blog-typ/lib"
)

type CmdResult interface{}

type Cmd interface {
	Run(args ...string) (CmdResult, error)	
}

func Parse(args []string, config lib.Config) (Cmd, error) {
	if len(args) < 2 {
		return nil, errors.New("Not enough arguments")
	}

	switch args[1] {
	case "build":
		return NewBuildCmd(config.Posts, config.Build)
	case "run":
		return NewRunCmd(config.Build), nil
	}

	return nil, errors.New("Unrecognized command")
}
