package targets

import (
	"fmt"
)

// Target is an interface which specifies what drivers
// must implement.
type Target interface {
	Discover() ([]Host, error)
}

// Host represents a host returned by a target driver.
type Host struct {
	Name    string
	Address string
}

// New will return a target based on a given target driver.
func New(targetType string, options map[string]interface{}) (Target, error) {
	if targetType == "" {
		return nil, fmt.Errorf("a target type was not specified")
	}

	switch targetType {
	case "local":
		return NewLocal(options)
	case "lxd_containers":
		return NewLXDContainers(options)
	case "openstack_instances":
		return NewOpenStackInstances(options)
	case "textfile":
		return NewTextFile(options)
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}

	return nil, nil
}
