package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func actionPlan(c *cli.Context) error {
	// Make sure a task name was specified.
	taskName, err := getTask(c)
	if err != nil {
		return err
	}

	herd, err := newHerd(c)
	if err != nil {
		return err
	}

	// Get the requested steps from the herd.
	steps, err := herd.ListStepsForTask(taskName)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Print the plan.
	title.Printf("yak plan - %s\n", taskName)
	fmt.Println("")

	// For each step in the task.
	for _, step := range steps {
		// Print the step name.
		cyan.Println(step.Name)

		// Get the hosts required for the step.
		stepHosts, err := herd.GetHostsForStep(step)
		if err != nil {
			return err
		}

		// For each host
		for _, host := range stepHosts {
			// Print the hosts which the step will be run on.
			fmt.Fprintln(w, magenta.Sprintf("  - host=%s\ttarget=\"%s\"\tconnection=\"%s\"",
				host.Name, host.TargetName, host.ConnectionName))
		}

		// Print any notification actions to run.
		if step.Notify != "" {
			fmt.Fprintln(w, blue.Sprintf("  - notify: %s", step.Notify))
		}
		w.Flush()
	}

	return nil
}
