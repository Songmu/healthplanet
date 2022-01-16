healthplanet
=======

[![Test Status](https://github.com/Songmu/healthplanet/workflows/test/badge.svg?branch=main)][actions]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/healthplanet)][PkgGoDev]

[actions]: https://github.com/Songmu/healthplanet/actions?workflow=test
[license]: https://github.com/Songmu/healthplanet/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/healthplanet

healthplanet is a cli and client library for healthplanet.jp.

## Synopsis

```console
% healthplanet metrics
% healthplanet request -status=innerscan
```

```go
var token = "deadbeef"
var cli *healthplanet.Client = healthplanet.NewClient(token)
ret, err := cli.Status(context.Background(), "innerscan", time.Now().AddDate(0, 0, -7), time.Now())
```

## Description

healthplanet is a cli and client library for healthplanet.jp.

## Installation

```console
# go install
% go install github.com/Songmu/healthplanet/cmd/healthplanet@latest

# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/healthplanet/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/healthplanet/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/healthplanet/main/install.sh | sh -s [vX.Y.Z]
```

## Author

[Songmu](https://github.com/Songmu)
