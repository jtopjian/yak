Yak Files
=========

A Yak file is a YAML file which tells `yak` what to do. Multiple
Yak files can exist in the same directory. Collectively, this is
called a Herd.

Multiple Yak files can define targets, connections, and tasks,
though each target, connection, and task must be unique to
the Herd as a whole.

Format
------

A Yak file is a standard YAML file. Targets, Connections, Tasks,
and Notifiers all take a defined format, though:

### Targets

Targets are defined as follows:

```yaml
targets:
  name-of-target:
    type: target-driver
    options:
      key: value
      key: value
```

The `name-of-target` can be arbitrary and should describe the group
of targets you are referring to.

Each `target-driver` accepts a set of options. See the
[targets](targets.md) documentation for information on the available
drivers.

### Connections

Connections are defined as follows:

```yaml
connections:
  name-of-connection:
    type: connection-driver
    options:
      key: value
      key: value
    targets:
      - name-of-target
      - name-of-target
```

The `name-of-connection` can be arbitrary and should describe the
connection method.

Connections must target at least one Target.

Each `connection-driver` accepts a set of options. See the
[connections](connections.md) documentation for information on the
available drivers.

### Notifiers

Notifiers are single steps. They are defined as follows:

```yaml
notifiers:
  - name: name-of-steps
    action: action-to-run
    input:
      key: value
      key: value

  - name: name-of-other-steps
    action: some-other-action
    input:
      key: value
      key: value
```

See the [actions](actions.md) documentation for information on the
available actions.

### Tasks

A Task is a group of step. Yak will run steps based on
the task you specify on the command-line.

It is not possible to execute a single step outside of a Task.
If you want to execute a single step, then make a Task with
only one step.

Tasks are defined as follows:

```yaml
task::task-name:
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

Tasks must be specified in the above format, including
the `task::`.

A Step can target specific Targets. If no Targets are specified,
the Step will be run on *all* targets.

A Step can define a single Notification step.
