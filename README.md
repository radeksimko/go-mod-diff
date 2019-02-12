# go-mod-diff

To make comparison of dependencies easier.

## Supports

[Govendor](https://github.com/kardianos/govendor) as a legacy format.

## Usage

Run from the root of Go Module enabled repository (where `go.mod` is):
```
$ go-mod-diff /tmp/0.11-vendor.json
```

## Example output

![screen shot 2019-02-12 at 21 44 51](https://user-images.githubusercontent.com/287584/52670013-7bd3be00-2f0f-11e9-91cd-30bc609b6006.png)

## Dependencies

### `./go-src`

All packages in this folder are just copies of some internal packages from https://github.com/golang/go which cannot be imported directly as a result of being `internal`.
Use `go-update.sh` to update packages here.
