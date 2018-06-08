Yak Configuration
=================

Yak itself is managed through a configuration file called `yak.cfg`.
The configuration file can be in one of the following locations:

1. `/etc/yak/yak.cfg`
2. `~/.config/yak/yak.cfg`
3. Current directory.
4. Set via the `YAK_CONFIG_FILE` environment variable or the `--config` command flag.

> The config file is named `yak.cfg and not `yak.yaml` in order to avoid being
> read when `yak` is run.

Authentication Information
--------------------------

Targets and Connections might require authentication information. For
example, discovering LXD containers requires connecting and authenticating
to a remote LXD server. Instead of specifying the authentication
information in the Yak file, you can place it in the configuration file.
This has the benefit of not storing authentication information in the same
place as your Yak files and then accidentally committing it to a repository.

Authentication configuration has the following format:

```yaml
auth:
  some-name:
    key: value
    key: value
```

As you can see, auth information is a simple hash/map of key/value pairs.

When a Target or Connection requires this information, it will have an
option called `auth`:

```yaml
targets:
  lxd:
    type: lxd_containers
    auth: some-name
```

### LXD Authentication

Use the following to target and connect LXD containers:

* `remote` (optional) - The name of the LXD remote server. Defaults
  to `local`.

* `address` (optional) - The address to the remote LXD server.

* `port` (optional) - The port of the LXD server. Defaults to 8443.

* `password` (optional) - The password to authenticate to the LXD remote.

* `scheme` (optional) - The scheme to use for the LXD remote. Valid values
  are `https` or `unix`. Defaults to `https`.

* `config_dir` (optional) - The LXC config directory.

* `accept_remote_certificate` - Whether to accept the LXD remote's certificate.
  Valid values are `true` and `false`. Defaults to `false`.

* `generate_client_certificate` - Whether to generate a client LXC certificate.
  Valid values are `true` and `false`. Defaults to `false`.

### OpenStack Authentication

Yak supports authenticating through a `clouds.yaml` file. You can specify the
name of the cloud as the `auth` option.

If you would prefer to add the authentication information to `yak.cfg`, the
following options are supported:

* `identity_endpoint` (Required) - The identity endpoint of the cloud.

* `username` (Required) - The username.

* `password` (Required) - The password.

* `tenant_id` (Optional) - The tenant/project ID.

* `tenant_name` (Optional) - The tenant/project name.

* `domain_id` (Optional) - The domain ID.

* `domain_name` (Optional) - The domain name.

* `project_domain_id` (Optional) - The project-scoped domain id.

* `project_domain_name` (Optional) - The project-scoped domain name.

* `user_domain_id` (Optional) - The user-scoped domain id.

* `user_domain_name` (Optional) - The user-scoped domain name.

* `region` - (Optional) - The OpenStack region to use.
