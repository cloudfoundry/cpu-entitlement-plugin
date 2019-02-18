# Cloud Foundry CPU Entitlement Plugin

Examine CPU performance of Cloud Foundry applications.

## Installation

Get the binary `URL` for [your platform](https://github.com/cloudfoundry/cpu-entitlement-plugin/releases)

```bash
$ cf install-plugin $URL
```

## Usage

Collect CPU metrics for existing Cloud Foundry applications by running:

```bash
$ cf cpu-entitlement $APP_NAME
```

## Building

To install the latest version:

```bash
$ cd cpu-entitlement-plugin
$ make
$ make install
```
