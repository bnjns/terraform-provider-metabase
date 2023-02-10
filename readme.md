<div align="center">

### Metabase Terraform Provider

![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/bnjns/terraform-provider-metabase/build.yml?branch=main&style=flat-square)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/bnjns/terraform-provider-metabase?display_name=tag&label=version&sort=semver&style=flat-square)
![GitHub issues](https://img.shields.io/github/issues/bnjns/terraform-provider-metabase?style=flat-square)

---

A Terraform provider that lets you manage your Metabase instance, because why not Terraform the world?
</div>

## 🧐 About

[Metabase](https://www.metabase.com/) is a simple analytics tool which allows anyone to easily learn and make decisions
from their company's data. It allows you to query data directly from your databases (called "questions"), which you can
store and share with others, as well as generate reports.

As with everything, software engineers are incredibly lazy and like to avoid configuring things manually so this
provider lets you manage your Metabase instance using Terraform and any standard infrastructure-as-code tooling and
processes you may already have.

⚠️ This provider is still a work-in-progress and not all features are available. See the v1 project for details.

## 🏁 Getting Started

### Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.17

For local development, you may also want one of:
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- Java Runtime Environment >= 1.8

### Installing

Simply clone this repo to your desired location:

```sh
$ git clone git@github.com:bnjns/terraform-provider-metabase.git
```

Install the Go dependencies:

```sh
$ go mod download
```

## 🎈 Usage

### Building the provider

To build the provider and install into your `GOPATH`:

```sh
$ go install
```

### Configuring Terraform

You can configure Terraform to use a [local build](#building-the-provider) by adding the following to you `~/.terraformrc` file:

```hcl
provider_installation {
  dev_overrides {
    "bnjns/metabase" = "</path/to/GOPATH>/bin"
  }

  direct {}
}
```

> **Note:** You must include `direct {}` otherwise all other providers will fail to install.

### Running Metabase

It is recommended that you run a local copy of Metabase to test against when developing, and inspect the API. You can
either use Docker to run Metabase or run it manually.

Once Metabase is started, you'll need to navigate to `http://localhost:3000` and configure Super User. Make sure you
remember the username (email) and password so that you can configure the provider when testing.

#### Using Docker

Simply use the included [`docker`](docker-compose.yml) set-up to run Metabase:

```sh
$ docker-compose up metabase
```

This will start Metabase on port 3000.

This also includes a selection of other databases, which you can use to experiment with. To start the entire network:

```sh
$ docker-compose up
```

#### Running Metabase manually

[Download the JAR](https://www.metabase.com/docs/latest/operations-guide/running-the-metabase-jar-file.html) to a
sensible location and run:

```sh
$ java \
    -DMB_PASSWORD_COMPLEXITY=weak \
    -DMB_SEND_EMAIL_ON_FIRST_LOGIN_FROM_NEW_DEVICE='false' \
    -jar metabase.jar
```

This will start Metabase on port 3000.

#### Setting up Metabase

When Metabase first starts, it will require some basic initial set up before it can be used. You can set it up manually
by visiting <https://localhost:3000> and following the steps or using the included script:

```sh
$ scripts/setup_metabase.sh
```

This will create a user with email `example@example.com` and password `password`.

### Generating the documentation

The documentation can be auto-generated using `tfplugindocs`:

```sh
go generate
```

### Running the tests

Make sure you have [Metabase running](#running-metabase) and configured using the included script, as both the client
and provider tests require an actual instance of Metabase to interact with.

To run the unit tests for the client:

```sh
$ go test -v ./...
```

To run the provider acceptance tests:

```sh
$ TF_ACC=1 go test -v ./... -run "^TestAcc"
```

> **Note:** While tests should randomly generate unique names in order to prevent conflicts, you may need to stop and
> restart Metabase between test runs (ensure you use `--force-recreate` if using Docker).

## 🚀 Releasing

Releasing is handled automatically by [GitHub Actions](.github/workflows/release.yml) and
Hashicorp's `terraform-provider-release` action. An admin will simply tag the latest release to trigger the pipeline.

## ⛏️ Built Using

- [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework)

## ✍️ Authors

- [@bnjns](https://github.com/bnjns)
