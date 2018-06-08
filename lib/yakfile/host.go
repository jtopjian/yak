package yakfile

import (
	"sync"

	"github.com/jtopjian/yak/lib/connections"
)

// Host represents a host discovered by a target.
type Host struct {
	Name           string
	Address        string
	TargetName     string
	ConnectionName string
	ConnectionType string
	Connection     connections.Connection

	mux sync.Mutex
}

func (r *Host) SetConnection(connName string, connInfo *Connection) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.Connection == nil {
		r.ConnectionName = connName
		r.ConnectionType = connInfo.Type

		// In order to create the connection, a target and connection
		// must be glued together.
		switch connInfo.Type {
		case "lxd":
			connInfo.Options["host"] = r.Name
		default:
			connInfo.Options["host"] = r.Address
		}

		conn, err := connections.New(connInfo.Type, connInfo.Options)
		if err != nil {
			return err
		}

		r.Connection = conn
	}

	return nil
}
