package testing

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

func TestLXD_Basic(t *testing.T) {
	os.Setenv("YAK_CONFIG_FILE", "fixtures/yak.cfg")

	config := &yakfile.Connection{
		Type: "lxd",
		Options: map[string]interface{}{
			"yak_auth": "lxd",
			"host":     "c1",
		},
	}

	lxd, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := lxd.Connect(); err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "echo hi",
	}

	rr, err := lxd.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hi", rr.Stdout)

	ro.Command = "asdf"
	rr, err = lxd.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "/bin/bash: asdf: command not found", rr.Stderr)

	ro.Command = `foo=bar; sleep 1; echo foobar >&2; echo $foo ; echo 123 >&2`
	rr, err = lxd.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "bar", rr.Stdout)
	assert.Equal(t, "foobar\n123", rr.Stderr)
}

func TestLXD_CommandTimeout(t *testing.T) {
	os.Setenv("YAK_CONFIG_FILE", "fixtures/yak.cfg")

	config := &yakfile.Connection{
		Type: "lxd",
		Options: map[string]interface{}{
			"yak_auth": "lxd",
			"host":     "c1",
		},
	}

	lxd, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := lxd.Connect(); err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "sleep 6; echo timeout",
		Timeout: 5,
	}

	rr, err := lxd.RunCommand(ro)
	assert.Equal(t, true, rr.Timeout)
}

func TestLXD_ConnectTimeout(t *testing.T) {
	os.Setenv("YAK_CONFIG_FILE", "fixtures/yak.cfg")

	config := &yakfile.Connection{
		Type: "lxd",
		Options: map[string]interface{}{
			"yak_auth": "badlxd",
			"host":     "c1",
			"timeout":  5,
		},
	}

	lxd, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	err = lxd.Connect()
	assert.Equal(t, "timed out connecting to foobar", err.Error())
}

func TestLXD_CopyFileDelete(t *testing.T) {
	os.Setenv("YAK_CONFIG_FILE", "fixtures/yak.cfg")

	config := &yakfile.Connection{
		Type: "lxd",
		Options: map[string]interface{}{
			"yak_auth": "lxd",
			"host":     "c1",
		},
	}

	lxd, err := connections.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	if err := lxd.Connect(); err != nil {
		t.Fatal(err)
	}

	cfo := connections.CopyFileOptions{
		Source:      "fixtures/hello.txt",
		Destination: "/tmp/yakfoo.txt",
	}

	fr, err := lxd.FileUpload(cfo)
	if err != nil {
		t.Fatal(err)
	}

	ro := connections.RunOptions{
		Command: "cat /tmp/yakfoo.txt",
	}

	rr, err := lxd.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, fr.Success)
	assert.Equal(t, "Hello, World!", rr.Stdout)

	tmpfile, err := ioutil.TempFile("/tmp", "yak")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	cfo = connections.CopyFileOptions{
		Source:      "/tmp/yakfoo.txt",
		Destination: tmpfile.Name(),
	}

	fr, err = lxd.FileDownload(cfo)
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
		Path: "/tmp/yakfoo.txt",
	}

	fr, err = lxd.FileDelete(fo)
	if err != nil {
		t.Fatal(err)
	}

	ro.Command = "stat /tmp/yakfoo.txt"
	rr, err = lxd.RunCommand(ro)
	if err != nil {
		t.Fatal(err)
	}

	if rr.ExitCode != 1 {
		t.Fatalf("file still exist")
	}

	assert.Equal(t, true, fr.Success)
}
