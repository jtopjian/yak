Yak Documentation
=================

Yak Files
---------

A Yak file is a YAML file which tells Yak what to do. Multiple Yak files
can exist in a single directory. A collection of Yak files is called a Herd,
but this is as far as silly naming schemes go.

See the [yak files](yakfiles.md) doc for more details.

Targets
-------

Targets are remote hosts. They can be dynamically discovered or statically
defined. Tasks can be applied to one, multiple, or all defined targets.

See the [targets](targets.md) doc for more details.

Connections
-----------

Connections define how to connect to a target; for example, `ssh`.
Connections can be applied to one or more Targets.

See the [connections](connections.md) doc for more details.

Tasks, Steps, and Notifiers
---------------------------

A task is a set of steps that `yak` will run.
Each step can be applied to one, multiple, or all targets. Steps
are run in a top-to-bottom sequence.

Notifiers are special steps that are only ever run when a step
notifies them.

See the [tasks](tasks.md) doc for more details.

Actions
-------

Actions are what the steps will execute on the target. An Action
can be something generic such as executing an arbitrary command
or a more complex action such as managing the state of an `apt`
package.

See the [actions](actions.md) doc for more details.

Configuration
-------------

Yak itself has a configuration file called `yak.cfg`. See the
[config](config.md) doc for more details.
