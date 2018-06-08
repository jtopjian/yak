package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/jtopjian/yak/lib/utils"
)

// Default is a default, hard-coded config file.
// This is for when a config file cannot be found.
const Default = `
auth:
`

// Conf represents a Yak configuration file.
type Conf struct {
	Auth map[string]AuthEntry `yaml:"auth"`
}

// AuthEntry is an auth entry in a yak configuration file.
type AuthEntry struct {
	Options map[string]interface{} `yaml:"options,inline" required:"true"`
}

// ReadFile will read in the specified configuration file.
func ReadFile(file string) (*Conf, error) {
	var conf Conf

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, err
	}

	for _, entry := range conf.Auth {
		if err := utils.ValidateTags(entry); err != nil {
			return nil, err
		}
	}

	return &conf, nil
}

// ReadDefault will read in the specified configuration file.
func ReadDefault() (*Conf, error) {
	var conf Conf

	yamlFile, err := ioutil.ReadAll(strings.NewReader(Default))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, err
	}

	for _, entry := range conf.Auth {
		if err := utils.ValidateTags(entry); err != nil {
			return nil, err
		}
	}

	return &conf, nil
}

// FindAndLoad will search some predefined locations for a yak.yaml file.
func FindAndLoad() (*Conf, error) {
	if v := os.Getenv("YAK_CONFIG_FILE"); v != "" {
		if ok := fileExists(v); ok {
			return ReadFile(v)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(cwd, "yak.cfg")
	if ok := fileExists(filename); ok {
		return ReadFile(filename)
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	homeDir := currentUser.HomeDir
	if homeDir != "" {
		filename := filepath.Join(homeDir, ".config/yak/yak.cfg")
		if ok := fileExists(filename); ok {
			return ReadFile(filename)
		}
	}

	if ok := fileExists("/etc/yak/yak.cfg"); ok {
		return ReadFile("/etc/yak/yak.cfg")
	}

	return ReadDefault()
}

// GetAuthEntry will return a given entry.
func (r Conf) GetAuthEntry(entry string) (*AuthEntry, error) {
	if v, ok := r.Auth[entry]; ok {
		return &v, nil
	}

	return nil, fmt.Errorf("entry %s does not exist", entry)
}

// fileExists checks for the existence of a file at a given location.
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
