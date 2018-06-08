file-upload
-----------

`file-upload` will upload a file to a target

### example

```
action: file-upload
  input:
    source: /path/to/local/source.txt
    destination: /path/to/remote/destination.txt
    uid: 0
    gid: 0
    mode: 640
```

### options

* `source` (required) - The path to the source file.

* `destination` (required) - The path to the destinationa file.

* `uid` (optional) - The owner UID of the file. Defaults to `0`.

* `gid` (optional) - The owner GID of the file. Defaults to `0`.

* `mode` (optional) - The permissions of the file. Defaults to `640`.

* `timeout` (optional) - The amount of time before the upload times out.

