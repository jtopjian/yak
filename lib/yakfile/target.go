package yakfile

import (
	"fmt"
	"sync"

	"github.com/jtopjian/yak/lib/targets"
	"github.com/jtopjian/yak/lib/utils"
)

// Target represents a target of hosts to discover.
type Target struct {
	Name    string                 `yaml:"-"`
	Auth    string                 `yaml:"auth"`
	Type    string                 `yaml:"type" required:"true"`
	Options map[string]interface{} `yaml:"options"`

	Hosts []Host `yaml:"-"`
	mux   sync.Mutex
}

// UnmarshalYAML is a custom unmarshaler to help initialize and
// validate a target.
func (r *Target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp Target
	var s struct {
		tmp `yaml:",inline"`
	}

	err := unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to parse YAML: %s", err)
	}

	*r = Target(s.tmp)

	if err := utils.ValidateTags(r); err != nil {
		return err
	}

	// If options werent' specified, create an empty map.
	if r.Options == nil {
		r.Options = make(map[string]interface{})
	}

	// Add the auth entry to options.
	r.Options["auth"] = r.Auth

	return nil
}

// DiscoverHosts will run Discover and return the hosts.
func (r *Target) DiscoverHosts() ([]Host, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.Hosts != nil {
		return r.Hosts, nil
	}

	t, err := targets.New(r.Type, r.Options)
	if err != nil {
		return nil, err
	}

	discoveredHosts, err := t.Discover()
	if err != nil {
		return nil, err
	}

	var hosts []Host
	for _, host := range discoveredHosts {
		hosts = append(hosts, Host{
			TargetName: r.Name,
			Name:       host.Name,
			Address:    host.Address,
		})
	}

	r.Hosts = hosts

	return r.Hosts, nil
}
