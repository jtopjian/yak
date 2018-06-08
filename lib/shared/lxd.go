package shared

import (
	"encoding/pem"
	"fmt"
	"os"

	lxd "github.com/lxc/lxd/client"
	lxd_config "github.com/lxc/lxd/lxc/config"
	"github.com/lxc/lxd/shared"
	lxd_api "github.com/lxc/lxd/shared/api"
)

const (
	LXDDefaultRemote = "local"
	LXDDefaultPort   = 8443
)

// LXDAuth represents options used to connect to an LXD server.
type LXDAuth struct {
	Remote             string `mapstructure:"remote"`
	Address            string `mapstructure:"address"`
	Port               int    `mapstructure:"port"`
	Password           string `mapstructure:"password"`
	Scheme             string `mapstructure:"scheme"`
	ConfigDir          string `mapstructure:"config_dir"`
	AcceptRemoteCert   bool   `mapstructure:"accept_remote_certificate"`
	GenerateClientCert bool   `mapstructure:"generate_client_certificate"`
}

// Authenticate will authenticate to an LXD server and return a client.
func (r LXDAuth) Authenticate() (lxd.ContainerServer, error) {
	// Set some default values.
	if r.Remote == "" {
		r.Remote = LXDDefaultRemote
	}

	if r.Scheme == "" {
		r.Scheme = "https"
	}

	if r.Port == 0 {
		r.Port = LXDDefaultPort
	}

	if r.ConfigDir == "" {
		r.ConfigDir = os.ExpandEnv("$HOME/.config/lxc")
	}

	// Connect to the LXD server
	var lxdConfig *lxd_config.Config
	if conf, err := lxd_config.LoadConfig(r.ConfigDir); err != nil {
		lxdConfig = &lxd_config.DefaultConfig
		lxdConfig.ConfigDir = r.ConfigDir
	} else {
		lxdConfig = conf
	}

	var addr string
	switch r.Scheme {
	case "https":
		addr = fmt.Sprintf("https://%s:%s", r.Address, r.Port)
	case "unix":
		addr = fmt.Sprintf("unix:%s", r.Address)
	}

	lxdConfig.Remotes[r.Remote] = lxd_config.Remote{Addr: addr}

	if r.Scheme == "https" {
		if r.GenerateClientCert {
			if err := lxdConfig.GenerateClientCertificate(); err != nil {
				return nil, fmt.Errorf("could not genrate a client certificate: %s", err)
			}
		}

		serverCertf := lxdConfig.ServerCertPath(r.Remote)
		if !shared.PathExists(serverCertf) {
			_, err := lxdConfig.GetContainerServer(r.Remote)
			if err != nil {
				if r.AcceptRemoteCert {
					if err := lxdGetRemoteCertificate(lxdConfig, r.Remote); err != nil {
						return nil, fmt.Errorf("could not obtain LXD server certificate: %s", err)
					}
				} else {
					err := fmt.Errorf("unable to communicate with LXD server. Either set " +
						"accept_server_certificate to true or add the LXD server out of band of " +
						"yak and try again.")
					return nil, err
				}
			}
		}

		client, err := lxdConfig.GetContainerServer(r.Remote)
		if err != nil {
			return nil, fmt.Errorf("could not create an LXD client: %s", err)
		}

		err = lxdAuthenticateToServer(client, r.Remote, r.Password)
		if err != nil {
			return nil, fmt.Errorf("could not authenticate to LXD server: %s", err)
		}
	}

	client, err := lxdConfig.GetContainerServer(r.Remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create LXD client: %s", err)
	}

	return client, nil
}

func lxdGetRemoteCertificate(config *lxd_config.Config, remote string) error {
	addr := config.Remotes[remote]
	certificate, err := shared.GetRemoteCertificate(addr.Addr)
	if err != nil {
		return err
	}

	serverCertDir := config.ConfigPath("servercerts")
	if err := os.MkdirAll(serverCertDir, 0750); err != nil {
		return fmt.Errorf("Could not create server cert dir: %s", err)
	}

	certf := fmt.Sprintf("%s/%s.crt", serverCertDir, remote)
	certOut, err := os.Create(certf)
	if err != nil {
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	certOut.Close()

	return nil
}

func lxdAuthenticateToServer(client lxd.ContainerServer, remote string, password string) error {
	srv, _, err := client.GetServer()
	if srv.Auth == "trusted" {
		return nil
	}

	req := lxd_api.CertificatesPost{
		Password: password,
	}
	req.Type = "client"

	err = client.CreateCertificate(req)
	if err != nil {
		return fmt.Errorf("Unable to authenticate with remote server: %s", err)
	}

	_, _, err = client.GetServer()
	if err != nil {
		return err
	}

	return nil
}
