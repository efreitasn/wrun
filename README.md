# wrun
wrun is a tool to run commands whenever the files in a directory change.

**This is still an early project and the API is not stable, i.e., breaking changes might happen.**

## Installing
You can install it using go

```shell
go get -u github.com/efreitasn/wrun
```

or download one of the binaries available on the [releases page](https://github.com/efreitasn/wrun/releases).

## How to use
Add a wrun.json file with the following structure in the directory to be watched.

```json
{
  "PRECMD": [
    "sleep",
    "100"
  ],
  "CMD": [
    "echo",
    "foobar"
  ]
}
```

and run

```shell
wrun
```

## Ignored files
For now, it only ignores hidden files and directories (those start with a `.`).

## Thanks
* [watcher](https://github.com/radovskyb/watcher)