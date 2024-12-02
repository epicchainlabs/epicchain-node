
<p align="center">
  <a href="https://fs.epicchain.org">epicchain</a> is a decentralized distributed object storage integrated with the <a href="https://epicchain.org">epicchain Blockchain</a>.
</p>

---

[![Report](https://goreportcard.com/badge/github.com/epicchainlabs/epicchain-node)](https://goreportcard.com/report/github.com/epicchainlabs/epicchain-node)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/nspcc-dev/epicchain-node?sort=semver)
![License](https://img.shields.io/github/license/nspcc-dev/epicchain-node.svg?style=popout)

# Overview

epicchain nodes are organized in a peer-to-peer network that stores and distributes users' data. Any epicchain user may participate in the network, either providing storage resources to earn compensation or storing their data at a competitive price.

Users can reliably store object data in the epicchain network and have full control over data placement due to the decentralized architecture and flexible storage policies. Nodes execute these storage policies, allowing users to specify parameters such as geographical location, reliability level, number of nodes, type of disks, capacity, and more.

Deep [epicchain Blockchain](https://epicchain.org) integration allows epicchain to be used by dApps directly from [epicchainVM](https://docs.epicchain.org/docs/en-us/basic/technology/epicchainvm.html) on the [Smart Contract](https://docs.epicchain.org/docs/en-us/intro/glossary.html) code level. This enables dApps to manipulate large amounts of data affordably.

epicchain offers a native [gRPC API](https://github.com/epicchainlabs/epicchain-api) and protocol gateways for popular protocols such as [AWS S3](https://github.com/epicchainlabs/epicchain-s3-gw), [HTTP](https://github.com/epicchainlabs/epicchain-http-gw), [FUSE](https://wikipedia.org/wiki/Filesystem_in_Userspace), and [sFTP](https://en.wikipedia.org/wiki/SSH_File_Transfer_Protocol), allowing developers to integrate applications without extensive code rewriting.

# Supported Platforms

Currently, we support GNU/Linux on amd64 CPUs with AVX/AVX2 instructions. More platforms will be supported after release `1.0`.

The latest version of epicchain-node is compatible with epicchain-contract [v0.19.1](https://github.com/epicchainlabs/epicchain-contract/releases/tag/v0.19.1).

# Building

To compile all binaries, you need Go 1.20+ and `make`:
```sh
make all
```
The resulting binaries will appear in the `bin/` folder.

To make a specific binary, use:
```sh
make bin/epicchain-<name>
```
See the list of all available commands in the `cmd` folder.

## Building with Docker

Building can also be performed in a container:
```sh
make docker/all                      # build all binaries
make docker/bin/epicchain-<name>     # build a specific binary
```

## Docker Images

To create Docker images suitable for use in [epicchain-dev-env](https://github.com/epicchainlabs/epicchain-dev-env/), use:
```sh
make images
```

# Running

## CLI

`epicchain-cli` allows users to manage containers and objects by connecting to any node in the target network. Detailed descriptions of all commands and options are provided within the tool, but some specific concepts have additional documentation:
* [Sessions](docs/cli-sessions.md)
* [Extended headers](docs/cli-xheaders.md)
* [Exit codes](docs/cli-exit-codes.md)

`epicchain-adm` is a network setup and management utility for network administrators. Refer to [docs/cli-adm.md](docs/cli-adm.md) for more information.

Both `epicchain-cli` and `epicchain-adm` can take a configuration file as a parameter to simplify working with the same network/wallet. See [cli.yaml](config/example/cli.yaml) for an example. Control service-specific configuration examples are [ir-control.yaml](config/example/ir-control.yaml) and [node-control.yaml](config/example/node-control.yaml) for IR and SN nodes, respectively.

## Node

There are two kinds of nodes: inner ring nodes and storage nodes. Most users will run storage nodes, while inner ring nodes are special and similar to epicchain consensus nodes in their network role. Both accept parameters from YAML or JSON configuration files and environment variables.

See [docs/sighup.md](docs/sighup.md) on how to reconfigure nodes without a restart.

See [docs/storage-node-configuration.md](docs/storage-node-configuration.md) on how to configure a storage node.

### Example Configurations

These examples contain all possible configurations of epicchain nodes. While the parameters are correct, the provided values are for informational purposes only and not recommended for direct use. Real networks and configurations will differ significantly.

See [node.yaml](node.yaml) for configuration notes.
- Storage node:
  - YAML (with comments): [node.yaml](config/example/node.yaml)
  - JSON: [node.json](config/example/node.json)
  - Environment: [node.env](config/example/node.env)
- Inner ring:
  - YAML: [ir.yaml](config/example/ir.yaml)
  - Environment: [ir.env](config/example/ir.env)

# Private Network

For epicchain development, consider using [epicchain-dev-env](https://github.com/epicchainlabs/epicchain-dev-env/). For developing applications using epicchain, the lighter [epicchain-aio](https://github.com/epicchainlabs/epicchain-aio) container is recommended. For manual deployment, refer to [docs/deploy.md](docs/deploy.md).

# Contributing

We welcome contributions! Please read the [contributing guidelines](CONTRIBUTING.md) before starting. Create a new issue first, describing the feature or topic you plan to implement.

# Credits

epicchain is maintained by [epicchainSPCC](https://nspcc.ru) with contributions from the community. Please see [CREDITS](CREDITS.md) for details.

# License

- [GNU General Public License v3.0](LICENSE)