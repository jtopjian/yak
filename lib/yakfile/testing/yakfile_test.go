package testing

import (
	"testing"

	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/stretchr/testify/assert"
)

var nilMap = make(map[string]interface{})
var iTrue = true

var expectedSample = yakfile.Yakfile{
	Dir: "fixtures",
	Vars: map[string]string{
		"foo": "bar",
		"bar": "baz",
	},
	Varfiles: []string{
		"/path/to/file.yaml",
	},
	Targets: map[string]yakfile.Target{
		"openstack-yyc": yakfile.Target{
			Type: "openstack_instances",
			Options: map[string]interface{}{
				"cloud": "yyc",
				"tags":  []interface{}{"memcached"},
				"auth":  "",
			},
		},
		"textfile": yakfile.Target{
			Type: "textfile",
			Options: map[string]interface{}{
				"file": "hosts.txt",
				"auth": "",
			},
		},
	},
	Connections: map[string]yakfile.Connection{
		"local": yakfile.Connection{
			Name: "local",
			Type: "local",
			Options: map[string]interface{}{
				"auth": "",
			},
			Targets: []string{
				"textfile",
			},
		},
		"ssh": yakfile.Connection{
			Name: "ssh",
			Type: "ssh",
			Options: map[string]interface{}{
				"private_key": "/path/to/id_rsa",
				"port":        22,
				"auth":        "",
			},
			Targets: []string{
				"openstack-yyc",
			},
		},
	},
	Notifiers: []yakfile.Step{
		yakfile.Step{
			Name:   "apt-get update",
			Action: "exec",
			Input: map[string]interface{}{
				"cmd": "apt-get update -qq",
			},
			Targets: []string{
				"_all",
			},
		},
		yakfile.Step{
			Name:   "restart memcached",
			Action: "exec",
			Input: map[string]interface{}{
				"cmd": "service memcached restart",
			},
			Targets: []string{
				"_all",
			},
		},
	},
	Tasks: map[string]yakfile.Task{
		"task::stats": yakfile.Task{
			Defaults: yakfile.TaskDefaults{
				Sudo:  &iTrue,
				Limit: 5,
			},
			Steps: []yakfile.Step{
				yakfile.Step{
					Name:   "collect memcached stats",
					Action: "cmd",
					Input: map[string]interface{}{
						"cmd":  "foo -bar",
						"sudo": true,
					},
					Targets: []string{
						"textfile",
					},
				},
			},
		},
		"task::state": yakfile.Task{
			Defaults: yakfile.TaskDefaults{
				Sudo:    &iTrue,
				Limit:   5,
				Targets: []string{"textfile"},
			},
			Steps: []yakfile.Step{
				yakfile.Step{
					Name:   "install apt repo",
					Action: "apt.repo",
					Input: map[string]interface{}{
						"name":   "repcached",
						"source": "abc",
						"sudo":   true,
					},
					Targets: []string{
						"openstack-yyc",
					},
					Notify: "apt-get update",
				},
				yakfile.Step{
					Name:   "install memcached",
					Action: "apt.get",
					Input: map[string]interface{}{
						"name":    "repcached",
						"version": "foo",
						"sudo":    true,
					},
					Targets: []string{
						"textfile",
					},
				},
				yakfile.Step{
					Name:   "configure memory limits",
					Action: "file.line",
					Input: map[string]interface{}{
						"name":  "/etc/memcached/memcached.conf",
						"line":  "-m 64",
						"match": "^-m",
						"sudo":  false,
					},
					Targets: []string{
						"openstack-yyc",
					},
					Notify: "restart memcached",
				},
			},
		},
	},
}

var expectedAnother = yakfile.Yakfile{
	Dir: "fixtures",
	Tasks: map[string]yakfile.Task{
		"task::state": yakfile.Task{
			Defaults: yakfile.TaskDefaults{
				Limit: 5,
			},
			Steps: []yakfile.Step{
				yakfile.Step{
					Name:   "do something",
					Action: "exec",
					Input: map[string]interface{}{
						"cmd": "echo hello",
					},
					Targets: []string{
						"_all",
					},
					Notify: "apt-get update",
				},
			},
		},
	},
}

func TestHerd(t *testing.T) {
	files := []string{
		"fixtures/sample.yaml",
		"fixtures/another.yaml",
	}

	expectedSample.Targets["openstack-yyc"].Options["_dir"] = "fixtures"
	expectedSample.Targets["textfile"].Options["_dir"] = "fixtures"

	expectedSample.Connections["local"].Options["_dir"] = "fixtures"
	expectedSample.Connections["ssh"].Options["_dir"] = "fixtures"

	expectedHerd := yakfile.Herd{expectedSample, expectedAnother}

	actualHerd, err := yakfile.NewHerd(files)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedHerd, actualHerd)

}

func TestBadYakfiles(t *testing.T) {
	var testCases = []struct {
		file []string
		err  string
	}{
		{
			[]string{"fixtures/bad-duplicate-step.yaml"},
			"duplicate step detected: task::state install memcached",
		},
		{
			[]string{"fixtures/bad-duplicate-notifier.yaml"},
			"duplicate notifier detected: apt-get update",
		},
		{
			[]string{"fixtures/bad-missing-notifier.yaml"},
			"missing notifier: foobar",
		},
	}

	for _, v := range testCases {
		_, err := yakfile.NewHerd(v.file)
		assert.Equal(t, err.Error(), v.err)
	}
}
