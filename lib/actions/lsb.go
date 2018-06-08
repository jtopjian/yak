package actions

import (
	"fmt"
	"regexp"
)

type LSBInfo struct {
	DistributionID string
	Description    string
	Release        string
	Codename       string
}

func GetLSBInfo(b BaseFields) (*LSBInfo, error) {
	var lsbInfo LSBInfo

	distributorRe := regexp.MustCompile("Distributor ID:\\s+(.+)\n")
	descriptionRe := regexp.MustCompile("Description:\\s+(.+)\n")
	releaseRe := regexp.MustCompile("Release:\\s+(.+)\n")
	codenameRe := regexp.MustCompile("Codename:\\s+(.+)")

	eo := ExecOptions{
		Command: "/usr/bin/lsb_release -a",
		Sudo:    b.Sudo,
		Timeout: b.Timeout,
	}

	b.logDebug(fmt.Sprintf("running command: %s", eo.Command))
	rr, err := exec(b.ctx, b.conn, eo)
	if err != nil {
		b.logDebug(rr.Stderr)
		return nil, fmt.Errorf("unable to run lsb_info: %s", err)
	}

	if v := distributorRe.FindStringSubmatch(rr.Stdout); len(v) > 1 {
		lsbInfo.DistributionID = v[1]
	}

	if v := descriptionRe.FindStringSubmatch(rr.Stdout); len(v) > 1 {
		lsbInfo.Description = v[1]
	}

	if v := releaseRe.FindStringSubmatch(rr.Stdout); len(v) > 1 {
		lsbInfo.Release = v[1]
	}

	if v := codenameRe.FindStringSubmatch(rr.Stdout); len(v) > 1 {
		lsbInfo.Codename = v[1]
	}

	return &lsbInfo, nil
}
