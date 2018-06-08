package yakfile

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/jtopjian/yak/lib/utils"
)

// task name format.
var taskNameRegex = regexp.MustCompile(`task::(\w)`)

// Yakfile represents the structure of a yak manifest.
type Yakfile struct {
	Vars        map[string]string     `yaml:"vars"`
	Varfiles    []string              `yaml:"varfiles"`
	Notifiers   []Step                `yaml:"notifiers"`
	Targets     map[string]Target     `yaml:"targets"`
	Connections map[string]Connection `yaml:"connections"`
	Tasks       map[string]Task       `yaml:",inline"`

	// Dir is meant for internal use only.
	// It is publicly accessible only for testing.
	Dir string `yaml:"-"`
}

// Herd represents a collection of yakfiles.
type Herd []Yakfile

// NewHerd will create a Herd from a collection of Yakfiles.
func NewHerd(files []string) (Herd, error) {
	var herd Herd

	for _, file := range files {
		yak, err := readYakfile(file)
		if err != nil {
			return nil, err
		}

		herd = append(herd, *yak)
	}

	if err := herd.ParseAndValidate(); err != nil {
		return nil, err
	}

	return herd, nil
}

// ParseAndValidate will validate, set defaults, and do other actions
// to build a functional herd.
func (r Herd) ParseAndValidate() error {
	notifiers := make(map[string]bool)
	stepNames := make(map[string]bool)
	targetNames := make(map[string]bool)
	connNames := make(map[string]bool)

	for i, yak := range r {
		if err := utils.ValidateTags(yak); err != nil {
			return err
		}

		// ensure there are no duplicate notifiers.
		for _, n := range yak.Notifiers {
			if _, ok := notifiers[n.Name]; ok {
				return fmt.Errorf("duplicate notifier detected: %s", n.Name)
			}
			notifiers[n.Name] = true
		}

		// ensure tasks are formatted correctly.
		// ensure there are no duplicate tasks.
		for taskName, task := range yak.Tasks {
			v := taskNameRegex.FindStringSubmatch(taskName)
			if v == nil {
				return fmt.Errorf("Invalid task name: %s", task)
			}

			for _, step := range task.Steps {
				v := fmt.Sprintf("%s %s", taskName, step.Name)
				if _, ok := stepNames[v]; ok {
					return fmt.Errorf("duplicate step detected: %s", v)
				}
				stepNames[v] = true
			}
		}

		// ensure there are no duplicate targets and
		// ensure the targets are valid.
		for name, info := range yak.Targets {
			if _, ok := targetNames[name]; ok {
				return fmt.Errorf("duplicate target discovery detected: %s", name)
			}
			targetNames[name] = true

			// Bootstrap the name of the target into the target struct
			info.Name = name

			// Add a special option of _dir with the current dir to the target.
			r[i].Targets[name].Options["_dir"] = yak.Dir
		}

		// ensure there are no duplicate connections and
		// ensure the connections are valid.
		for name, info := range yak.Connections {
			if _, ok := connNames[name]; ok {
				return fmt.Errorf("duplicate connection detected: %s", name)
			}
			connNames[name] = true

			// Bootstrap the name of the connection into the connection struct.
			info.Name = name

			r[i].Connections[name] = info

			// add an option of the yakfile directory.
			r[i].Connections[name].Options["_dir"] = yak.Dir
		}
	}

	for _, yak := range r {
		// ensure notifies exist
		for _, task := range yak.Tasks {
			for _, step := range task.Steps {
				if step.Notify != "" {
					if _, ok := notifiers[step.Notify]; !ok {
						return fmt.Errorf("missing notifier: %s", step.Notify)
					}
				}
			}
		}
	}

	return nil
}

// GetConnection returns the connection for a specific target.
func (r Herd) GetConnection(targetName string) (string, *Connection, error) {
	// "local" is a special connection type.
	// If it was specified, build out an explicit local connection.
	if targetName == "local" {
		local := Connection{
			Type:    "local",
			Options: map[string]interface{}{},
			Targets: []string{"local"},
		}

		return "local", &local, nil
	}

	for connName, conn := range r.ListConnections() {
		for _, t := range conn.Targets {
			if targetName == t {
				return connName, &conn, nil
			}
		}
	}

	return "", nil, fmt.Errorf("connection for target %s not found", targetName)
}

// GetHostsForStep will discover hosts for a given step and then determine
// their connection configuration. The hosts which are targeted by the step
// will be returned.
func (r Herd) GetHostsForStep(step Step) ([]Host, error) {
	var hosts []Host
	t := make(map[string]Target)

	for _, targetName := range step.Targets {
		switch targetName {
		case "_all":
			t = r.ListTargets()
		default:
			targetInfo, err := r.GetTarget(targetName)
			if err != nil {
				return nil, err
			}
			t[targetName] = *targetInfo
		}
	}

	for targetName, targetInfo := range t {
		discoveredHosts, err := targetInfo.DiscoverHosts()
		if err != nil {
			return nil, err
		}

		connName, connInfo, err := r.GetConnection(targetName)
		if err != nil {
			return nil, err
		}

		for _, host := range discoveredHosts {
			if err := host.SetConnection(connName, connInfo); err != nil {
				return nil, err
			}

			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}

// GetNotify returns a notify based on name.
func (r Herd) GetNotify(name string) (*Step, error) {
	for _, yak := range r {
		for _, n := range yak.Notifiers {
			if n.Name == name {
				return &n, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to find notifier %s", name)
}

// GetTarget returns a single target specified by name.
func (r Herd) GetTarget(targetName string) (*Target, error) {
	// "local" is a special target type.
	// If it was specified, build an explicit local target.
	if targetName == "local" {
		local := Target{
			Type:    "local",
			Options: map[string]interface{}{},
		}

		return &local, nil
	}

	allTargets := r.ListTargets()
	for name, t := range allTargets {
		if name == targetName {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("target %s not found", targetName)
}

// ListConnections returns all connections from a Herd.
func (r Herd) ListConnections() map[string]Connection {
	c := make(map[string]Connection)

	for _, yak := range r {
		for k, v := range yak.Connections {
			c[k] = v
		}
	}

	return c
}

// ListTargets returns all targets from a Herd.
func (r Herd) ListTargets() map[string]Target {
	t := make(map[string]Target)

	for _, yak := range r {
		for k, v := range yak.Targets {
			t[k] = v
		}
	}

	return t
}

// ListStepsForTask will return a task group from a Herd.
func (r Herd) ListStepsForTask(task string) ([]Step, error) {
	taskName := fmt.Sprintf("task::%s", task)

	for _, yak := range r {
		if v, ok := yak.Tasks[taskName]; ok {
			return v.Steps, nil
		}
	}

	return nil, fmt.Errorf("task %s not found", task)
}

// readYakfile will read a yaml file and parse it as a Yakfile.
func readYakfile(file string) (*Yakfile, error) {
	var yak Yakfile

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %s", file, err)
	}

	err = yaml.Unmarshal(yamlFile, &yak)
	if err != nil {
		return nil, fmt.Errorf("unable to parse YAML in %s: %s", file, err)
	}

	dir := filepath.Dir(file)
	yak.Dir = dir

	return &yak, nil
}
