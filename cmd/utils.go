package main

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"github.com/fatih/color"

	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"

	"github.com/jtopjian/yak/lib/actions"
	"github.com/jtopjian/yak/lib/yakfile"
)

var (
	title   = color.New(color.FgYellow).Add(color.Underline)
	cyan    = color.New(color.FgCyan)
	blue    = color.New(color.FgBlue)
	magenta = color.New(color.FgMagenta)
)

// newHerd will build a herd given a directory.
func newHerd(c *cli.Context) (yakfile.Herd, error) {
	log := getLogger()

	v := path.Join(c.String("dir"), "*.yaml")

	files, err := filepath.Glob(v)
	if err != nil {
		return nil, err
	}

	log.Debugf("yak files: %v", files)

	return yakfile.NewHerd(files)
}

// getLogger is a convenience function to create and return a logger.
func getLogger() *logrus.Logger {
	log := logrus.New()
	//log.Formatter = new(TextFormatter)
	if debug {
		log.Level = logrus.DebugLevel
	}

	return log
}

// getTask is a convenience function to look up the specified task.
func getTask(c *cli.Context) (string, error) {
	if c.NArg() == 0 {
		return "", fmt.Errorf("no task specified")
	}

	return c.Args()[0], nil
}

// runStep is a convenience function to run a step or notify.
func runStep(ctx context.Context, host yakfile.Host, step yakfile.Step) (bool, error) {
	log := ctx.Value("log").(*logrus.Logger)

	log.Debugf("attempting to connect to %s via %s",
		host.Name, host.ConnectionType)

	if err := host.Connection.Connect(); err != nil {
		return false, err
	}

	l := log.WithFields(logrus.Fields{
		"host": host.Name,
	})
	ctx = context.WithValue(context.Background(), "log", l)
	changed, err := actions.RunStep(ctx, host.Connection, step)
	if err != nil {
		log.Error(err)
	}

	return changed, err
}
