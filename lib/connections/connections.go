package connections

import (
	"fmt"
	"io"
	"os"
)

// Connection is an interface which specifies what drivers
// must implement.
type Connection interface {
	Connect() error
	Close()

	RunCommand(RunOptions) (*RunResult, error)

	FileInfo(FileOptions) (*FileResult, error)
	FileDelete(FileOptions) (*FileResult, error)
	FileUpload(CopyFileOptions) (*FileResult, error)
	FileDownload(CopyFileOptions) (*FileResult, error)
}

// RunOptions represents options for running commands.
type RunOptions struct {
	Command string
	Timeout int
	Log     *io.Writer
}

// RunResult respresents the result of an command execution.
type RunResult struct {
	ExitCode int
	Stderr   string
	Stdout   string
	Timeout  bool
	Applied  bool
}

// CopyFileOptions represents options for copying files.
type CopyFileOptions struct {
	Source      string
	Destination string
	UID         int
	GID         int
	Mode        os.FileMode
	Timeout     int
}

// FileOptions represents options for managing a generic file.
type FileOptions struct {
	Path    string
	UID     int
	GID     int
	Mode    os.FileMode
	Timeout int
}

// FileResult represents the result of an file action.
type FileResult struct {
	Exists   bool
	Success  bool
	Timeout  bool
	Applied  bool
	FileInfo FileInfo
}

// FileInfo represents information about a file.
type FileInfo struct {
	Name string
	UID  int
	GID  int
	Type string
	Size int64
	Mode int
}

// New will return a connection based on a given connection type.
func New(connType string, options map[string]interface{}) (Connection, error) {
	if connType == "" {
		return nil, fmt.Errorf("a connection type was not specified")
	}

	switch connType {
	case "local":
		return NewLocal(options)
	case "lxd":
		return NewLXD(options)
	case "ssh":
		return NewSSH(options)
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connType)
	}

	return nil, nil
}
