package targets

import (
	"fmt"

	"github.com/jtopjian/yak/lib/config"
	"github.com/jtopjian/yak/lib/shared"

	lxd "github.com/lxc/lxd/client"
	lxd_api "github.com/lxc/lxd/shared/api"

	"github.com/mitchellh/mapstructure"
)

// LXDContainers represents an lxd_containers target driver.
type LXDContainers struct {
	AuthEntry string            `mapstructure:"auth"`
	Config    map[string]string `mapstructure:"config"`
	UseIPv6   bool              `mapstructure:"use_ipv6"`
	Interface string            `mapstructure:"interface"`

	client lxd.ContainerServer
}

// NewLXDContainers will return an LXDContainers.
func NewLXDContainers(options map[string]interface{}) (*LXDContainers, error) {
	var lxdc LXDContainers
	var lxdAuth shared.LXDAuth

	err := mapstructure.Decode(options, &lxdc)
	if err != nil {
		return nil, err
	}

	if lxdc.AuthEntry == "" {
		return nil, fmt.Errorf("auth is required for lxd_containers")
	}

	yakConf, err := config.FindAndLoad()
	if err != nil {
		return nil, err
	}

	lxda, err := yakConf.GetAuthEntry(lxdc.AuthEntry)
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(lxda.Options, &lxdAuth)
	if err != nil {
		return nil, err
	}

	if lxdc.Interface == "" {
		lxdc.Interface = "eth0"
	}

	client, err := lxdAuth.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate to LXD: %s", err)
	}

	lxdc.client = client

	return &lxdc, nil
}

// Discover implements the Target interface for an lxd_containers driver.
// It returns a set of containers from an LXD server.
func (r LXDContainers) Discover() ([]Host, error) {
	var hosts []Host
	var filteredContainers []lxd_api.Container

	inet := "inet"
	if r.UseIPv6 {
		inet = "inet6"
	}

	containers, err := r.client.GetContainers()
	if err != nil {
		return nil, fmt.Errorf("unable to get containers: %s", err)
	}

	// Filter the containers by tags.
	for _, container := range containers {
		if !container.IsActive() {
			continue
		}

		if r.Config != nil {
			for key, val := range r.Config {
				cVal, ok := container.Config[key]
				if !ok {
					continue
				}

				if val == cVal {
					filteredContainers = append(filteredContainers, container)
				}
			}
		} else {
			filteredContainers = append(filteredContainers, container)
		}
	}

	// Determine network configuration of containers.
	for _, container := range filteredContainers {
		host := Host{
			Name: container.Name,
		}

		cstate, _, err := r.client.GetContainerState(container.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to get container state for %s: %s", container.Name, err)
		}

		for i, network := range cstate.Network {
			if i != r.Interface {
				continue
			}

			for _, ip := range network.Addresses {
				if ip.Family == inet {
					host.Address = ip.Address
				}
			}
		}

		if r.UseIPv6 && host.Address != "" {
			host.Address = fmt.Sprintf("[%s]", host.Address)
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}
