package testing

import (
	"testing"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

func TestSSH_Basic(t *testing.T) {
	config := &yakfile.Connection{
		Type: "ssh",
		Options: map[string]interface{}{
			"host":        "localhost",
			"user":        "ubuntu",
			"private_key": "/root/.ssh/id_rsa",
			"shell":       "/bin/bash",
			"timeout":     5,
		},
	}

	ssh, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := ssh.Connect(); err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "echo hi",
	}

	rr, err := ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hi", rr.Stdout)

	ro.Command = "asdf"
	rr, err = ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "/bin/bash: asdf: command not found", rr.Stderr)

	ro.Command = `foo=bar; sleep 1; echo foobar >&2; echo \$foo ; echo 123 >&2`
	rr, err = ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "bar", rr.Stdout)
	assert.Equal(t, "foobar\n123", rr.Stderr)
}

func TestSSH_CommandTimeout(t *testing.T) {
	config := &yakfile.Connection{
		Type: "ssh",
		Options: map[string]interface{}{
			"host":        "localhost",
			"user":        "ubuntu",
			"private_key": "/root/.ssh/id_rsa",
			"shell":       "/bin/bash",
		},
	}

	ssh, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := ssh.Connect(); err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "sleep 6; echo timeout",
		Timeout: 5,
	}

	rr, err := ssh.RunCommand(ro)
	assert.Equal(t, true, rr.Timeout)
}

func TestSSH_ConnectTimeout(t *testing.T) {
	config := &yakfile.Connection{
		Type: "ssh",
		Options: map[string]interface{}{
			"host":        "localhost2",
			"user":        "ubuntu",
			"private_key": "/root/.ssh/id_rsa",
			"shell":       "/bin/bash",
			"timeout":     5,
		},
	}

	ssh, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	err = ssh.Connect()
	assert.Equal(t, "timed out connecting to localhost2:22", err.Error())
}

func TestSSH_Bastion(t *testing.T) {
	config := &yakfile.Connection{
		Type: "ssh",
		Options: map[string]interface{}{
			"host":                "localhost",
			"user":                "ubuntu",
			"private_key":         "/root/.ssh/id_rsa",
			"shell":               "/bin/bash",
			"timeout":             5,
			"bastion_host":        "localhost",
			"bastion_user":        "ubuntu",
			"bastion_private_key": "/root/.ssh/id_rsa",
		},
	}

	ssh, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := ssh.Connect(); err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "echo hi",
	}

	rr, err := ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hi", rr.Stdout)
}

func TestSSH_CopyFileDelete(t *testing.T) {
	config := &yakfile.Connection{
		Type: "ssh",
		Options: map[string]interface{}{
			"host":        "localhost",
			"user":        "ubuntu",
			"private_key": "/root/.ssh/id_rsa",
			"shell":       "/bin/bash",
		},
	}

	ssh, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := ssh.Connect(); err != nil {
		t.Fatal(err)
	}

	cfo := connections.CopyFileOptions{
		Source:      "fixtures/hello.txt",
		Destination: "/tmp/yakfoo.txt",
	}

	fr, err := ssh.FileUpload(cfo)
	if err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "cat /tmp/yakfoo.txt",
	}

	rr, err := ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, fr.Success)
	assert.Equal(t, "Hello, World!", rr.Stdout)

	fo := connections.FileOptions{
		Path: "/tmp/yakfoo.txt",
	}

	fr, err = ssh.FileDelete(fo)
	if err != nil {
		t.Fatal(err)
	}

	ro.Command = "stat /tmp/yakfoo.txt"
	rr, err = ssh.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	if rr.ExitCode != 1 {
		t.Fatalf("file still exists")
	}

}
