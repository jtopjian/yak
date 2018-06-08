package yakfile

import (
	"fmt"
	"strings"

	"github.com/jtopjian/yak/lib/utils"
)

// Task represents a task.
type Task struct {
	Defaults TaskDefaults `yaml:"defaults"`
	Steps    []Step       `yaml:"steps"`
}

// UnmarshalYAML is a custom unmarshaler to help initialize and
// validate a task.
func (r *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp Task
	var s struct {
		tmp `yaml:",inline"`
	}

	err := unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to parse YAML: %s", err)
	}

	*r = Task(s.tmp)

	if err := utils.ValidateTags(r); err != nil {
		return err
	}

	// Apply defaults to all steps in the task.
	for i, step := range r.Steps {
		if step.Targets[0] == "_all" {
			if r.Defaults.Targets != nil {
				r.Steps[i].Targets = r.Defaults.Targets
			}
		}

		if step.Limit == 0 {
			step.Limit = r.Defaults.Limit
		}

		if _, ok := step.Input["sudo"]; !ok {
			if r.Defaults.Sudo != nil {
				step.Input["sudo"] = *r.Defaults.Sudo
			}
		}
	}

	return nil
}

// TaskDefaults represents defaults which should be applied to all
// steps in the task.
type TaskDefaults struct {
	Limit   int      `yaml:"limit" default:"5"`
	Sudo    *bool    `yaml:"sudo"`
	Targets []string `yaml:"targets"`
}

// UnmarshalYAML is a custom unmarshaler to help initialize and
// validate task defaults.
func (r *TaskDefaults) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp TaskDefaults
	var s struct {
		tmp `yaml:",inline"`
	}

	err := unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to parse YAML: %s", err)
	}

	*r = TaskDefaults(s.tmp)

	if err := utils.ValidateTags(r); err != nil {
		return err
	}

	return nil
}

// Step represents the structure of a step within a task.
type Step struct {
	Action  string                 `yaml:"action" required:"true"`
	Input   map[string]interface{} `yaml:"input"`
	Limit   int                    `yaml:"limit"`
	Name    string                 `yaml:"name" required:"true"`
	Notify  string                 `yaml:"notify"`
	Targets []string               `yaml:"targets"`
	Timeout int                    `yaml:"timeout"`
}

// UnmarshalYAML is a custom unmarshaler to help initialize and
// validate a step.
func (r *Step) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp Step
	var s struct {
		tmp `yaml:",inline"`
	}

	err := unmarshal(&s)
	if err != nil {
		return fmt.Errorf("unable to parse YAML: %s", err)
	}

	*r = Step(s.tmp)

	if err := utils.ValidateTags(r); err != nil {
		return err
	}

	// Determine if an action includes parameters.
	parts := strings.SplitN(r.Action, " ", 2)
	r.Action = parts[0]
	if len(parts) == 2 {
		params, err := utils.ParseSimplified(parts[1])
		if err != nil {
			return fmt.Errorf("error parsing action: %s", err)
		}

		r.Input = params
	}

	// If no targets were specified, add an entry for all.
	if len(r.Targets) == 0 {
		r.Targets = []string{"_all"}
	}

	return nil
}
