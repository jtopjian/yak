package yakfile

import (
	"fmt"

	"github.com/jtopjian/yak/lib/utils"
)

// Connection represents a connection configuration
type Connection struct {
	Name    string                 `yaml:"-"`
	Auth    string                 `yaml:"auth"`
	Type    string                 `yaml:"type" required:"true"`
	Options map[string]interface{} `yaml:"options"`
	Targets []string               `yaml:"targets" required:"true"`
}

// UnmarshalYAML is a custom unmarshaler to help initialize and
// validate a connection.
func (r *Connection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp Connection
	var s struct {
		tmp `yaml:",inline"`
	}

	err := unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to parse YAML: %s", err)
	}

	*r = Connection(s.tmp)

	if err := utils.ValidateTags(r); err != nil {
		return err
	}

	// If options weren't specified, create an empty map.
	if r.Options == nil {
		r.Options = make(map[string]interface{})
	}

	// Add the auth entry to options.
	r.Options["auth"] = r.Auth

	return nil
}
