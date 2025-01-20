# Solana RPC Exporter

A Prometheus exporter that connects to a Solana RPC node and exposes various metrics about Solana cluster performance, health, and more.

> **Inspired by** [asymmetric-research/solana-exporter](https://github.com/asymmetric-research/solana-exporter). This project has been adapted specifically for **running against a Solana RPC node** instead of a validator node, and focuses on RPC-related metrics and cluster-wide information.

---

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
5. [Configuration](#configuration)
6. [Exposed Metrics](#exposed-metrics)
7. [Prometheus Configuration](#prometheus-configuration)
8. [Development](#development)
9. [License](#license)

---

## Overview

**Solana RPC Exporter** queries a Solana RPC endpoint for various statistics—such as slot height, block height, epoch info, transaction counts—and exposes them as Prometheus metrics. This allows you to easily monitor your Solana node(s) (or remote RPC endpoints) in your existing Prometheus + Grafana stack.

By focusing on RPC interactions rather than validator-specific internals, this exporter is well-suited for:

- **Lightweight setups** that only have an RPC node rather than a full validator.
- **Third-party services** that rely on remote RPC endpoints for data.
- **Multiple clusters** (e.g., mainnet-beta, testnet, devnet, localnet) where you want a unified exporter.

---

## Features

- **Slot and block metrics**: current slot, block height, epoch boundaries, etc.
- **RPC node health checks**: node commitment, network health, and more.
- **Transaction metrics**: total transaction counts, transaction pacing, etc.
- **Flexible configuration**: set custom RPC URLs, poll intervals, timeouts, etc.
- **Prometheus-compatible**: exposes metrics over an HTTP endpoint.
- **Extensible codebase**: you can add custom queries or instrumentation.

---

## Installation

### 1. Clone & Build Locally

If you have Go installed, you can clone this repo and build:

```bash
git clone https://github.com/naviat/solana-rpc-exporter.git
cd solana-rpc-exporter
go build -o solana-rpc-exporter cmd/solana-exporter/main.go
```

This will produce an executable named `solana-rpc-exporter`

### 2. Docker (Optional)

Build and run a Docker image:

```shell
docker build -t solana-rpc-exporter .
docker run -p 8080:8080 solana-rpc-exporter \
  --rpc-url http://your-solana-rpc:8899 \
  --network mainnet-beta
```

## Usage

Run the exporter with default settings (pointing to a local Solana RPC node):

```shell
./solana-rpc-exporter \
  --rpc-url http://localhost:8899 \
  --listen-address :8080 \
  --network mainnet-beta
```

The metrics will be available at:

```shell
http://localhost:8080/metrics
```

You can then configure Prometheus to scrape `localhost:8080/metrics`.

## Configuration

The exporter supports several CLI flags and environment variables. Below is a summary of the most common options:

| **Flag / Env Var**    | **Default**                | **Description**                                                            |
|-----------------------|----------------------------|----------------------------------------------------------------------------|
| `--rpc-url`           | `http://localhost:8899`    | URL of the Solana RPC endpoint.                                            |
| `--listen-address`    | `:8080`                    | Host and port for the exporter’s HTTP server.                              |
| `--network`           | `mainnet-beta`             | Network name (`mainnet-beta`, `testnet`, `devnet`, or `localnet`).         |
| `--http-timeout`      | `60` (seconds)             | Timeout (in seconds) for HTTP requests to the Solana RPC node.             |
| `--slot-pace`         | `1s`                       | Interval between successive slot polls.                                    |
| `--debug`             | `false`                    | Enable verbose logging (useful for debugging).                             |

> **Tip**: Use `--help` or consult the documentation for additional flags and corresponding environment variables (e.g., `SOLANA_URL`, `HTTP_TIMEOUT`, etc.).

---

## Exposed Metrics

Below is a selection of key metrics that the exporter publishes at `/metrics`:

| **Metric & Labels**                        | **Value**            | **Type** | **Help**                                                                             |
|--------------------------------------------|----------------------|----------|--------------------------------------------------------------------------------------|
| `solana_network_epoch`                     | `822`                | gauge    | **(Inferred)** Current epoch for the network (not explicitly labeled in the sample). |
| `solana_node_block_height`                 | `3.43303272e+08`     | gauge    | Current block height of the RPC node.                                                |
| `solana_node_first_available_block`        | `3.47071418e+08`     | gauge    | First available block in the RPC node's ledger.                                      |
| `solana_node_health`                       | `1`                  | gauge    | Health status of the RPC node (1 = healthy, 0 = unhealthy).                          |
| `solana_node_minimum_ledger_slot`          | `3.47071417e+08`     | gauge    | Lowest slot that the RPC node has information about in its ledger.                   |
| `solana_node_num_slots_behind`             | `0`                  | gauge    | Number of slots the RPC node is behind.                                              |
| `solana_node_slot_height`                  | `3.55233915e+08`     | gauge    | Current slot height of the RPC node.                                                 |
| `solana_node_transaction_count`            | `1.499279778e+10`    | gauge    | Total number of transactions processed by the RPC node.                              |
| `solana_node_version_info`                 | `1`                  | gauge    | Version information of the RPC node.                                                 |

These metrics can be scraped by Prometheus and then visualized in your preferred dashboarding tool (e.g., Grafana).

## Prometheus Configuration

In your prometheus.yml, add:

```yaml
scrape_configs:
  - job_name: 'solana-rpc-exporter'
    static_configs:
      - targets: ['localhost:8080']
```

Adjust if you’re running on a different host or port.

## Development

### 1. Clone the repository

```shell
git clone https://github.com/naviat/solana-rpc-exporter.git
cd solana-rpc-exporter
```

### 2. Install dependencies

```shell
go mod tidy
```

### 3. Run tests

```shell
go test -v ./cmd/solana-exporter
go test -v ./pkg/rpc
```

### 4. Build

```shell
go build -o solana-rpc-exporter cmd/solana-exporter/main.go
```

### 5. Run locally

```shell
./solana-rpc-exporter --rpc-url=http://localhost:8899 --listen-address=:8080 --network=mainnet-beta
```

Feel free to open pull requests or issues to add more metrics, fix bugs, or improve performance.

## License

This project is licensed under the [MIT License](./LICENSE).

Happy exporting! If you have any questions or suggestions, please open an [issue](https://github.com/naviat/solana-rpc-exporter/issues). We welcome feedback and contributions.
