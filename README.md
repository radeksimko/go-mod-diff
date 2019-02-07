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

<img width="950" alt="screen shot 2019-02-05 at 23 27 46" src="https://user-images.githubusercontent.com/287584/52308480-b7193e80-299d-11e9-9403-de11c636075a.png">

## Dependencies

### `./go-src`

All packages in this folder are just copies of some internal packages from https://github.com/golang/go which cannot be imported directly as a result of being `internal`.

```sh
OLD_IMPORT_PATH="cmd/go/internal"
NEW_REL_PATH="go-src/cmd/go/_internal"

NEW_IMPORT_PATH="github.com/radeksimko/go-mod-diff/${NEW_REL_PATH}"
mkdir -p $NEW_REL_PATH
cp -r $GOPATH/src/github.com/golang/go/src/${OLD_IMPORT_PATH}/{modfile,module,semver} ${NEW_REL_PATH}/
find ./${NEW_REL_PATH} -name '*.go' | xargs -I{} sed -i -e "s:${OLD_IMPORT_PATH}:${NEW_IMPORT_PATH}:" {}
```
