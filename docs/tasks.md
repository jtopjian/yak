Tasks and Steps
===============

A Step is a unit of work applied to a target. Steps are
grouped into Tasks. All Steps must be part of a Task.
Tasks are unique to a Herd, but a Step can be repeated
in Tasks.

A Task is defined as follows:

```yaml
task::task-name:
  defaults:
    sudo: true

  steps:
    - name: name-of-step
      action: action-to-run
      input:
        key: value
        key: value
      notify: name-of-notify-step
      targets:
        - name-of-target

    - name: name-of-other-step
      action: some-other-action
      input:
        key: value
        key: value
      notify: name-of-other-notify-step
```

Task Defaults
-------------
You can specify the following defaults in a task. These values
will be inherited by all steps.

* `limit` (optional) - Limits the number of targets
  being executed at once.

* `sudo` (optional) - Run the action with "sudo".

* `targets` (optional) - A list of targets to run the step
  on.

Step Attributes
---------------
Steps have the following attributes:

* `name` (required) - An arbitrary name for the step.

* `action` (required) - The action to perform. See
  [actions](actions.md) for a list of actions.

* `input` (required) - The input of the action.

* `notify` (optional) - A notify step to run upon success.

* `targets` (optional) - A list of targets to run the step
  on. If not specified, the step will be run on all targets.

* `limit` (optional) - Limits the number of targets being
  executed at once. If not specified, a limit of `5` is used.

Notifiers
---------
Notifers are single steps which can only be triggered by another
step.

They are defined as follows:

```yaml
notifiers:
  - name: name-of-step
    action: action-to-run
    input:
      key: value
      key: value

  - name: name-of-other-step
    action: some-other-action
    input:
      key: value
      key: value
```
