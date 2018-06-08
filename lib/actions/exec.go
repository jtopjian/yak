package actions

import (
	"context"
	"fmt"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"

	"github.com/sirupsen/logrus"
)

// ExecOptions represents options for an exec action.
type ExecOptions struct {
	Command string   `mapstructure:"cmd"`
	Dir     string   `mapstructure:"dir"`
	Env     []string `mapstructure:"env"`
	Sudo    bool     `mapstructure:"sudo"`
	Timeout int      `mapstructure:"timeout"`
	Unless  string   `mapstructure:"unless"`

	ContextLogger
}

// Exec is meant to be run from a step.
// It will execute an arbitrary command.
func Exec(ctx context.Context, conn connections.Connection, step yakfile.Step) (*connections.RunResult, error) {
	var internal bool
	var eo ExecOptions

	if log, ok := ctx.Value("log").(*logrus.Entry); ok {
		log = log.WithFields(logrus.Fields{
			"action": "exec",
		})
		eo.ctx = context.WithValue(ctx, "log", log)
	}

	err := mapstructure.Decode(step.Input, &eo)
	if err != nil {
		return nil, err
	}

	if eo.Command == "" {
		return nil, fmt.Errorf("cmd is required with exec")
	}

	if _, ok := step.Input["_internal"]; ok {
		internal = true
	}

	cmd := eo.Command
	unless := eo.Unless

	if eo.Sudo {
		cmd = fmt.Sprintf(`sudo %s`, cmd)
		unless = fmt.Sprintf(`sudo %s`, unless)
	}

	if eo.Dir != "" {
		cmd = fmt.Sprintf(`cd %s && %s`, eo.Dir, cmd)
		unless = fmt.Sprintf(`cd %s && %s`, eo.Dir, unless)
	}

	for _, env := range eo.Env {
		cmd = fmt.Sprintf(`%s && %s`, env, cmd)
		unless = fmt.Sprintf(`%s && %s`, env, unless)
	}

	ro := connections.RunOptions{
		Command: cmd,
		Timeout: eo.Timeout,
	}

	if eo.Unless != "" {
		if !internal {
			eo.logInfo(fmt.Sprintf("running unless command: %s", unless))
		}
		uo := connections.RunOptions{
			Command: unless,
			Timeout: eo.Timeout,
		}

		ur, err := conn.RunCommand(uo)
		if ur.ExitCode == 0 {
			ur.Applied = false
			return ur, err
		}
	}

	if !internal {
		eo.logInfo(fmt.Sprintf("running command: %s", cmd))
	}

	return conn.RunCommand(ro)
}

// exec will execute an arbitrary command.
// It is meant to be used internally by other resources.
// It builds an ad-hoc step and passes it to Exec.
func exec(ctx context.Context, conn connections.Connection, eo ExecOptions) (*connections.RunResult, error) {
	step := yakfile.Step{
		Action: "exec",
		Name:   "exec",
		Input: map[string]interface{}{
			"cmd":       eo.Command,
			"sudo":      eo.Sudo,
			"timeout":   eo.Timeout,
			"dir":       eo.Dir,
			"env":       eo.Env,
			"unless":    eo.Unless,
			"_internal": true,
		},
	}

	return Exec(ctx, conn, step)
}
