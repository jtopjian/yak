Actions
=======

Actions are what a step will execute on a target.

The following actions are available:

* [`apt.key`](actions/aptkey.md)
* [`apt.pkg`](actions/aptpkg.md)
* [`apt.ppa`](actions/aptppa.md)
* [`apt.source`](actions/aptsource.md)
* [`cron.entry`](actions/cronentry.md)
* [`exec`](actions/exec.md)
* [`file-upload`](actions/file-upload.md)
* [`file-download`](actions/file-download.md)
* [`file-delete`](actions/file-delete.md)

Input
-----

All actions take some form of input. Input can be specified
two different ways:

### Explicit

```yaml
action: apt.pkg
input:
  name: memcached
  state: present
  sudo: true
```

### Simplified

```yaml
action: apt.pkg name=memcached state=present sudo=true
```

Action Internals
----------------

Internally, actions are classified under two types: core
and compound. A "core" action is something like `exec` or
`file-upload`. A "compound" action leverages multiple calls
to a `core` action. You'll only ever need to know about
this if you decide to develop an action for Yak.
