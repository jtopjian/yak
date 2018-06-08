package testing

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

func TestLocal_Basic(t *testing.T) {
	config := &yakfile.Connection{
		Type: "local",
		Options: map[string]interface{}{
			"shell": "/bin/bash",
		},
	}

	local, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "echo hi",
	}

	rr, err := local.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hi", rr.Stdout)

	ro.Command = "asdf"
	rr, err = local.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "/bin/bash: asdf: command not found", rr.Stderr)

	ro.Command = "foo=bar; sleep 1; echo foobar >&2; echo $foo ; echo 123 >&2"
	rr, err = local.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "bar", rr.Stdout)
	assert.Equal(t, "foobar\n123", rr.Stderr)
}

func TestLocal_CommandTimeout(t *testing.T) {
	config := &yakfile.Connection{
		Type: "local",
		Options: map[string]interface{}{
			"shell": "/bin/bash",
		},
	}

	local, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "sleep 6",
		Timeout: 5,
	}

	_, err = local.RunCommand(ro)
	assert.Equal(t, err.Error(), "timeout")
}

func TestLocal_CopyFile(t *testing.T) {
	config := &yakfile.Connection{
		Type: "local",
		Options: map[string]interface{}{
			"shell": "/bin/bash",
		},
	}

	local, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("/tmp", "yak")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	cfo := connections.CopyFileOptions{
		Source:      "fixtures/hello.txt",
		Destination: tmpfile.Name(),
	}

	fr, err := local.FileUpload(cfo)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, fr.Success)
	assert.Equal(t, "Hello, World!\n", string(actual))
}

func TestLocal_FileDelete(t *testing.T) {
	config := &yakfile.Connection{
		Type: "local",
		Options: map[string]interface{}{
			"shell": "/bin/bash",
		},
	}

	local, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("/tmp", "yak")
	if err != nil {
		t.Fatal(err)
	}

	cfo := connections.CopyFileOptions{
		Source:      "fixtures/hello.txt",
		Destination: tmpfile.Name(),
	}

	fr, err := local.FileUpload(cfo)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, fr.Success)
	assert.Equal(t, "Hello, World!\n", string(actual))

	fo := connections.FileOptions{
		Path: tmpfile.Name(),
	}

	fr, err = local.FileDelete(fo)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(tmpfile.Name()); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestLocal_FileInfo(t *testing.T) {
	config := &yakfile.Connection{
		Type: "local",
		Options: map[string]interface{}{
			"shell": "/bin/bash",
		},
	}

	local, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	fo := connections.FileOptions{
		Path: "fixtures/hello.txt",
	}

	fr, err := local.FileInfo(fo)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "file", fr.FileInfo.Type)
	assert.Equal(t, 644, fr.FileInfo.Mode)
	assert.Equal(t, int64(14), fr.FileInfo.Size)
}
