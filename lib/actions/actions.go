package actions

import (
	"context"
	"fmt"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"
)

func RunStep(ctx context.Context, conn connections.Connection, step yakfile.Step) (bool, error) {
	action := step.Action

	switch action {
	// Core Actions
	case "exec":
		rr, err := Exec(ctx, conn, step)
		if rr.ExitCode != 0 {
			err = fmt.Errorf(rr.Stderr)
		}
		return rr.Applied, err

	case "delete-file":
		fr, err := FileDelete(ctx, conn, step)
		return fr.Applied, err

	case "download-file":
		fr, err := FileDownload(ctx, conn, step)
		return fr.Applied, err

	case "upload-file":
		fr, err := FileUpload(ctx, conn, step)
		return fr.Applied, err

	// Compound Actions
	case "apt.key":
		return AptKeyAction(ctx, conn, step)

	case "apt.ppa":
		return AptPPAAction(ctx, conn, step)

	case "apt.pkg":
		return AptPkgAction(ctx, conn, step)

	case "apt.source":
		return AptSourceAction(ctx, conn, step)

	case "cron.entry":
		return CronEntryAction(ctx, conn, step)

	default:
		return false, fmt.Errorf("action %s not supported", action)
	}
}
