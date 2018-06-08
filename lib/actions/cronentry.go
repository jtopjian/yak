package actions

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jtopjian/yak/lib/connections"
	"github.com/jtopjian/yak/lib/utils"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/mitchellh/mapstructure"
)

// CronEntry represents options for a cron.entry action.
type CronEntry struct {
	BaseFields `mapstructure:",squash"`

	// User is the user who owns the cron entry.
	User string `mapstructure:"user" default:"root"`

	// Command is the command which cron will run.
	Command string `mapstructure:"command" required:"true"`

	// Minute is the minute field of the cron entry.
	Minute string `mapstructure:"minute" default:"*"`

	// Hour is the hour field of the cron entry.
	Hour string `mapstructure:"hour" default:"*"`

	// DayOfMonth is the day of the month field of the cron entry.
	DayOfMonth string `mapstructure:"day_of_month" default:"*"`

	// Month is the month field of the cron entry.
	Month string `mapstructure:"month" default:"*"`

	// DayOfWeek is the day of the week field of the cron entry.
	DayOfWeek string `mapstructure:"day_of_week" default:"*"`
}

// CronEntryAction will perform a full state cycle for a cron.entry.
func CronEntryAction(
	ctx context.Context,
	conn connections.Connection,
	step yakfile.Step,
) (change bool, err error) {
	var ce CronEntry

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &ce,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return
	}

	err = decoder.Decode(step.Input)
	if err != nil {
		return
	}

	err = utils.ValidateTags(&ce)
	if err != nil {
		return
	}

	ce.conn = conn
	ce.setLogger(ctx, "cron.entry", ce.Name, ce.State)

	exists, err := ce.Exists()
	if err != nil {
		return
	}

	if ce.State == "absent" {
		if exists {
			err = ce.Delete()
			change = true
			return
		}

		return
	}

	if !exists {
		err = ce.Create()
		change = true
		return
	}

	return
}

// Exists will determine if a cron.entry exists.
func (r CronEntry) Exists() (bool, error) {
	r.logDebug("checking if installed")

	entries, err, stderr := r.getEntries()
	if err != nil {
		return false, fmt.Errorf("unable to check status of cron.entry %s: %s", r.Name, err)
	}

	if stderr != nil {
		r.logInfo("not installed")
		return false, nil
	}

	var exists bool
	for _, line := range entries {
		if line == r.entry() {
			exists = true
		}
	}

	if exists {
		r.logInfo("exists")
		return true, nil
	}

	r.logInfo("not installed")
	return false, nil
}

// Create will create a cron.entry.
func (r CronEntry) Create() error {
	r.logInfo("adding")

	entries, err, _ := r.getEntries()
	if err != nil {
		return fmt.Errorf("unable to add cron.entry %s: %s", r.Name, err)
	}

	var newEntries []string
	var added bool
	for _, line := range entries {
		if strings.Contains(line, fmt.Sprintf(`# %s`, r.Name)) {
			line = r.entry()
			added = true
		}
		newEntries = append(newEntries, line)
	}

	if !added {
		newEntries = append(newEntries, r.entry())
	}

	newEntries = append(newEntries, "\n")

	if err := r.pushEntries(newEntries); err != nil {
		return fmt.Errorf("unable to add cron.entry %s: %s", r.Name, err)
	}

	return nil
}

// Delete will delete a cron.entry.
func (r CronEntry) Delete() error {
	r.logInfo("deleting")

	entries, err, stderr := r.getEntries()
	if err != nil {
		return fmt.Errorf("unable to delete cron.entry %s: %s", r.Name, err)
	}

	if stderr != nil {
		return fmt.Errorf("unable to add cron.entry %s: %s", r.Name, stderr)
	}

	var newEntries []string
	for _, line := range entries {
		if line != r.entry() {
			newEntries = append(newEntries, line)
		}
	}

	newEntries = append(newEntries, "\n")

	if err := r.pushEntries(newEntries); err != nil {
		return fmt.Errorf("unable to delete cron.entry %s: %s", r.Name, err)
	}

	return nil
}

// entry returns the formatted cron entry.
func (r CronEntry) entry() string {
	entry := fmt.Sprintf(`%s %s %s %s %s %s # %s`,
		r.Minute, r.Hour, r.DayOfMonth, r.Month,
		r.DayOfWeek, r.Command, r.Name)

	return entry
}

// getEntries returns the cron entries from a remote host.
func (r CronEntry) getEntries() ([]string, error, error) {
	eo := ExecOptions{
		Command: fmt.Sprintf("crontab -u %s -l", r.User),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return nil, nil, err
	}

	if rr.ExitCode != 0 {
		return nil, nil, fmt.Errorf("%s", rr.Stderr)
	}

	return strings.Split(rr.Stdout, "\n"), nil, nil
}

// pushEntries pushes new entries to a remote host.
func (r CronEntry) pushEntries(entries []string) error {
	tmpfile, err := ioutil.TempFile("/tmp", "cron.entry")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(strings.Join(entries, "\n"))); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	cfo := CopyFileOptions{
		Source:      tmpfile.Name(),
		Destination: tmpfile.Name(),
	}

	if _, err := fileUpload(r.ctx, r.conn, cfo); err != nil {
		return err
	}

	eo := ExecOptions{
		Command: fmt.Sprintf(`crontab -u %s %s`, r.User, tmpfile.Name()),
		Sudo:    r.Sudo,
		Timeout: r.Timeout,
	}

	r.logDebug("running command: %s", eo.Command)
	rr, err := exec(r.ctx, r.conn, eo)
	if err != nil {
		r.logDebug(rr.Stderr)
		return err
	}

	if rr.ExitCode != 0 {
		return fmt.Errorf("%s", rr.Stderr)
	}

	fo := FileOptions{
		Path: tmpfile.Name(),
	}

	_, err = fileDelete(r.ctx, r.conn, fo)
	if err != nil {
		return err
	}

	return nil
}
