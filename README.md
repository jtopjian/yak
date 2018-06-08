*DO NOT USE*

This project is unusable and very incomplete.

This was something I did for fun on my winter vacation. I have no
intention of continuing it.

Yak
===

Yak is a task execution and configuration management tool. Yes, another one.

It can execute commands locally or on remote hosts. Hosts can be
dynamically discovered or statically defined. Yak can connect to hosts
using standard SSH but also has the ability to use other methods such as
LXD.

Quickstart
----------

1. Download Yak.
2. Create a directory:

    ```bash
    $ mkdir demo
    $ cd demo
    ```

3. Add some hosts to a file:

    ```bash
    $ echo example1.com >> hosts.txt
    $ echo example2.com >> hosts.txt
    ```

4. Create a Yak file called `demo.yaml`:

    ```yaml
    targets:
      myhosts:
        type: textfile
        options:
          file: hosts.txt

    connections:
      ssh:
        type: ssh
        options:
          user: ubuntu
          private_key: /path/id/rsa
        targets:
          - myhosts

    task::install-memcached:
      steps:
        - name: install memcached
          action: apt.pkg
          input:
            name: memcached
            state: present
            sudo: true
    ```

5. Run the `install-memcached` task:

    ```bash
    $ yak run install-memcached
    ```

Documentation
-------------

See the [docs](/docs) directory.

Why??
-----

Mostly to scratch an itch.

A myriad of similar tools already exist. I've pulled my favorite features
from many of these tools (notably Terraform, Puppet, Ansible, and
StackStorm) and created something I don't believe quite exists at the
moment:

* Yak is written in Go, so it's an all-in-one binary.

* Yak is agent-less. If I'm managing 20 LXD containers, running an agent
  on each one feels wasteful.

* Yak can do more than just apply a standard configuration to a host. It
  can also run arbitrary tasks. You can have a dedicated task to bootstrap
  a Galera cluster and a dedicated task to manage the configuration state
  of the cluster once it's runing.

* Yak natively supports connection methods other than SSH. It doesn't just
  fork a process to call an external binary, either. For example, Yak can
  communicate with LXD servers directly through the LXD API. You don't need
  to have the LXD/LXC tools installed on your workstation to do this.

### Where is Yak best used?

I see Yak as being most useful *after* infrastructure has been deployed. Use
Terraform to build your physical and virtual infrastructure and then use Yak
to configure and maintain it.

### Why the name "Yak"?

Fiddling with Configuration Management is the epitome of Yak Shaving.

Building from Source
--------------------

```bash
$ go get -u github.com/jtopjian/yak/...
$ cd $GOPATH/src/github.com/jtopjian/yak/cmd
$ go build -o yak ./
```
