apt.ppa
-------

`apt.ppa` will manage an `apt`-based PPA.

### example

```
task::install_redis:
  - name: install redis ppa
    action: apt.ppa
    input:
      name: chris-lea/redis-server
      state: present
```

### options

* `name` (required) - The name of the PPA.

* `state` (optional) - The state of the PPA. This can either be:
  a version number, `present` or `absent`. Defaults to `present`.

* `refresh` (optional) - Whether to perform an `apt-get update`
  when the state changes. Defaults to `false`.

* `sudo` - Whether or not sudo is required. Valid values are
  `true` or `false`.

* `timeout` - How long the command should run before it times out.
