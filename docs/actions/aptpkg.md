apt.pkg
-------

`apt.pkg` will manage an `apt`-based package.

### example

```
task::install_memcached:
  - name: install memcached
    action: apt.pkg
    input:
      name: memcached
      state: present
```

### options

* `name` (required) - The name of the package.

* `state` (optional) - The state of the package. This can either be:
  a version number, `present`, `latest`, or `absent`. Defaults to
  `present`.

* `sudo` - Whether or not sudo is required. Valid values are
  `true` or `false`.

* `timeout` - How long the command should run before it times out.

