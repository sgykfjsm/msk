# msk

**Manage TiDB Cloud clusters, Support daily operations, and Keep your workloads efficient.**

`msk` is a CLI tool designed to assist in the management and cost-effective operation of TiDB Cloud clusters. It provides functionalities to retrieve cluster metadata, generate usage notices, and send notifications to your preferred communication channels.

## Features

* [x] Extract metadata of all clusters from TiDB Cloud and store it in a designated database.
* [ ] Query running clusters from your database and generate a usage summary.
* [ ] Upload the usage summary to S3 for logging or notification purposes.
* [ ] Notify administrators via Slack or other channels using the generated summary.

## Installation

```bash
git clone https://github.com/sgykfjsm/msk.git
cd msk
go build -o msk
```

## Usage

```bash
NAME:
   msk - Manage TiDB Cloud clusters, Support daily operations, and Keep your workloads efficient

USAGE:
   msk [global options] [command [command options]]

VERSION:
   v0.1.0

COMMANDS:
   clusterinfo      Get information about TiDB Clusters and save it to a database
   generate-notice  Generate a notice from the information collected by clusterinfo and save it to S3
   notify           Notify via messaging service using the generated notice
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Requirements

* Go 1.22 or later
* AWS credentials configured (for S3 upload)
* Slack webhook URL or other notification credentials (optional)

## License

MIT License
