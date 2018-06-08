file-delete
-----------

`file-delete` will delete a file on a target host.

### example

```
action: file-delete
  input:
    path: /path/to/remote/file.txt
```

### options

* `path` (required) - The path to the file to delete.

* `timeout` (optional) - The amount of time before the deletion times out.
