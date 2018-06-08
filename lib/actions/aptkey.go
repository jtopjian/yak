package actions

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/utils"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"
)

// AptKey represents options for an apt.key action.
type AptKey struct {
	BaseFields `mapstructure:",squash"`

	// KeyServer is an optional remote server to obtain the key from.
	// If KeyServer is not used, RemoteKeyFile must be used.
	KeyServer string `mapstructure:"key_server"`

	// RemoteKeyFile is the URL to a public key.
	// If RemoteKeyFile is not used, KeyServer must be used.
	RemoteKeyFile string `mapstructure:"remote_key_file"`
}

// AptKeyAction will perform a full state cycle for an apt.key.
func AptKeyAction(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (change bool, err error) {

	var key AptKey

	err = mapstructure.Decode(step.Input, &key)
	if err != nil {
		return
	}

	err = utils.ValidateTags(&key)
	if err != nil {
		return
	}

	if key.KeyServer == "" && key.RemoteKeyFile == "" {
		err = fmt.Errorf(
			"unable to add apt.key %s: one of key_server or remote_key_file must be specified", key.Name)
		return
	}

	key.conn = conn
	key.setLogger(ctx, "apt.key", key.Name, key.State)

	exists, err := key.Exists()
	if err != nil {
		return
	}

	if key.State == "absent" {
		if exists {
			err = key.Delete()
			change = true
			return
		}

		return
	}

	if !exists {
		err = key.Create()
		change = true
		return
	}

	return
}

// Exists will determine if an apt.key exists.
func (r AptKey) Exists() (bool, error) {
	eo := ExecOptions{
		Command: fmt.Sprintf("apt-key export %s", r.Name),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("checking if installed")
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return false, fmt.Errorf("unable to check status of apt.key %s: %s", r.Name, err)
	}

	if rr.Stdout == "" {
		r.logInfo("not installed")
		return false, nil
	}

	/*
		_, err = aptKeyGetName(rr.Stdout)
		if err != nil {
			return false, err
		}
	*/

	r.logInfo("installed")
	return true, nil
}

// Create will create a key via apt-key.
func (r AptKey) Create() error {
	var cfo CopyFileOptions

	eo := ExecOptions{
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logInfo("adding")

	if r.RemoteKeyFile != "" {
		k, err := aptKeyGetRemoteKeyFile(r.RemoteKeyFile)
		if err != nil {
			return err
		}

		tmpfile, err := ioutil.TempFile("/tmp", "apt.key")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())

		if _, err = tmpfile.Write([]byte(k)); err != nil {
			return err
		}

		if err = tmpfile.Close(); err != nil {
			return err
		}

		cfo.Source = tmpfile.Name()
		cfo.Destination = tmpfile.Name()
		if _, err := fileUpload(r.ctx, r.conn, cfo); err != nil {
			return err
		}

		eo.Command = fmt.Sprintf("apt-key add %s", tmpfile.Name())
		r.logDebug("running command: %s", eo.Command)
		rr, err := exec(r.ctx, r.conn, eo)
		if err != nil {
			return err
		}

		if rr.ExitCode != 0 {
			r.logDebug(rr.Stderr)
			return fmt.Errorf("unable to add key: %s", err)
		}

		fo := FileOptions{
			Path: tmpfile.Name(),
		}
		fr, err := fileDelete(r.ctx, r.conn, fo)
		if err != nil {
			return err
		}

		if !fr.Success {
			return fmt.Errorf("unable to delete temporary key from remote host: %s", tmpfile.Name())
		}
	}

	if r.KeyServer != "" {
		eo.Command = fmt.Sprintf("apt-key adv --keyserver %s --recv-keys %s",
			r.KeyServer, r.Name)

		r.logDebug("running command: %s", eo.Command)
		rr, err := exec(r.ctx, r.conn, eo)
		if err != nil {
			return err
		}

		if rr.Stderr != "" {
			return fmt.Errorf("unable to add key: %s", err)
		}
	}

	r.logInfo("installed")
	return nil
}

// Delete deletes a key managed by apt.key.
func (r AptKey) Delete() error {
	eo := ExecOptions{
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logInfo("deleting")
	eo.Command = fmt.Sprintf("apt-key del %s", r.Name)
	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		return err
	}

	if rr.ExitCode != 0 {
		r.logDebug(rr.Stderr)
		return fmt.Errorf("unable to delete key: %s", err)
	}

	r.logInfo("deleted")
	return nil
}

// aptKeyGetRemoteKeyFile is an internal function that will
// download a key located at a remote URL.
func aptKeyGetRemoteKeyFile(v string) (key string, err error) {
	res, err := http.Get(v)
	if err != nil {
		return
	}

	k, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	key = string(k)

	return
}

// aptKeyGetShortID is an internal function that will print the
// short key ID of a public key.
func aptKeyGetShortID(key string) (fingerprint string, err error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(key))
	if err != nil {
		return
	}

	if len(el) == 0 {
		err = fmt.Errorf("Error determining fingerprint of key")
		return
	}

	fingerprint = el[0].PrimaryKey.KeyIdShortString()

	return
}

// aptKeyGetName is an internal function that will get the
// maintainer name of a public key.
func aptKeyGetName(key string) (name string, err error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(key))
	if err != nil {
		return
	}

	if len(el) == 0 {
		err = fmt.Errorf("Error determining userid of key")
		return
	}

	identities := el[0].Identities
	for k, _ := range identities {
		if name == "" {
			name = k
		}
	}

	return
}

func aptKeyParseList(list string) (keys []string) {
	keyRe := regexp.MustCompile("^pub.+/(.+) [0-9-]+$")
	for _, line := range strings.Split(list, "\n") {
		v := keyRe.FindStringSubmatch(line)
		if v != nil {
			keys = append(keys, v[1])
		}
	}

	return
}
