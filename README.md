# Cloud Foundry CPU Entitlement Plugin

Examine the CPU usage of Cloud Foundry applications, relative to their CPU entitlement.

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

## Why do I need this plugin?

The CPU entitlement plugin lets you see how much CPU your application is using
compared to its entitlement. Your CPU entitlement is calculated based on the
requested memory limit of your application: for example, a 128MB application is
entitled to use twice as much CPU as a 64MB application. The exact mapping from
memory to CPU is determined by the platform operator.

## Can I use this plugin on my Cloud Foundry?

???

### Will my application be throttled?

Current versions of Cloud Foundry allow applications to use more CPU than they
are entitled to, if CPU time is available at that moment, regardless of how the
application has behaved in the past. This behaviour will change in the future;
applications will still be allowed to temporarily go above their entitlement but
preference will be given to those applications that have been using less than
their entitlement over a rolling window of time. Applications will never be
forced below their entitlement.

## For operators

The current way Cloud Foundry allocates CPU makes it very hard to size cells.
For example, on a cell able to host 10 (equally sized) applications, Cloud
Foundry will assign a 10% of the CPU to each application. While these
applications may occasionally need 10% of the CPU (e.g. when starting up or when
under load), they don't most of the time. An operator may then over-commit the
cell, allowing for 20 equally sized applications to run on it, with a resulting
entitlement of 5%. This would still allow applications to spike to 10% usage,
provided some CPU is available. In presence of rogue applications trying to use
all CPU available at any time, this might not be possible and the well behaved
applications would be penalised.

What the operator really needs is the ability to allocate an average 5% CPU
usage to every application, allowing them to use 10% of the CPU when they need
to whilst still forcing to average at 5%. This is what the future behaviour will
be.

Operators can use this plugin to visualise which applications will be throttled
when the new behaviour is introduced. The metrics currently reported by Cloud
Foundry (e.g. by running `cf app`) can't provide this information, as they
represent absolute CPU usage and are dependant on the CPU usage of all the other
applications running on the same cell.

For example, an application entitled to 25% of the system CPU and using 25% of
it is using all the CPU it's been allocated and probably needs to be scaled up
to keep up with its workload. `cf app` will report a usage of 25%, while this
plugin will report a usage of 100%. On the other hand, an application entitled
to 50% of the system CPU and using 25% of it is idling 50% of the time: `cf app`
will still report a 25% usage, while this plugin will report a usage of 50%.
