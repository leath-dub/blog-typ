package main

import (
	"blog-typ/cmd"
	"blog-typ/lib"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

func ConfigFromFile(fpath string) (lib.Config, error) {
	var result lib.Config

	_, err := toml.DecodeFile(fpath, &result)

	return result, err
}

func main() {
	conf, err := ConfigFromFile("./config.toml")
	if err != nil {
		slog.Error("Reading Config", "error", err)
		os.Exit(1)
	}

	c, err := cmd.Parse(os.Args, conf)
	if err != nil {
		slog.Error("Parsing Command", "error", err)
		os.Exit(1)
	}

	res, err := c.Run()
	if err != nil {
		slog.Error("Running build command", "error", err)
		os.Exit(1)
	}

	if buildRes, ok := res.(*cmd.BuildCmdResult); !ok {
		slog.Error("Failed to interpret `build' command result")
		os.Exit(1)
	} else if err := buildRes.Commit(); err != nil {
		slog.Error("Commiting build result", "error", err)
		os.Exit(1)
	}
}
