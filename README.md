# go-mod-diff [![CircleCI](https://circleci.com/gh/radeksimko/go-mod-diff.svg?style=svg)](https://circleci.com/gh/radeksimko/go-mod-diff)

To make comparison of Go dependencies easier.

## Why?

`go mod init` is capable of migrating dependencies and its versions from
other dependency managers (such as govendor).
It does so on best effort basis.

Pinning to the exact same version is not always possible after migration:

 - some package managers version on package level, not modules, which results in multiple versions of a module
    - Go modules enforce 1 version per module (usually repository)
 - some package managers have different ways of pinning to tags or revisions
    - Go modules prefer semver-based pinning
 - some package managers don't track transitive dependencies
    - Go modules track all transitive dependencies

For these (and more) reasons this tool aims to help with comparison of dependency versions before and after the migration.

## Supports

 - [Govendor](https://github.com/kardianos/govendor) as input format.

## Usage

Run from the root of Go Module enabled repository (where `go.mod` is):
```
$ go-mod-diff /tmp/0.11-vendor.json
```

## Example output

![screen shot 2019-02-12 at 21 44 51](https://user-images.githubusercontent.com/287584/52670013-7bd3be00-2f0f-11e9-91cd-30bc609b6006.png)
