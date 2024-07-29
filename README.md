# Terraform Provider Wundergraph

This is the Terraform provider for [Wundergraph](https://wundergraph.com/).

See the [terraform registry documentation](https://registry.terraform.io/providers/labd/wundergraph/latest/docs) on how
to use it.

## Currently supported resources

Currently, the checked resources are supported. Support for additional resources will come when they are required in
projects, or contributed.

- [x] Namespace
- [x] Federated graph
- [x] Federated subgraph
- [ ] Linting rules
- [ ] OIDC Provider
- [ ] Persisted operations
- [ ] Monograph
- [ ] Router tokens
- [ ] Webhooks
- [ ] Organization

# Development

Most users will just want to use the provider through terraform. If you want to do local development see below for more
information.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (
see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
