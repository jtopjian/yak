package testing

import (
	"os"
	"testing"

	"github.com/jtopjian/yak/lib/config"

	"github.com/stretchr/testify/assert"
)

func TestConf_BasicAuth(t *testing.T) {
	c, err := config.Read("fixtures/yak.cfg")
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]config.AuthEntry{
		"lxd1": config.AuthEntry{
			Options: map[string]interface{}{
				"remote":                    "remote",
				"address":                   "1.2.3.4",
				"port":                      8443,
				"password":                  "foobar",
				"accept_remote_certificate": true,
			},
		},
	}

	assert.Equal(t, expected, c.Auth)

	e, err := c.GetAuthEntry("lxd1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected["lxd1"], *e)
}

func TestConf_FindAndLoad(t *testing.T) {
	os.Setenv("YAK_CONFIG_FILE", "fixtures/yak.cfg")

	c, err := config.FindAndLoad()
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]config.AuthEntry{
		"lxd1": config.AuthEntry{
			Options: map[string]interface{}{
				"remote":                    "remote",
				"address":                   "1.2.3.4",
				"port":                      8443,
				"password":                  "foobar",
				"accept_remote_certificate": true,
			},
		},
	}

	assert.Equal(t, expected, c.Auth)

	e, err := c.GetAuthEntry("lxd1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected["lxd1"], *e)
}
