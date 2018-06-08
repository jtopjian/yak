package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"

	"github.com/sirupsen/logrus"
)

// CopyFileOptions represents options for a copy action.
type CopyFileOptions struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
	UID         int    `mapstructure:"uid"`
	GID         int    `mapstructure:"gid"`
	Mode        int    `mapstructure:"mode"`
	Timeout     int    `mapstructure:"timeout"`

	ContextLogger
}

// FileOptions represents options for general file actions.
type FileOptions struct {
	Path    string `mapstructure:"path"`
	Timeout int    `mapstructure:"timeout"`

	ContextLogger
}

func fileCopy(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
	action string,
) (*connections.FileResult, error) {

	var cfo CopyFileOptions

	if log, ok := ctx.Value("log").(*logrus.Entry); ok {
		log = log.WithFields(logrus.Fields{
			"action": fmt.Sprintf("file-%s", action),
		})
		cfo.ctx = context.WithValue(ctx, "log", log)
	}

	err := mapstructure.Decode(step.Input, &cfo)
	if err != nil {
		return nil, err
	}

	if cfo.Source == "" {
		return nil, fmt.Errorf("source is required for file uploads")
	}

	if cfo.Destination == "" {
		return nil, fmt.Errorf("destination is required for file uploads")
	}

	cfo.logInfo(fmt.Sprintf("uploading %s to %s", cfo.Source, cfo.Destination))

	cCFO := connections.CopyFileOptions{
		Source:      cfo.Source,
		Destination: cfo.Destination,
		UID:         cfo.UID,
		GID:         cfo.GID,
		Mode:        os.FileMode(cfo.Mode),
		Timeout:     cfo.Timeout,
	}

	switch action {
	case "upload":
		return conn.FileUpload(cCFO)
	case "download":
		return conn.FileDownload(cCFO)
	}

	return nil, nil
}

// FileUpload will upload a file to a target host.
func FileUpload(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (*connections.FileResult, error) {

	return fileCopy(ctx, conn, step, "upload")
}

// fileUpload will upload a file to a target host.
// It is meant to be used internally by other resources.
// It builds an ad-hoc step and passes it to FileUpload.
func fileUpload(
	ctx context.Context,
	conn connections.Connection,
	cfo CopyFileOptions,
) (*connections.FileResult, error) {

	step := yakfile.Step{
		Action: "file-upload",
		Name:   "file-upload",
		Input: map[string]interface{}{
			"source":      cfo.Source,
			"destination": cfo.Destination,
			"uid":         cfo.UID,
			"gid":         cfo.GID,
			"mode":        cfo.Mode,
			"timeout":     cfo.Timeout,
		},
	}

	return FileUpload(ctx, conn, step)
}

// fileUploadAndMove is a convenience function to upload a
// file to a target host and then move it to a secondary
// location. This is useful for cases when `sudo` is required
// to place the file in the final location.
func fileUploadAndMove(
	ctx context.Context,
	conn connections.Connection,
	cfo CopyFileOptions,
	eo ExecOptions,
	finalDestination string,
) (*connections.FileResult, error) {

	fileTask := yakfile.Step{
		Action: "file-upload",
		Name:   "file-upload",
		Input: map[string]interface{}{
			"source":      cfo.Source,
			"destination": cfo.Destination,
			"uid":         cfo.UID,
			"gid":         cfo.GID,
			"mode":        cfo.Mode,
			"timeout":     cfo.Timeout,
		},
	}

	fr, err := FileUpload(ctx, conn, fileTask)
	if err != nil {
		return nil, err
	}

	if !fr.Success {
		return nil, fmt.Errorf("unable to upload file to %s", cfo.Destination)
	}

	eo.Command = fmt.Sprintf(`mv "%s" "%s"`, cfo.Destination, finalDestination)
	rr, err := exec(ctx, conn, eo)
	if err != nil {
		return nil, err
	}

	if rr.ExitCode != 0 {
		return nil, fmt.Errorf("%s", rr.Stderr)
	}

	return fr, nil
}

// FileDownload will download a file from a target host.
func FileDownload(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (*connections.FileResult, error) {

	return fileCopy(ctx, conn, step, "download")
}

// fileDownload will download a file from a target host.
// It is meant to be used internally by other resources.
// It builds an ad-hoc step and passes it to FileDownload.
func fileDownload(
	ctx context.Context,
	conn connections.Connection,
	cfo CopyFileOptions,
) (*connections.FileResult, error) {

	step := yakfile.Step{
		Action: "file-download",
		Name:   "file-download",
		Input: map[string]interface{}{
			"source":      cfo.Source,
			"destination": cfo.Destination,
			"uid":         cfo.UID,
			"gid":         cfo.GID,
			"mode":        cfo.Mode,
			"timeout":     cfo.Timeout,
		},
	}

	return FileDownload(ctx, conn, step)
}

// FileExists will determine if a file exists on a target host.
func FileExists(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (*connections.FileResult, error) {
	var fo FileOptions

	if log, ok := ctx.Value("log").(*logrus.Entry); ok {
		log = log.WithFields(logrus.Fields{
			"action": "file-exists",
		})
		fo.ctx = context.WithValue(ctx, "log", log)
	}

	err := mapstructure.Decode(step.Input, &fo)
	if err != nil {
		return nil, err
	}

	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file exists")
	}

	fo.logInfo(fmt.Sprintf("checking existence of %s", fo.Path))

	cFO := connections.FileOptions{
		Path:    fo.Path,
		Timeout: fo.Timeout,
	}

	return conn.FileInfo(cFO)
}

// fileExists will determine if a file exists on a target host.
// It is meant to be used internally by other resources.
// It builds an ad-hoc step and passes it to FileExists.
func fileExists(
	ctx context.Context,
	conn connections.Connection,
	fo FileOptions,
) (*connections.FileResult, error) {

	step := yakfile.Step{
		Action: "file-exists",
		Name:   "file-exists",
		Input: map[string]interface{}{
			"path":    fo.Path,
			"timeout": fo.Timeout,
		},
	}

	return FileExists(ctx, conn, step)
}

// FileDelete will delete a file on a target host.
func FileDelete(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (*connections.FileResult, error) {
	var fo FileOptions

	if log, ok := ctx.Value("log").(*logrus.Entry); ok {
		log = log.WithFields(logrus.Fields{
			"action": "file-delete",
		})
		fo.ctx = context.WithValue(ctx, "log", log)
	}

	err := mapstructure.Decode(step.Input, &fo)
	if err != nil {
		return nil, err
	}

	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file delete")
	}

	fo.logInfo(fmt.Sprintf("deleting %s", fo.Path))

	cFO := connections.FileOptions{
		Path:    fo.Path,
		Timeout: fo.Timeout,
	}

	return conn.FileDelete(cFO)
}

// fileDelete will delete a file from a target host.
// It is meant to be used internally by other resources.
// It builds an ad-hoc step and passes it to FileDelete.
func fileDelete(
	ctx context.Context,
	conn connections.Connection,
	fo FileOptions,
) (*connections.FileResult, error) {

	step := yakfile.Step{
		Action: "file-delete",
		Name:   "file-delete",
		Input: map[string]interface{}{
			"path":    fo.Path,
			"timeout": fo.Timeout,
		},
	}

	return FileDelete(ctx, conn, step)
}
