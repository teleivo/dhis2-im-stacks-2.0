# Stacks 2.0

This repo is just to play with different ideas contributing to
https://github.com/dhis2-sre/im-manager Stacks 2.0.

Directory

* cue - is a playground investigating https://cuelang.org
* draft - is a playground investigating a path without using cue

## Draft

Browse the code in `./draft`. Run some dummy examples and analysis on our stacks using

```sh
go run main.go
```

The types for stacks and parameters are in in `./draft/stack/stack.go`.
`./draft/main.go` shows you some dummy scenarios or uses.

A simple diagram of our stacks is created using https://d2lang.com/tour/install.

![stacks](./draft/stacks.svg?raw=true "Stacks")

You can adapt the code, watch it change when rerunning the code creating the diagram using

```sh
d2 --watch stacks.d2
```

