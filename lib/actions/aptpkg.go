package actions

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/utils"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"
)

// AptPkg represents options for an apt.pkg action.
type AptPkg struct {
	BaseFields `mapstructure:",squash"`
}

// AptPkgAction will perform a full state cycle for an apt.pkg.
func AptPkgAction(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (change bool, err error) {

	var pkg AptPkg

	err = mapstructure.Decode(step.Input, &pkg)
	if err != nil {
		return
	}

	err = utils.ValidateTags(&pkg)
	if err != nil {
		return
	}

	pkg.conn = conn
	pkg.setLogger(ctx, "apt.pkg", pkg.Name, pkg.State)

	exists, err := pkg.Exists()
	if err != nil {
		return
	}

	if pkg.State == "absent" {
		if exists {
			err = pkg.Delete()
			change = true
			return
		}

		return
	}

	if !exists || pkg.State == "latest" {
		err = pkg.Create()
		change = true
		return
	}

	return
}

// Exists will determine if an apt.pkg exists.
func (r AptPkg) Exists() (bool, error) {
	eo := ExecOptions{
		Command: fmt.Sprintf("apt-cache policy %s", r.Name),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("checking if installed")

	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return false, fmt.Errorf("unable to check status of apt.pkg %s: %s", r.Name, err)
	}

	if rr.Stdout == "" {
		r.logInfo("not installed")
		return false, nil
	}

	installedVersion, _ := aptPkgParseAptCache(rr.Stdout)

	switch r.State {
	case "present", "absent", "":
		switch installedVersion {
		case "(none)":
			r.logInfo("not installed")
			return false, nil
		default:
			r.logInfo("installed")
			return true, nil
		}
	}

	if r.State != installedVersion {
		r.logInfo("will be installed")
		return false, nil
	}

	r.logInfo("installed")
	return true, nil
}

func (r AptPkg) Create() error {
	eo := ExecOptions{
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	eo.Env = []string{
		"DEBIAN_FRONTEND=noninteractive",
		"APT_LISTBUGS_FRONTEND=none",
		"APT_LISTCHANGES_FRONTEND=none",
	}

	var createArgs string
	if r.State != "present" && r.State != "latest" {
		createArgs = fmt.Sprintf("%s=%s", r.Name, r.State)
	} else {
		createArgs = r.Name
	}

	eo.Command = fmt.Sprintf(
		"apt-get install -y --allow-downgrades --allow-remove-essential "+
			"--allow-change-held-packages -o DPkg::Options::=--force-confold %s",
		createArgs)

	r.logInfo("installing")
	r.logDebug("running command: %s", eo.Command)

	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to install apt.pkg %s: %s", r.Name, err)
	}

	r.logInfo("installed")
	return nil
}

func (r AptPkg) Delete() error {
	eo := ExecOptions{
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	eo.Env = []string{
		"DEBIAN_FRONTEND=noninteractive",
		"APT_LISTBUGS_FRONTEND=none",
		"APT_LISTCHANGES_FRONTEND=none",
	}

	eo.Command = fmt.Sprintf("apt-get purge -q -y %s", r.Name)

	r.logInfo("removing")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to remove apt.pkg %s: %s", r.Name, err)
	}
	r.logDebug(rr.Stderr)

	r.logInfo("removed")
	return nil
}

// apkgPkgParseAptCache is an internal function that will parse the
// output of apt-cache policy and return the version information.
func aptPkgParseAptCache(stdout string) (installed, candidate string) {
	installedRe := regexp.MustCompile("Installed: (.+)\n")
	candidateRe := regexp.MustCompile("Candidate: (.+)\n")

	if v := installedRe.FindStringSubmatch(stdout); len(v) > 1 {
		installed = v[1]
	}

	if v := candidateRe.FindStringSubmatch(stdout); len(v) > 1 {
		candidate = v[1]
	}

	return
}
