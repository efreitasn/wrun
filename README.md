# wrun
wrun is an inotify-based CLI that runs specified commands whenever the contents in the current directory change.

## Installing
You can install it by building the CLI yourself of download one of the release files available on the [releases page](https://github.com/efreitasn/wrun/releases). Instructions for both are below.

### Building (requires Go v1.13)

Once the repository is cloned, run

```shell
make
```

### From a release file
Extract the contents of the tarball or of the zip file to a directory. In that directory, run:

```shell
make install
```

## Watched events
The following events are watched: `IN_CREATE`, `IN_DELETE`, `IN_CLOSE_WRITE`, `IN_MOVED_FROM`, `IN_MOVED_TO`. To learn more about the inotify API, click [here](http://man7.org/linux/man-pages/man7/inotify.7.html).

## Using
To start watching, run `wrun start` in the directory to be watched. Note that this directory needs to have a config file.

### Config file
The easiest way to create a config file(`wrun.json`) is by running `wrun init`, which will create a config file in the current directory with all of the options set to their respective default values. In order to help editing the JSON properties, the generated config file also has a reference to a JSON schema. Nevertheless, a description for each one of the available fields is provided below.

> Some properties exist both globally and per command (e.g. `delayToKill` and `fatalIfErr`). The command version, if exists, always takes precedence over the global version.

#### `delayToKill`
The time in milliseconds to wait after sending a SIGINT and before sending a SIGKILL to a command. Defaults to 1000.

#### `fatalIfErr`
Whether to skip subsequent commands in case the current one returns an error. Defaults to false.

#### `ignoreRegExps`
List of regular expressions to ignore. The `wrun.json` file and the `.git` directory are always ignored. To learn more about the syntax of the regular expressions, click [here](https://github.com/google/re2/wiki/Syntax). Every directory path matched against these regular expressions ends with a `/`.

#### `cmds`
List of commands to be executed sequentially.

##### `cmd.fatalIfErr`
The same as the global version, except that it is command-wide.

##### `cmd.delayToKill`
The same as the global version, except that it is command-wide.

##### `cmd.terms`
The terms of the command, also known as arguments. The first term is always the command's name. For example, the terms for

```shell
touch a.txt
```
are
```shell
["touch", "a.txt"]
```

It's common to think of terms as the line that runs the command splitted by spaces, which is true, except if you want to provide a string with more than one word as an argument. For example

```shell
grep "some phrase here" file.txt
```

has a different list of terms than:

```shell
grep some phrase here file.txt
```

In the former, the list of terms is `["grep", "some phrase here", "file.txt"]`), and the `grep` command receives two arguments. In the latter, the list of terms is `["grep", "some", "phrase", "here", "file.txt"]`), and the `grep` command receives four arguments.