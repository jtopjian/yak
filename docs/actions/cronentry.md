cron.entry
----------

`cron.entry` will manage cron entry.

### example

```
task::some_cron_entry
  - name: add some cron
    action: cron.entry
    input:
      name: descriptive name
      state: present
      command: ls
      minute: */5
      hour: 2
```

### options

* `name` (required) - A descriptive name for the cron entry.

* `state` (optional) - The state of the entry. This can either be:
  a version number, `present` or `absent`. Defaults to `present`.

* `command` (required) - The command to perform.

* `minute` (optional) - The minute entry of the cron. Defaults to `*`.

* `hour` (optional) - The hour entry of the cron. Defaults to `*`.

* `day_of_month` (optional) - The day_of_month entry of the cron. Defaults to `*`.

* `month` (optional) - The month entry of the cron. Defaults to `*`.

* `day_of_week` (optional) - The day_of_week entry of the cron. Defaults to `*`.

* `sudo` - Whether or not sudo is required. Valid values are
  `true` or `false`.

* `timeout` - How long the command should run before it times out.
