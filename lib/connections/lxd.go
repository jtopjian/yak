package connections

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jtopjian/yak/lib/config"
	"github.com/jtopjian/yak/lib/shared"

	lxd "github.com/lxc/lxd/client"
	lxd_api "github.com/lxc/lxd/shared/api"

	"github.com/mitchellh/mapstructure"
)

const (
	LXDCommandTimeout    = 60
	LXDConnectionTimeout = 300
	LXDDefaultShell      = "/bin/bash"
)

// LXD represents an LXD connection.
type LXD struct {
	AuthEntry string `mapstructure:"auth"`
	Host      string `mapstructure:"host"`
	Shell     string `mapstructure:"shell"`
	Timeout   int    `mapstructure:"timeout"`

	client lxd.ContainerServer
}

// NewLXD will return an LXD client.
func NewLXD(options map[string]interface{}) (*LXD, error) {
	var lxdConfig LXD

	err := mapstructure.Decode(options, &lxdConfig)
	if err != nil {
		return nil, err
	}

	if lxdConfig.AuthEntry == "" {
		return nil, fmt.Errorf("auth is required for an LXD connection")
	}

	if lxdConfig.Shell == "" {
		lxdConfig.Shell = LXDDefaultShell
	}

	return &lxdConfig, nil
}

// Connect implements the Connect method of the Connection interface.
// It will connect to an LXD server.
func (r *LXD) Connect() error {
	var err error
	var lxdAuth shared.LXDAuth

	// If a connection has already been made, don't do anything.
	if r.client != nil {
		return nil
	}

	yakConf, err := config.FindAndLoad()
	if err != nil {
		return err
	}

	lxda, err := yakConf.GetAuthEntry(r.AuthEntry)
	if err != nil {
		return err
	}

	err = mapstructure.Decode(lxda.Options, &lxdAuth)
	if err != nil {
		return err
	}

	connectTimeout := LXDConnectionTimeout
	if r.Timeout > 0 {
		connectTimeout = r.Timeout
	}

	err = retryFunc(connectTimeout, func() error {
		r.client, err = lxdAuth.Authenticate()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			return fmt.Errorf("timed out connecting to %s", lxdAuth.Remote)
		}
	}

	return nil
}

// RunCommand implements the Run method of the Connection interface.
func (r LXD) RunCommand(ro RunOptions) (*RunResult, error) {
	var rr RunResult
	var outBuf, errBuf bytes.Buffer

	// validate options
	if ro.Command == "" {
		return nil, fmt.Errorf("a command is required")
	}

	timeout := LXDCommandTimeout
	if ro.Timeout > 0 {
		timeout = ro.Timeout
	}

	// Set up the output
	log := ioutil.Discard
	if ro.Log != nil {
		log = *ro.Log
	}

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	args := lxd.ContainerExecArgs{
		Stdin:    ioutil.NopCloser(bytes.NewReader(nil)),
		Stderr:   errW,
		Stdout:   outW,
		DataDone: make(chan bool),
		Control:  nil,
	}

	outTee := io.TeeReader(outR, &outBuf)
	errTee := io.TeeReader(errR, &errBuf)
	outDoneCh := make(chan struct{})
	errDoneCh := make(chan struct{})
	go printOutput(log, outTee, outDoneCh)
	go printOutput(log, errTee, errDoneCh)

	//cmd := strings.Replace(ro.Command, `"`, `\"`, -1)
	req := lxd_api.ContainerExecPost{
		Command:     []string{r.Shell, "-c", ro.Command},
		WaitForWS:   true,
		Interactive: false,
	}

	err := timeoutFunc(timeout, func() error {
		op, err := r.client.ExecContainer(r.Host, req, &args)
		if err != nil {
			return err
		}

		// Wait for completion.
		if err := op.Wait(); err != nil {
			return fmt.Errorf("failed to complete: %s", err)
		}

		<-args.DataDone

		rr.ExitCode = int(op.Metadata["return"].(float64))

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			rr.Timeout = true
		}
	}

	outW.Close()
	errW.Close()
	<-outDoneCh
	<-errDoneCh

	rr.Stdout = strings.TrimSpace(outBuf.String())
	rr.Stderr = strings.TrimSpace(errBuf.String())
	rr.Applied = true

	return &rr, err
}

