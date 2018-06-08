package actions

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/utils"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"
)

// AptSource represents options for an apt.source action.
type AptSource struct {
	BaseFields `mapstructure:",squash"`

	URI          string `mapstructure:"uri" required:"true"`
	Distribution string `mapstructure:"distribution" required:"true"`
	Component    string `mapstructure:"component"`
	IncludeSrc   bool   `mapstructure:"include_src"`
	Refresh      bool   `mapstructure:"refresh" default:"true"`
}

// AptSourceAction will perform a full state cycle for an apt.source.
func AptSourceAction(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (change bool, err error) {

	var as AptSource

	err = mapstructure.Decode(step.Input, &as)
	if err != nil {
		return
	}

	err = utils.ValidateTags(&as)
	if err != nil {
		return
	}

	as.conn = conn
	as.setLogger(ctx, "apt.source", as.Name, as.State)

	exists, err := as.Exists()
	if err != nil {
		return
	}

	if as.State == "absent" {
		if exists {
			err = as.Delete()
			change = true
			return
		}

		return
	}

	if !exists {
		err = as.Create()
		change = true
		return
	}

	return
}

// Exists will determine if an apt.source exists.
func (r AptSource) Exists() (bool, error) {
	path := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", r.Name)
	entry := fmt.Sprintf("deb %s %s %s", r.URI, r.Distribution, r.Component)
	srcEntry := fmt.Sprintf("deb-src %s %s %s", r.URI, r.Distribution, r.Component)

	r.logDebug("checking if %s exists", path)

	eo := ExecOptions{
		Command: fmt.Sprintf(`cat "%s"`, path),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		return false, fmt.Errorf("unable to check status of apt.source %s: %s", r.Name, err)
	}

	if rr.ExitCode != 0 {
		r.logInfo("not installed")
		return false, nil
	}

	var exists bool
	var srcExists bool
	for _, line := range strings.Split(rr.Stdout, "\n") {
		if line == entry {
			exists = true
		}

		if line == srcEntry {
			srcExists = true
		}
	}

	if exists {
		if r.IncludeSrc && !srcExists {
			r.logInfo("not installed")
			return false, nil
		}

		r.logInfo("exists")
		return true, nil
	}

	r.logInfo("not installed")
	return false, nil
}

// Create will create an apt.source file.
func (r AptSource) Create() error {
	path := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", r.Name)
	entry := fmt.Sprintf("deb %s %s %s", r.URI, r.Distribution, r.Component)
	srcEntry := fmt.Sprintf("\ndeb-src %s %s %s", r.URI, r.Distribution, r.Component)

	r.logInfo("adding")

	tmpfile, err := ioutil.TempFile("/tmp", "apt.source")
	if err != nil {
		return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(entry)); err != nil {
		return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
	}

	if r.IncludeSrc {
		if _, err := tmpfile.Write([]byte(srcEntry)); err != nil {
			return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
		}
	}

	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
	}

	cfo := CopyFileOptions{
		Source:      tmpfile.Name(),
		Destination: tmpfile.Name(),
	}

	eo := ExecOptions{
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	if _, err := fileUploadAndMove(r.ctx, r.conn, cfo, eo, path); err != nil {
		return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
	}

	if r.Refresh {
		eo.Command = "apt-get update -qq"
		rr, err := exec(r.ctx, r.conn, eo)
		if err != nil {
			r.logDebug(rr.Stderr)
			return fmt.Errorf("unable to add apt.source %s: %s", r.Name, err)
		}
	}

	return nil
}

// Delete will delete an apt.source file.
func (r AptSource) Delete() error {
	path := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", r.Name)
	eo := ExecOptions{
		Command: fmt.Sprintf(`rm "%s"`, path),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logInfo("deleting")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.source %s: %s", r.Name, err)
	}

	if rr.ExitCode != 0 {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.source %s: %s", r.Name, err)
	}

	if r.Refresh {
		eo.Command = "apt-get update -qq"
		rr, err = exec(r.ctx, r.conn, eo)
		if err != nil {
			r.logDebug(rr.Stderr)
			return fmt.Errorf("unable to delete apt.source %s: %s", r.Name, err)
		}
	}

	return nil
}
