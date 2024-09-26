# ⚠️ This plugin is deprecated ⚠️
There is now a CPU entitlement metric in CF CLI. In CF CLI v8 there is a new column "CPU Entitlement". In CF CLI v9+ this "CPU Entitlemnet metric" will replace the previous "CPU" metric. You read more about it in the [github issue here](https://github.com/cloudfoundry/cli/issues/2812).


# Cloud Foundry CPU Entitlement Plugin (EXPERIMENTAL)

Examine the CPU usage of Cloud Foundry applications, relative to their CPU
entitlement.

## Installation

Get the binary `URL` for [your
platform](https://github.com/cloudfoundry/cpu-entitlement-plugin/releases)

```bash
$ cf install-plugin $URL
```

## Usage

Collect CPU metrics for existing Cloud Foundry applications by running:

```bash
$ cf cpu-entitlement $APP_NAME
```

_Note: Getting information about previous spikes requires cf-deployment version >v12.1.0!_

## Building

_Note: Dependencies for cpu-entitlement-plugin are managed using `go modules`. You do not need
to clone this repo to your GOPATH. Please ensure you have go version >=1.11 installed._

To install the latest version:

```bash
$ cd cpu-entitlement-plugin
$ make install

# for more options run: make help
```

## What is the cpu-entitlement-plugin?

The CPU entitlement plugin lets you see how much CPU your application is using
compared to its entitlement. Your CPU entitlement is calculated based on the
requested memory limit of your application: for example, a 128MB application is
entitled to use twice as much CPU as a 64MB application. The exact mapping from
memory to CPU is determined by the platform operator.

This plugin allows Cloud Foundry application developers to make better decisions
about scaling applications and initial resource allocation of applications.

### Will my application be throttled?

Current versions of Cloud Foundry allow applications to use more CPU than they
are entitled to if CPU time is available at that moment, regardless of how the
application has behaved in the past. This behaviour will change in the future;
applications will still be allowed to temporarily exceed their entitlement but
preference will be given to those applications that have been using less than
their entitlement over a rolling window of time. Applications will never be
forced below their entitlement.

## For operators

An application entitled to 25% of the system CPU and using 25% of
it is using all the CPU it's entitled to and probably needs to be scaled up to
keep up with its workload. `cf app` will report a usage of 25%, while this
plugin will report a usage of 100%. On the other hand, an application entitled
to 50% of the system CPU and using 25% of it is idling 50% of the time: `cf app`
will still report a 25% usage, while this plugin will report a usage of 50%.

Eventually we intend to use the entitlement usage metrics to automatically make
decisions about application CPU throttling. Operators can use this plugin to
visualise which applications will be throttled when the new behaviour is
introduced. The metrics currently reported by Cloud Foundry (e.g. by running `cf
app`) can't provide this information, as they represent absolute CPU usage and
are dependant on the CPU usage of all the other applications running on the same
cell. For example, an application averaging at 130% usage of its entitlement may
be throttled to allow other applications to temporarily exceed their
entitlement.

Operators may wish to over-commit or under-commit on the number of CPU shares
available to applications for entitlement. This configuration and the outcomes
are documented in the [garden BOSH
release](https://github.com/cloudfoundry/garden-runc-release/blob/develop/docs/cpu-entitlement.md).

### <a name="developer-workflow"></a> Developer Workflow

- Clone [CI repository](https://github.com/cloudfoundry/wg-app-platform-runtime-ci) (next to where this code is cloned), and make sure latest
is pulled by running `git pull`

  ```bash
  mkdir -p ~/workspace
  cd ~/workspace
  git clone https://github.com/cloudfoundry/wg-app-platform-runtime-ci.git
  ```
- [Git](https://git-scm.com/) - Distributed version control system
- [Go](https://golang.org/doc/install#install) - The Go programming
  language

##### Test With Docker

Running tests for this repo

- `./scripts/create-docker-container.bash`: This will create a docker container with appropriate mounts.
- `./scripts/test-in-docker-locally.bash`: Create docker container and run all tests and setup in a single script.
  - `./scripts/test-in-docker-locally.bash <package> <sub-package>`: For running tests under a specific package: e.g. `./scripts/test-in-docker-locally.bash reporter`

When inside docker container: 
- `/repo/scripts/docker/build-binaries.bash`: This will build binaries required for running tests e.g. cpu-entitlement-plugin
- `/repo/scripts/docker/test.bash`: This will run all unit-tests in this repo
- `CONFIG=<PATH-TO-CONFIG.json> /repo/scripts/docker/test.bash <e2e,integration>`: This will run all test that are dependent on a CF.

Here is an example for config.json:
```json
{
  "api": "https://api.<REPLACE_ME>",
  "admin_password": "<REPLACE_ME>",
  "admin_username": "admin",
  "ca_cert": "<REPLACE_ME> e.g. $(echo '' | openssl s_client -showcerts -servername api.${CF_SYSTEM_DOMAIN} -connect api.${CF_SYSTEM_DOMAIN}:443 -prexit 2>/dev/null | openssl x509 )"
}
```
