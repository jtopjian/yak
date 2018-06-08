package targets

// Local represents a local target driver.
type Local struct{}

// NewLocal will return a Local.
func NewLocal(options map[string]interface{}) (*Local, error) {
	var local Local

	// all options of local are ignored.
	return &local, nil
}

// Discover implements the Target interface for a local driver.
// It returns nothing as it's meant to be used with the `local`
// connection driver to run commands on the local host.
func (r Local) Discover() ([]Host, error) {
	var hosts []Host

	hosts = append(hosts, Host{Name: "local", Address: "local"})

	return hosts, nil
}
