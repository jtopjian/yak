apt.source
----------

`apt.source` will manage an `apt`-based source file.

### example

```
task::install_rabbitmq:
  - name: install rabbit apt source
    action: apt.source
    input:
      name: rabbitmq
      uri: http://www.rabbitmq.com/debian/
      distribution: testing
      component: main
      include_src: true
```

### options

* `name` (required) - A descriptive name of the source file.

* `state` (optional) - The state of the PPA. This can either be:
  a version number, `present` or `absent`. Defaults to `present`.

* `uri` (required) - The URI of the apt repository.

* `distribution` (required) - The distribution of the apt repository.

* `component` (optional) - The component of the apt repository.

* `include_src` (optional) - Whether to include the source repository
  as well. Defaults to `false`.

* `refresh` (optional) - Whether to perform an `apt-get update`
  when the state changes. Defaults to `false`.

* `sudo` - Whether or not sudo is required. Valid values are
  `true` or `false`.

* `timeout` - How long the command should run before it times out.
