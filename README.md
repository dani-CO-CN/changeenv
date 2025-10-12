# envchanger

`envchanger` is a tiny CLI helper that lets you jump between environment directories inside an infra repository. When you are deep inside `my-infra/dev/...` you can resolve (and use) the matching `prod` path automatically.

## Installation

```bash
go build ./cmd/changeenv
```

This produces a binary named `changeenv` in the current directory.

## Usage

`changeenv` prints the target directory to standard output. Combine it with `cd` to move the shell to the resolved path. The tool looks for the `dev`, `test`, or `prod` segment in your current path and swaps it with the requested environment:

```bash
# Switch from the current dev folder to the matching prod folder.
cd "$(changeenv prod)"
```

For convenience, create a shell function:

```bash
cenv() { cd "$(changeenv "$1")"; }
```

Run `changeenv --help` to see available commands and flags.

### Flags

- `--create` — when used with `configure`, append the shell helper automatically.

Example:

```bash
changeenv prod
```

### Configure shell helper

Run `changeenv configure` to print the helper function and the suggested commands that append it to your shell configuration file. The command also checks whether the directory containing the `changeenv` binary is already on your `PATH` and prints the export you can add if needed. Use `changeenv configure --create` to append both snippets automatically.

## Development

```bash
go build ./...
go test ./...
go run ./cmd/changeenv --help
```

The CLI logic lives in `cmd/changeenv/main.go` and uses Cobra for argument parsing. The path handling helpers are located in `internal/envpath`.
