package connections

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/mitchellh/mapstructure"
)

const (
	LocalCommandShell   = "/bin/bash"
	LocalCommandTimeout = 60

	LocalFileCopyMaxBytes = 4096
)

// Local represents a local connection.
type Local struct {
	c     *exec.Cmd
	Shell string `mapstructure:"shell"`
}

// NewLocal will return a Local connection.
func NewLocal(options map[string]interface{}) (*Local, error) {
	var local Local

	err := mapstructure.Decode(options, &local)
	if err != nil {
		return nil, err
	}

	if local.Shell == "" {
		local.Shell = LocalCommandShell
	}

	return &local, nil
}

// Connect implements the Connect method of the Connection interface.
// It's only here to satisfy the interface.
func (r Local) Connect() error {
	return nil
}

// RunCommand implements the RunCommand method of the Connection interface.
func (r Local) RunCommand(ro RunOptions) (*RunResult, error) {
	var err error
	var rr RunResult
	var outBuf, errBuf bytes.Buffer

	// Validate options
	if ro.Command == "" {
		return nil, fmt.Errorf("a command is required")
	}

	// Build the command
	cmdArgs := []string{r.Shell, "-c", ro.Command}
	r.c = exec.Command(cmdArgs[0], cmdArgs[1:]...)

	// Set up the output
	log := ioutil.Discard
	if ro.Log != nil {
		log = *ro.Log
	}

	outR, outW := io.Pipe()
	if err != nil {
		return nil, err
	}

	errR, errW := io.Pipe()
	if err != nil {
		return nil, err
	}

	r.c.Stdout = outW
	r.c.Stderr = errW

	outTee := io.TeeReader(outR, &outBuf)
	errTee := io.TeeReader(errR, &errBuf)
	outDoneCh := make(chan struct{})
	errDoneCh := make(chan struct{})
	go printOutput(log, outTee, outDoneCh)
	go printOutput(log, errTee, errDoneCh)

	timeout := LocalCommandTimeout
	if ro.Timeout > 0 {
		timeout = ro.Timeout
	}

	err = timeoutFunc(timeout, func() error {
		if err := r.c.Start(); err != nil {
			return err
		}

		if err := r.c.Wait(); err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				rr.ExitCode = int(exit.ProcessState.Sys().(syscall.WaitStatus) / 256)
				return nil
			}

			return err
		}

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
// It peforms a local file copy.
func (r Local) FileUpload(fo CopyFileOptions) (*FileResult, error) {
	return r.copyFile(fo)
}

// FileDownload implements the FileDownload method of the Connection interface.
// It peforms a local file copy.
func (r Local) FileDownload(fo CopyFileOptions) (*FileResult, error) {
	return r.copyFile(fo)
}

// FileInfo implements the FileInfo method of the Connection interface.
func (r Local) FileInfo(fo FileOptions) (*FileResult, error) {
	var fr FileResult
	var fi FileInfo
	var err error

	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file exists")
	}

	timeout := LocalCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	err = timeoutFunc(timeout, func() error {
		stat, err := os.Stat(fo.Path)
		if err == nil {
			fi.Name = stat.Name()
			fi.Size = stat.Size()
			fi.UID = int(stat.Sys().(*syscall.Stat_t).Uid)
			fi.GID = int(stat.Sys().(*syscall.Stat_t).Gid)

			mode := fmt.Sprintf("%o", int(stat.Mode().Perm()))
			fi.Mode, _ = strconv.Atoi(mode)

			if stat.IsDir() {
				fi.Type = "directory"
			}

			if stat.Mode() == os.ModeSymlink {
				fi.Type = "symlink"
			}

			if stat.Mode() == os.ModeSocket {
				fi.Type = "socket"
			}

			if fi.Type == "" {
				fi.Type = "file"
			}
		}

		return err
	})

	if err != nil {
		if err.Error() == "timeout" {
			fr.Timeout = true
		}

		if os.IsNotExist(err) {
			fr.Exists = false
			fr.Success = true
		}

		return &fr, nil
	}

	fr.FileInfo = fi
	fr.Exists = true
	fr.Success = true
	fr.Applied = true

	return &fr, nil
}

// FileDelete implements the FileDelete method of the Connection interface.
func (r Local) FileDelete(fo FileOptions) (*FileResult, error) {
	var fr FileResult
	var err error

	// validate options
	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file delete")
	}

	timeout := LocalCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	err = timeoutFunc(timeout, func() error {
		if err := os.Remove(fo.Path); err != nil {
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
// It peforms no action.
func (r Local) Close() {
	return
}

func (r Local) copyFile(fo CopyFileOptions) (*FileResult, error) {
	var fr FileResult

	// validate options
	if fo.Source == "" {
		return nil, fmt.Errorf("source is required for file copy")
	}

	if fo.Destination == "" {
		return nil, fmt.Errorf("destination is required for file copy")
	}

	if fo.Mode == 0 {
		fo.Mode = os.FileMode(0640)
	}

	timeout := LocalCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	destination, err := os.OpenFile(fo.Destination, os.O_RDWR|os.O_CREATE, fo.Mode)
	if err != nil {
		return nil, err
	}
	defer destination.Close()

	source, err := os.Open(fo.Source)
	if err != nil {
		return nil, err
	}
	defer source.Close()

	err = timeoutFunc(timeout, func() error {
		buf := make([]byte, LocalFileCopyMaxBytes)
		if _, err := io.CopyBuffer(destination, source, buf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			fr.Timeout = true
		}

		return nil, err
	}

	destination.Close()
	source.Close()

	if err := os.Chown(fo.Destination, fo.UID, fo.GID); err != nil {
		return nil, err
	}

	fr.Success = true
	fr.Applied = true

	return &fr, err
}
