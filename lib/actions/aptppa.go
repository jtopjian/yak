package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/utils"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"
)

// AptPPA represents options for an apt.ppa action.
type AptPPA struct {
	BaseFields `mapstructure:",squash"`

	// Refresh will triger an apt-get update if set to true
	Refresh bool `mapstructure:"refresh" default:"true"`

	fileName string
}

// AptPPAAction will perform a full state cycle for an apt.ppa.
func AptPPAAction(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (change bool, err error) {

	var ppa AptPPA

	err = mapstructure.Decode(step.Input, &ppa)
	if err != nil {
		return
	}

	err = utils.ValidateTags(&ppa)
	if err != nil {
		return
	}

	ppa.conn = conn
	ppa.setLogger(ctx, "apt.ppa", ppa.Name, ppa.State)

	exists, err := ppa.Exists()
	if err != nil {
		return
	}

	if ppa.State == "absent" {
		if exists {
			err = ppa.Delete()
			change = true
			return
		}

		return
	}

	if !exists {
		err = ppa.Create()
		change = true
		return
	}

	return
}

// Exists will determine if an apt.ppa exists.
func (r AptPPA) Exists() (bool, error) {
	if r.fileName == "" {
		sourceFileName, err := r.sourceFileName()
		if err != nil {
			return false, err
		}

		r.fileName = "/etc/apt/sources.list.d/" + sourceFileName
		r.logDebug("ppa file: %s", r.fileName)
	}

	eo := ExecOptions{
		Command: fmt.Sprintf(`stat "%s"`, r.fileName),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("checking if installed")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		return false, fmt.Errorf("unable to check status of apt.ppa %s: %s", r.Name, err)
	}

	if rr.ExitCode == 0 {
		r.logInfo("installed")
		return true, err
	}

	r.logInfo("not installed")
	return false, nil
}

// Create will create a ppa.
func (r AptPPA) Create() error {
	eo := ExecOptions{
		Command: fmt.Sprintf("apt-add-repository -y ppa:%s", r.Name),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logInfo("adding")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to add apt.ppa %s: %s", r.Name, err)
	}

	if rr.ExitCode != 0 {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to add apt.ppa %s: %s", r.Name, err)
	}

	if r.Refresh {
		eo.Command = "apt-get update -qq"
		rr, err = exec(r.ctx, r.conn, eo)
		if err != nil {
			r.logDebug(rr.Stderr)
			return fmt.Errorf("unable to add apt.ppa %s: %s", r.Name, err)
		}
	}

	return nil
}

// Delete will delete a ppa.
func (r AptPPA) Delete() error {
	if r.fileName == "" {
		sourceFileName, err := r.sourceFileName()
		if err != nil {
			return err
		}

		r.fileName = "/etc/apt/sources.list.d/" + sourceFileName
		r.logDebug("ppa file: %s", r.fileName)
	}

	eo := ExecOptions{
		Command: fmt.Sprintf("apt-add-repository -y -r ppa:%s", r.Name),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logInfo("deleting")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.ppa %s: %s", r.Name, err)
	}

	if rr.ExitCode != 0 {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.ppa %s: %s", r.Name, err)
	}

	eo.Command = fmt.Sprintf("rm %s", r.fileName)
	r.logDebug("running command: %s", eo.Command)
	rr, err = exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.ppa %s: %s", r.Name, err)
	}

	if rr.ExitCode != 0 {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete apt.ppa %s: %s", r.Name, err)
	}

	if r.Refresh {
		eo.Command = "apt-get update -qq"
		rr, err = exec(r.ctx, r.conn, eo)
		if err != nil {
			r.logDebug(rr.Stderr)
			return fmt.Errorf("unable to delete apt.ppa %s: %s", r.Name, err)
		}
	}

	r.logInfo("deleted")

	return nil
}

func (r AptPPA) sourceFileName() (string, error) {
	name := r.Name

	lsbInfo, err := GetLSBInfo(r.BaseFields)
	if err != nil {
		return "", nil
	}

	distro := fmt.Sprintf("-%s-", strings.ToLower(lsbInfo.DistributionID))
	release := strings.ToLower(lsbInfo.Codename)

	name = strings.Replace(name, "/", distro, -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, ".", "_", -1)

	name = fmt.Sprintf("%s-%s.list", name, release)

	return name, nil
}
