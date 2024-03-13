# March Madness 2024

This repository contains code for running the March Madness 2024 event.

## Deploying

First, clone the repository. You'll need to fetch the required files and
generate the needed code:

```sh
# Fetch needed submodules
git submodule update --init --recursive
# Generate the needed code
go generate ./...
```

Then, you can drop into a Nix shell and start the server:

```sh
nix develop
go run .
```

Edit the `config.json` as needed.
