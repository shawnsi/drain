Drain
=====

TCP Connection Drainer

This tool uses iptables to stop new connections on given TCP ports.

Install
-------

```bash
$ go get github.com/shawnsi/drain
```

Usage
-----

```bash
$ sudo drain -h
```

How it Works
------------

This is the flow of `drain start` internals.

![drain start](https://raw.github.com/shawnsi/drain/develop/dot/flow.png)
