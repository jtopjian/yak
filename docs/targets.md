Targets
=======

Targets define the hosts which steps will be applied to. Targets
can be dynamically discovered or statically defined.

Targets are defined as follows:

```yaml
targets:
  name-of-target:
    type: target-driver
    options:
      key: value
      key: value
```

Target Drivers
--------------

Yak currently supports the following Target Drivers:

### local

The `local` driver will target the local host. To target the
localhost, specify a target of `local` in the *step*.

### lxd_containers

The `lxd_containers` driver will target LXD containers.

#### example

```yaml
targets:
  name-of-target:
    type: lxd_containers
    options:
      auth: some-name
      config:
        user.yak: memcached
```

#### options

* `auth` (required) - The Yak authentication entry in the Yak
  configuration file. See the [config](config.md) docs for more
  information.

* `config` (optional) - A set of key/value pairs to filter LXD
  containers by their configuration data. For example, use
  `lxc config set user.yak memcached` to set a custom config tag.

* `use_ipv6` (optional) - Connect via IPv6. Defaults to `false`.

* `interface` (optional) - The network interface to connect via.
  Defaults to `eth0`.

### textfile

The `textfile` driver will read hosts defined in a plain text file.

#### example

```yaml
targets:
  name-of-target:
    type: textfile
    options:
      file: /path/to/file.txt
```

#### options

* `file` (required) - The text file which defines the hosts. Each
line of the text file must contain only the resolvable name or IP
address of the host. An example file is:

```
# comment
host1.example.com
host2.example.com
// host3.example.com
192.168.100.1
fe80::f816:3eff:fe8c:c73a
```

### openstack_instances

The `openstack_instances` driver will query an OpenStack cloud
for instance which match the given `metadata`. If no `metadata`
is specified, then all instances are returned.

Yak uses the following order of precedence to determine the IP
address of the instance:

1. Fixed IPv4
2. Floating IP
3. Fixed IPv6 if `use_ipv6` was specified

#### example

```yaml
targets:
  name-of-target:
    type: openstack_instances
    options:
      cloud: cloud-yaml-entry
      metadata:
        key: value
        key: value
      network_name: accessible-network-name
      use_ipv6: true/false
```

#### options

* `cloud` (required) - Specifies a cloud defined in a `clouds.yaml` file.
For more information on `clouds.yaml`, see
[here](https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html).

* `metadata` (optional) - key/value pairs to match with metadata configured
on the instances.

* `network_name` (optional) - The name of a network to connect to the
instances. If no name is specified, the first detected NIC of the instance
will be used.

* `use_ipv6` (optional) - Whether or not to connect to the instances via IPv6.
