# CUE and Stacks 2.0

## Questions

### "Different" types of parameters

The whole point of "parameters" is to set environment variables that are used in the helmfile
templates of a stack.

To clarify what the different types of parameters mean and who "owns" them

* stack parameters are provided by us via stacks/dhis2/parameters/env.yaml?
  a user should not be able to provide these, right? yes
* system parameters are provided by us in helmfile.go
  a user should not be able to provide these, right? yes
* required and optional parameters are provided by users

A consumed parameter is a parameter that a stack needs in its template but that it will get from
another template. This means the stack consuming this parameter depends on the stack declaring it.

Hostname pattern and variables. How do these work again?

### Consumed Parameters

What do we currently do if I update a parameter that is consumed by another instance?
nothing

* an instance can currently only consume from one other instance? yes
* multiple instances can consume from the same instance? not from the same stack type. so pgadmin
  and dhis2-core can consume from the same dhis2-db but not 2 dhis2-core from the same dhis2-db.

## Conflicts

CUE unifies the config into one config or value. Our representation of parameters as a list of key
values leads to conflicts. Users can pass the same parameter twice in the list of parameters with
different values. CUE will error in such a case as the situation cannot lead to a single value for a
config.

This makes sense as such a situation can cause bugs if just ignored.

https://pkg.go.dev/os/exec#Cmd.Env
> // If Env contains duplicate environment keys, only the last
> // value in the slice for each duplicate key is used.

We should not allow such behavior from even taking place as this will be hard to debug.

```sh
cue export dhis2-partial.yaml stacks.cue
optionalParameters: incompatible list lengths (1 and 3)
optionalParameters.0.name: conflicting values "IMAGE_REPOSITORY" and "IMAGE_TAG":
    ./dhis2-partial.yaml:2:12
    ./stacks.cue:35:19
    ./stacks.cue:50:1
```

## Parameters as Map

An example of how merging data with schema works no matter where values come from.

Some of our parameters come from an encrypted yaml file.
"optional" and "required" parameters come from the user as JSON.
"system" parameters are set by us in our helmfile.go.

```sh
cue export dhis2-partial.yaml dhis2-partial.json stacks.cue
```

