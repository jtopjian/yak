package targets

import (
	"fmt"

	"github.com/jtopjian/yak/lib/config"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/openstack/clientconfig"

	"github.com/mitchellh/mapstructure"
)

// OpenStackInstances represents an openstack_instances target driver.
type OpenStackInstances struct {
	NetworkName string            `mapstructure:"network"`
	AuthEntry   string            `mapstructure:"auth"`
	Metadata    map[string]string `mapstructure:"metadata"`
	UseIPv6     bool              `mapstructure:"use_ipv6"`

	client *gophercloud.ServiceClient
}

// OpenStackAuth represents options for authenticating to OpenStack.
// This is for when clouds.yaml is not used.
type OpenStackAuth struct {
	IdentityEndpoint  string `mapstructure:"identity_endpoint"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	TenantID          string `mapstructure:"tenant_id"`
	TenantName        string `mapstructure:"tenant_name"`
	DomainID          string `mapstructure:"domain_id"`
	DomainName        string `mapstructure:"domain_name"`
	ProjectDomainID   string `mapstructure:"project_domain_id"`
	ProjectDomainName string `mapstructure:"project_domain_name"`
	UserDomainID      string `mapstructure:"user_domain_id"`
	UserDomainName    string `mapstructure:"user_domain_name"`
	Region            string `mapstructure:"region"`
}

// NewOpenStackInstances will return an OpenStackInstances.
func NewOpenStackInstances(options map[string]interface{}) (*OpenStackInstances, error) {
	var osi OpenStackInstances

	err := mapstructure.Decode(options, &osi)
	if err != nil {
		return nil, err
	}

	if osi.AuthEntry == "" {
		return nil, fmt.Errorf("auth is a required option for openstack_instances")
	}

	yakConf, err := config.FindAndLoad()
	if err != nil {
		return nil, err
	}

	authentry, err := yakConf.GetAuthEntry(osi.AuthEntry)
	if err == nil {
		// Load openstack config from yak.cfg.
		var osa OpenStackAuth
		err = mapstructure.Decode(authentry.Options, &osa)
		if err != nil {
			return nil, err
		}

		ao := &gophercloud.AuthOptions{
			IdentityEndpoint: osa.IdentityEndpoint,
			Username:         osa.Username,
			Password:         osa.Password,
			TenantID:         osa.TenantID,
			TenantName:       osa.TenantName,
			DomainID:         osa.DomainID,
			DomainName:       osa.DomainName,
		}

		if osa.ProjectDomainID != "" {
			ao.DomainID = osa.ProjectDomainID
		}

		if osa.ProjectDomainName != "" {
			ao.DomainName = osa.ProjectDomainName
		}

		if osa.UserDomainID != "" {
			ao.DomainID = osa.UserDomainID
		}

		if osa.UserDomainName != "" {
			ao.DomainName = osa.UserDomainName
		}

		providerClient, err := openstack.AuthenticatedClient(*ao)
		if err != nil {
			return nil, err
		}

		client, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{
			Region: osa.Region,
		})
		if err != nil {
			return nil, err
		}

		osi.client = client
	} else {
		// Try loading from clouds.yaml
		clientOpts := &clientconfig.ClientOpts{
			Cloud: osi.AuthEntry,
		}

		client, err := clientconfig.NewServiceClient("compute", clientOpts)
		if err != nil {
			return nil, err
		}

		osi.client = client
	}

	if osi.client == nil {
		return nil, fmt.Errorf("unable to determine openstack authentication")
	}

	return &osi, nil
}

// Discover implements the Target interface for an openstack_instances driver.
// It returns a set of hosts from an OpenStack cloud.
func (r OpenStackInstances) Discover() ([]Host, error) {
	var hosts []Host
	var filteredServers []servers.Server

	type nic struct {
		Name          string
		AccessNetwork bool
		FixedIPv4     string
		FixedIPv6     string
		FloatingIP    string
	}

	listOpts := servers.ListOpts{}
	allPages, err := servers.List(r.client, listOpts).AllPages()
	if err != nil {
		return nil, err
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, err
	}

	// First filter by metadata.
	for _, s := range allServers {
		if r.Metadata != nil {
			for iKey, iVal := range r.Metadata {
				for sKey, sVal := range s.Metadata {
					if iKey == sKey && iVal == sVal {
						filteredServers = append(filteredServers, s)
					}
				}
			}
		} else {
			filteredServers = append(filteredServers, s)
		}
	}

	// Now try to determine the network information.
	for _, s := range filteredServers {
		var nics []nic
		var accessNetworkExists bool

		for networkName, networkInfo := range s.Addresses {
			n := nic{
				Name: s.Name,
			}

			if networkName == r.NetworkName {
				n.AccessNetwork = true
				accessNetworkExists = true
			}

			for _, v := range networkInfo.([]interface{}) {
				v := v.(map[string]interface{})

				if v["OS-EXT-IPS:type"] == "fixed" {
					switch v["version"].(float64) {
					case 6:
						n.FixedIPv6 = fmt.Sprintf("[%s]", v["addr"].(string))
					default:
						n.FixedIPv4 = v["addr"].(string)
					}
				}

				if v["OS-EXT-IPS:type"] == "floating" {
					n.FloatingIP = v["addr"].(string)
				}
			}

			if accessNetworkExists {
				if n.AccessNetwork {
					// If an access network was found, reset the nics
					// slice to 0 so only the access_network is used.
					nics = nil
					nics = append(nics, n)
				}
			} else {
				nics = append(nics, n)
			}
		}

		var host Host

		// nics either contains all nics of the server
		// or just the access network.
		for _, n := range nics {
			host.Name = n.Name

			if host.Address == "" {
				if n.FixedIPv4 != "" {
					host.Address = n.FixedIPv4
				}

				if n.FloatingIP != "" {
					host.Address = n.FloatingIP
				}

				if r.UseIPv6 && n.FixedIPv6 != "" {
					host.Address = n.FixedIPv6
				}
			}
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}
