package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var (
	configFlag = cli.StringFlag{
		Name:   "config",
		Usage:  "yak configuration file",
		EnvVar: "YAK_CONFIG_FILE",
	}

	debugFlag = cli.BoolFlag{
		Name:   "debug,d",
		Usage:  "debug mode",
		EnvVar: "YAK_DEBUG,DEBUG",
	}

	dirFlag = cli.StringFlag{
		Name:   "dir",
		Usage:  "yakfile directory",
		EnvVar: "YAK_DIR,DIR",
		Value:  ".",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "yak"
	app.Usage = "An execution tool"
	app.Flags = []cli.Flag{
		configFlag,
		debugFlag,
		dirFlag,
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "run",
			Usage:  "run a task",
			Before: before,
			Action: actionRun,
			Flags: []cli.Flag{
				configFlag,
				debugFlag,
				dirFlag,
			},
		},

		cli.Command{
			Name:   "plan",
			Usage:  "show a yak plan",
			Before: before,
			Action: actionPlan,
			Flags: []cli.Flag{
				configFlag,
				debugFlag,
				dirFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
