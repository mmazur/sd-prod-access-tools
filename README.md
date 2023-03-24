# sd-prod-access-tools
SD Production Access Tools Manager

## Usage

```
$ go run spat.go
SD Prod Access Tools Manager

Usage:
  spat [command]

Available Commands:
  check       Check local and upstream versions of all managed tools
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize spat and install prod access tools
  upgrade     Upgraded (chosen) managed tools to their latest versions

Flags:
  -h, --help   help for spat

Use "spat [command] --help" for more information about a command.
```

## Version Check

```
$ go run spat.go check
Latest versions:
v0.0.35,        5 assets,       service/backplane-cli (gitlab)
v0.1.66,        14 assets,      openshift-online/ocm-cli (github)
v0.14.2,        5 assets,       openshift/osdctl (github)
v1.2.15,        6 assets,       openshift/rosa (github)
v0.17.0,        14 assets,      coreos/butane (github)
v2.43.0,        35 assets,      prometheus/prometheus (github)
```

## First run (init)

```
$ go run spat.go init
Creating directory structure…           DONE (in `$HOME/.spat`)
Installing tools…                       FAIL (not implemented yet)
```
