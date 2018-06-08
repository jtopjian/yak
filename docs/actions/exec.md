exec
----

`exec` will execute an arbitrary command on a target.

### example

```yaml
action: exec
  input:
    cmd: apt-get install -y memcached
    unless: "dpkg -l | grep -q memcached"
    sudo: true
```

### options

* `cmd` (required) - The command to run.

* `dir` (optional) - The directory to execute the command in.

* `env` (optional) - A list of `key=val` environment variables.

* `sudo` (optional) - Whether or not to use `sudo` to execute the command.

* `timeout` (optional) - How long the command should run before it times out.

* `unless` (optional) - If set, this command will be run first. If the exit code
  is `0`, then `cmd` will *not* be run.
