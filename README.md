# envchanger

`envchanger` is a tiny CLI helper that lets you jump between environment directories inside an infra repository. When you are deep inside `my-infra/dev/...` you can resolve (and use) the matching `prod` path automatically.

## Installation

```bash
go build ./cmd/changeenv
```

This produces a binary named `changeenv` in the current directory.

## Usage

`changeenv` prints the target directory to standard output. Combine it with `cd` to move the shell to the resolved path. By default, the tool recognizes `dev`, `test`, and `prod` environment segments in your current path and swaps them with the requested environment:

```bash
# Switch from the current dev folder to the matching prod folder.
cd "$(changeenv prod)"
```

For convenience, create a shell function:

```bash
cenv() { cd "$(changeenv "$1")"; }
```

Run `changeenv --help` to see available commands and flags.

## Custom Environments

You can add custom environment names beyond the built-in `dev`, `test`, and `prod`. This is useful for regional deployments (`prod-us-east-1`, `test-eu-central-1`) or additional stages (`staging`, `canary`).

### Option 1: Config File

Create `~/.cenvrc` with one environment per line:

```bash
# ~/.cenvrc
staging
prod-us-east-1
test-eu-central-1
canary
```

Comments are supported (lines starting with `#` or inline after environment names).

### Option 2: Environment Variable

Set `CENV_ENVIRONMENTS` with space-separated values:

```bash
export CENV_ENVIRONMENTS="staging prod-us-east-1 test-eu-central-1"
```

Add this to your `.bashrc` or `.zshrc` to make it permanent.

### Using Both

Both methods can be used together. Environments are loaded in this order:
1. Built-in defaults: `dev`, `test`, `prod`
2. Custom environments from `~/.cenvrc` (if it exists)
3. Custom environments from `$CENV_ENVIRONMENTS` (if set)

Example usage with custom environments:

```bash
# Navigate to a regional prod environment
cd ~/infra/prod-us-east-1/services/app

# Switch to the matching test region
cenv test-eu-central-1
# Now in: ~/infra/test-eu-central-1/services/app
```

## Configure shell helper

Run `changeenv configure` to print the helper function and the suggested commands that append it to your shell configuration file. The command also checks whether the directory containing the `changeenv` binary is already on your `PATH` and prints the export you can add if needed.

Use the `--create` flag to append both snippets automatically:

```bash
changeenv configure --create
```

## Development

```bash
go build ./...
go test ./...
go run ./cmd/changeenv --help
```

The CLI logic lives in `cmd/changeenv/main.go` and uses Cobra for argument parsing. The path handling helpers are located in `internal/envpath`.
