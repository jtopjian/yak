file-download
-------------

`file-download` will download a file from a target.

### example

```
action: file-download
  input:
    source: /path/to/remote/source.txt
    destination: /path/to/local/destination.txt
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

* `timeout` (optional) - The amount of time before the download times out