// FileUpload implements the FileUpload method of the Connection interface.
func (r LXD) FileUpload(cfo CopyFileOptions) (*FileResult, error) {
	var fr FileResult

	if cfo.Source == "" {
		return nil, fmt.Errorf("source is required for file upload")
	}

	if cfo.Destination == "" {
		return nil, fmt.Errorf("destination is required for file upload")
	}

	if cfo.Mode == 0 {
		cfo.Mode = os.FileMode(0640)
	}

	timeout := LXDCommandTimeout
	if cfo.Timeout > 0 {
		timeout = cfo.Timeout
	}

	local, err := os.Open(cfo.Source)
	if err != nil {
		return nil, err
	}
	defer local.Close()

	args := lxd.ContainerFileArgs{
		Type:    "file",
		Content: local,
		UID:     int64(cfo.UID),
		GID:     int64(cfo.GID),
		Mode:    int(cfo.Mode.Perm()),
	}

	err = timeoutFunc(timeout, func() error {
		err = r.client.CreateContainerFile(r.Host, cfo.Destination, args)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			fr.Timeout = true
		}
	}

	if err == nil {
		fr.Success = true
	}

	fr.Applied = true

	return &fr, err
}

// FileDownload implements the FileUpload method of the Connection interface.
func (r LXD) FileDownload(cfo CopyFileOptions) (*FileResult, error) {
	var fr FileResult

	if cfo.Source == "" {
		return nil, fmt.Errorf("source is required for file download")
	}

	if cfo.Destination == "" {
		return nil, fmt.Errorf("destination is required for file download")
	}

	if cfo.Mode == 0 {
		cfo.Mode = os.FileMode(0640)
	}

	timeout := LXDCommandTimeout
	if cfo.Timeout > 0 {
		timeout = cfo.Timeout
	}

	local, err := os.OpenFile(cfo.Destination, os.O_RDWR|os.O_CREATE, cfo.Mode)
	if err != nil {
		return nil, err
	}
	defer local.Close()

	err = timeoutFunc(timeout, func() error {
		buf, _, err := r.client.GetContainerFile(r.Host, cfo.Source)
		if err != nil {
			return err
		}

		_, err = io.Copy(local, buf)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			fr.Timeout = true
		}
	}

	if err == nil {
		fr.Success = true
	}

	fr.Applied = true

	return &fr, err
}

// FileInfo implements the FileInfo method of the Connection interface.
func (r LXD) FileInfo(fo FileOptions) (*FileResult, error) {
	var fr FileResult

	// validate options
	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file exists")
	}

	ro := RunOptions{
		Command: fmt.Sprintf(`stat -c"%u:%g:%n:%s:%a:%F" "%s"`, fo.Path),
		Timeout: fo.Timeout,
	}

	rr, err := r.RunCommand(ro)
	if err != nil {
		return &fr, err
	}

	if rr.ExitCode != 0 {
		fr.Exists = false
		fr.Success = true
		return &fr, nil
	}

	parts := strings.Split(rr.Stdout, ":")
	if len(parts) != 6 {
		return nil, fmt.Errorf("unable to get file information for %s", fo.Path)
	}

	uid, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("unable to get file information for %s", fo.Path)
	}

	gid, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to get file information for %s", fo.Path)
	}

	size, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to get file information for %s", fo.Path)
	}

	mode, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("unable to get file information for %s", fo.Path)
	}

	fi := FileInfo{
		UID:  uid,
		GID:  gid,
		Name: parts[2],
		Size: size,
		Mode: mode,
	}

	switch parts[5] {
	case "regular file":
		fi.Type = "file"
	case "directory":
		fi.Type = "directory"
	case "symbolic link":
		fi.Type = "symlink"
	case "socket":
		fi.Type = "socket"
	}

	fr.Exists = true
	fr.Success = true
	fr.Applied = true

	return &fr, nil
}

// FileDelete implements the FileDelete method of the Connection interface.
func (r LXD) FileDelete(fo FileOptions) (*FileResult, error) {
	var fr FileResult

	// validate options
	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file delete")
	}

	timeout := LXDCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	err := timeoutFunc(timeout, func() error {
		if err := r.client.DeleteContainerFile(r.Host, fo.Path); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			fr.Timeout = true
		}
	}

	if err == nil {
		fr.Success = true
	}

	fr.Applied = true

	return &fr, err
}

// Close implements the Close method of the Connection interface.
// It doesn't do anything since communication with LXD is not persistent.
func (r LXD) Close() {
	return
}

// copyFile is an internal function to manage both Upload and Download.
func (r LXD) copyFile(cfo CopyFileOptions, action string) (*FileResult, error) {
	var fr FileResult

	return &fr, nil
}
