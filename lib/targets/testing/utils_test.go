package testing

import (
	"testing"

	"github.com/jtopjian/yak/lib/targets"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverTargets(t *testing.T) {
	target := yakfile.Target{
		Type: "textfile",
		Options: map[string]interface{}{
			"file": "fixtures/hosts.txt",
		},
	}

	expected := &targets.DiscoveredTargets{
		Type: "textfile",
		Hosts: []targets.Host{
			targets.Host{Address: "host1.example.com", Name: "host1.example.com"},
			targets.Host{Address: "host2.example.com", Name: "host2.example.com"},
			targets.Host{Address: "192.168.100.1", Name: "192.168.100.1"},
			targets.Host{Address: "fe80::f816:3eff:fe8c:c73a", Name: "fe80::f816:3eff:fe8c:c73a"},
		},
	}

	actual, err := targets.DiscoverTargets(target)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)

}
