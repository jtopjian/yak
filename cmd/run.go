package main

import (
	"context"

	"github.com/jtopjian/yak/lib/yakfile"

	"github.com/remeh/sizedwaitgroup"
	"github.com/urfave/cli"
)

func actionRun(c *cli.Context) error {
	log := getLogger()

	// Make sure a task was specified.
	taskName, err := getTask(c)
	if err != nil {
		return err
	}

	herd, err := newHerd(c)
	if err != nil {
		return err
	}

	log.Infof("===> Task: %s", taskName)

	// Get the steps of the task.
	steps, err := herd.ListStepsForTask(taskName)
	if err != nil {
		return err
	}

	// For each step in the task.
	for i, step := range steps {
		log.Infof("===> Step [%02d/%02d]: %s", i+1, len(steps), step.Name)

		stepHosts, err := herd.GetHostsForStep(step)
		if err != nil {
			return err
		}

		swg := sizedwaitgroup.New(step.Limit)
		for _, host := range stepHosts {
			// Create a goroutine for each task execution.
			// Limit the amount of goroutines running at a
			// time by the task.Limit setting.
			swg.Add()
			go func(step yakfile.Step, host yakfile.Host) {
				defer swg.Done()

				ctx := context.WithValue(context.Background(), "log", log)
				changed, err := runStep(ctx, host, step)

				// if a change was made and there was no error,
				// run a notifier if one exists.
				if changed && err == nil && step.Notify != "" {
					n, err := herd.GetNotify(step.Notify)
					if err != nil {
						log.Errorf("unable to determine notify for step %s", step.Name)
						return
					}

					if err == nil {
						ctx := context.WithValue(context.Background(), "log", log)
						runStep(ctx, host, *n)
					}
				}
			}(step, host)
		}
		swg.Wait()
		log.Info("")
	}

	return nil
}
