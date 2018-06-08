apt.key
-------

`apt.key` will manage an apt key.

### example

```
task::install_rabbitmq
  - name: install rabbitmq key
    action: apt-key
    input:
      name: 6026DFCA
      remote_key_file: https://www.rabbitmq.com/rabbitmq-release-signing-key.asc
```

### options

* `name` (required) - The short ID of the key.

* `state` (optional) - The state of the key. This can either be:
  `present` or `absent`. Defaults to `present`.

* `remote_key_file` (optional) - The URL to a public key. Cannot be
  used with `key_server`.

* `key_server` (optional) - The remote server to obtain the key from.
  Cannot be used with `remote_key_file`.

* `sudo` - Whether or not sudo is required. Valid values are
  `true` or `false`.

* `timeout` - How long the command should run before it times out.
