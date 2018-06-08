package connections

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"

	"github.com/pkg/sftp"
)

const (
	SSHDefaultPort  = 22
	SSHDefaultShell = "/bin/bash"
	SSHDefaultUser  = "root"

	SSHCommandTimeout    = 60
	SSHConnectionTimeout = 300

	SCPMaxPacketSize = 32768
	SCPMaxBytes      = 4096
)

// SSH represents an SSH connection.
type SSH struct {
	Agent      bool   `mapstructure:"agent"`
	Host       string `mapstructure:"host"`
	PrivateKey string `mapstructure:"private_key"`
	Port       int    `mapstructure:"port"`
	Shell      string `mapstructure:"shell"`
	Timeout    int    `mapstructure:"timeout"`
	User       string `mapstructure:"user"`

	BastionUser       string `mapstructure:"bastion_user"`
	BastionPrivateKey string `mapstructure:"bastion_private_key"`
	BastionHost       string `mapstructure:"bastion_host"`
	BastionPort       int    `mapstructure:"bastion_port"`

	bastionConfig *ssh.ClientConfig
	bastionConn   *net.Conn
	client        *ssh.Client
	config        *ssh.ClientConfig
	sftp          *sftp.Client
}

// NewSSH will return an SSH client.
func NewSSH(options map[string]interface{}) (*SSH, error) {
	var sshConfig SSH

	err := mapstructure.Decode(options, &sshConfig)
	if err != nil {
		return nil, err
	}

	if sshConfig.Host == "" {
		return nil, fmt.Errorf("a host is requried for ssh")
	}

	var signer ssh.Signer
	if sshConfig.PrivateKey == "" {
		// If no private_key was specified, try using $user/.ssh/id_rsa.
		if homeDir, err := homedir.Dir(); err == nil {
			privateKey := filepath.Join(homeDir, ".ssh", "id_rsa")
			if _, err := os.Stat(privateKey); err == nil {
				sshConfig.PrivateKey = privateKey
			}
		}
	}

	if sshConfig.PrivateKey != "" {
		privateKey, err := homedir.Expand(sshConfig.PrivateKey)
		if err != nil {
			return nil, err
		}

		if _, err := os.Stat(privateKey); os.IsNotExist(err) {
			return nil, fmt.Errorf("private_key %s does not exist", privateKey)
		}

		key, err := ioutil.ReadFile(privateKey)
		if err != nil {
			return nil, err
		}

		signer, err = ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
	}

	var bastionSigner ssh.Signer
	if sshConfig.BastionPrivateKey != "" {
		key, err := ioutil.ReadFile(sshConfig.BastionPrivateKey)
		if err != nil {
			return nil, err
		}

		bastionSigner, err = ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
	}

	if sshConfig.Port == 0 {
		sshConfig.Port = SSHDefaultPort
	}

	if sshConfig.BastionPort == 0 {
		sshConfig.BastionPort = SSHDefaultPort
	}

	if sshConfig.User == "" {
		sshConfig.User = SSHDefaultUser
	}

	if sshConfig.BastionUser == "" {
		sshConfig.BastionUser = SSHDefaultUser
	}

	if sshConfig.Shell == "" {
		sshConfig.Shell = SSHDefaultShell
	}

	authMethod := ssh.PublicKeys(signer)
	if sshConfig.Agent {
		if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			authMethod = ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
		}
	}

	sshConfig.config = &ssh.ClientConfig{
		User: sshConfig.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if sshConfig.BastionHost != "" {
		bastionAuthMethod := ssh.PublicKeys(bastionSigner)
		if sshConfig.Agent {
			if bastionAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
				bastionAuthMethod = ssh.PublicKeysCallback(agent.NewClient(bastionAgent).Signers)
			}
		}

		sshConfig.bastionConfig = &ssh.ClientConfig{
			User: sshConfig.BastionUser,
			Auth: []ssh.AuthMethod{
				bastionAuthMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	return &sshConfig, nil
}

// Connect implements the Connect method of the Connection interface.
// It will connect to a host via SSH.
func (r *SSH) Connect() error {
	var err error

	// If a connection has already been made, don't do anything.
	if r.client != nil {
		return nil
	}

	connectTimeout := SSHConnectionTimeout
	if r.Timeout > 0 {
		connectTimeout = r.Timeout
	}

	host := fmt.Sprintf("%s:%d", r.Host, r.Port)
	bastionHost := fmt.Sprintf("%s:%d", r.BastionHost, r.BastionPort)

	err = retryFunc(connectTimeout, func() error {
		if r.BastionHost != "" {
			r.client, err = ssh.Dial("tcp", bastionHost, r.bastionConfig)
			if err != nil {
				return err
			}

			conn, err := r.client.Dial("tcp", host)
			if err != nil {
				return err
			}
			r.bastionConn = &conn

			return nil
		}

		r.client, err = ssh.Dial("tcp", host, r.config)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "timeout" {
			return fmt.Errorf("timed out connecting to %s", host)
		}
	}

	return nil
}

// RunCommand implements the Run method of the Connection interface.
func (r SSH) RunCommand(ro RunOptions) (*RunResult, error) {
	var rr RunResult
	var outBuf, errBuf bytes.Buffer

	// validate options
	if ro.Command == "" {
		return nil, fmt.Errorf("a command is required")
	}

	timeout := SSHCommandTimeout
	if ro.Timeout > 0 {
		timeout = ro.Timeout
	}

	// Set up a session
	session, err := r.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// Set up the output
	log := ioutil.Discard
	if ro.Log != nil {
		log = *ro.Log
	}

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	session.Stdout = outW
	session.Stderr = errW

	outTee := io.TeeReader(outR, &outBuf)
	errTee := io.TeeReader(errR, &errBuf)
	outDoneCh := make(chan struct{})
	errDoneCh := make(chan struct{})
	go printOutput(log, outTee, outDoneCh)
	go printOutput(log, errTee, errDoneCh)

	//cmd := strings.Replace(ro.Command, `"`, `\"`, -1)
	cmd := fmt.Sprintf(`%s -c "%s"`, r.Shell, ro.Command)

	err = timeoutFunc(timeout, func() error {
		if err := session.Start(cmd); err != nil {
			return err
		}

		if err := session.Wait(); err != nil {
			if exit, ok := err.(*ssh.ExitError); ok {
				rr.ExitCode = exit.Waitmsg.ExitStatus()
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
func (r SSH) FileUpload(cfo CopyFileOptions) (*FileResult, error) {
	return r.copyFile(cfo, "upload")
}

// FileDownload implements the FileUpload method of the Connection interface.
func (r SSH) FileDownload(cfo CopyFileOptions) (*FileResult, error) {
	return r.copyFile(cfo, "download")
}

// FileInfo implements the FileInfo method of the Connection interface.
func (r SSH) FileInfo(fo FileOptions) (*FileResult, error) {
	var fr FileResult
	var fi FileInfo
	var err error

	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file exists")
	}

	timeout := SSHCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	client, err := sftp.NewClient(r.client, sftp.MaxPacket(SCPMaxPacketSize))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	err = timeoutFunc(timeout, func() error {
		stat, err := client.Stat(fo.Path)
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
func (r SSH) FileDelete(fo FileOptions) (*FileResult, error) {
	var fr FileResult

	// validate options
	if fo.Path == "" {
		return nil, fmt.Errorf("path is required for file delete")
	}

	timeout := SSHCommandTimeout
	if fo.Timeout > 0 {
		timeout = fo.Timeout
	}

	client, err := sftp.NewClient(r.client, sftp.MaxPacket(SCPMaxPacketSize))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	err = timeoutFunc(timeout, func() error {
		if err := client.Remove(fo.Path); err != nil {
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
// It will close an SSH and bastion connection if they are opened.
func (r SSH) Close() {
	if r.bastionConn != nil {
		(*r.bastionConn).Close()
		r.bastionConn = nil
	}

	if r.client != nil {
		r.client.Close()
		r.client = nil
	}

}

// copyFile is an internal function to manage both Upload and Download.
func (r SSH) copyFile(cfo CopyFileOptions, action string) (*FileResult, error) {
	var fr FileResult

	// validate options
	if cfo.Source == "" {
		return nil, fmt.Errorf("source is required for file %s", action)
	}

	if cfo.Destination == "" {
		return nil, fmt.Errorf("destination is required for file %s", action)
	}

	if cfo.Mode == 0 {
		cfo.Mode = os.FileMode(0640)
	}

	timeout := SSHCommandTimeout
	if cfo.Timeout > 0 {
		timeout = cfo.Timeout
	}

	client, err := sftp.NewClient(r.client, sftp.MaxPacket(SCPMaxPacketSize))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var remote *sftp.File
	var local *os.File
	switch action {
	case "upload":
		remote, err = client.OpenFile(cfo.Destination, os.O_RDWR|os.O_CREATE)
		if err != nil {
			return nil, err
		}
		defer remote.Close()

		local, err = os.Open(cfo.Source)
		if err != nil {
			return nil, err
		}
		defer local.Close()
	case "download":
		local, err = os.OpenFile(cfo.Destination, os.O_RDWR|os.O_CREATE, cfo.Mode)
		if err != nil {
			return nil, err
		}
		defer local.Close()

		remote, err = client.Open(cfo.Source)
		if err != nil {
			return nil, err
		}
		defer remote.Close()
	}

	err = timeoutFunc(timeout, func() error {
		switch action {
		case "upload":
			if _, err := io.Copy(remote, io.LimitReader(local, SCPMaxBytes)); err != nil {
				return err
			}
		case "download":
			if _, err := io.Copy(local, io.LimitReader(remote, SCPMaxBytes)); err != nil {
				return err
			}
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
