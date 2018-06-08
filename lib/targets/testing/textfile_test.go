package testing

import (
	"testing"

	"github.com/jtopjian/yak/lib/targets"
	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

func TestTextFile(t *testing.T) {
	config := yakfile.Target{
		Type: "textfile",
		Options: map[string]interface{}{
			"file": "fixtures/hosts.txt",
		},
	}

	textfile, err := targets.New(config.Type, config.Options)
	if err != nil {
		t.Fatal(err)
	}

	expected := []targets.Host{
		targets.Host{Address: "host1.example.com", Name: "host1.example.com"},
		targets.Host{Address: "host2.example.com", Name: "host2.example.com"},
		targets.Host{Address: "192.168.100.1", Name: "192.168.100.1"},
		targets.Host{Address: "fe80::f816:3eff:fe8c:c73a", Name: "fe80::f816:3eff:fe8c:c73a"},
	}

	actual, err := textfile.Discover()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}
