package main

import (
	"os"

	"github.com/urfave/cli"
)

var (
	debug bool
)

func before(ctx *cli.Context) error {
	if ctx.GlobalBool("debug") {
		debug = true
	}

	if ctx.Bool("debug") {
		debug = true
	}

	if ctx.GlobalString("config") != "" {
		os.Setenv("YAK_CONFIG_FILE", ctx.GlobalString("config"))
	}

	if ctx.String("config") != "" {
		os.Setenv("YAK_CONFIG_FILE", ctx.String("config"))
	}

	return nil
}
