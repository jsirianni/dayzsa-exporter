[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/jsirianni/dayzsa-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/jsirianni/dayzsa-exporter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsirianni/dayzsa-exporter)](https://goreportcard.com/report/github.com/jsirianni/dayzsa-exporter)

# dayzsa-exporter

Dayzsa Exporter is a Prometheus exporter for DayZ Standalone servers. It depends
on [dayzsalauncher.com](https://dayzsalauncher.com/#/servercheck).

Metrics are exposed over HTTP at `http://localhost:9100`.

## Installation

### Linux (Systemd)

The exporter can be installed on most Linux distributions. You can download
the latest release from the [releases page](https://github.com/jsirianni/dayzsa-exporter/releases).

RPM packages can be installed using the following commands:

```bash
sudo dnf install dayzsa-exporter-amd64.rpm
```

Debian packages can be installed using the following commands:

```bash
sudo apt-get install -f ./dayzsa-exporter-amd64.deb
```

Once installed, the exporter can be started using the following command:

```bash
sudo systemctl enable dayzsa-exporter
sudo systemctl start dayzsa-exporter
```

You can view the logs with Journalctl:

```bash
sudo journalctl -u dayzsa-exporter -f
```

### Docker

The exporter can be run as a Docker container.

```bash
docker run -d -p 9100:9100 ghcr.io/jsirianni/dayzsa-exporter:latest
```

## Configuration

The exporter is configured using a configuration file located at
`/etc/dayzsa/config.yaml`. The configuration file is in YAML format.

```yaml
# Collection interval duration
# e.g. 60s, 1m, 5m, 1h
interval: 60s

# One or more servers to monitor
servers:
  - ip: "50.108.13.235"
    port: 2424
  - ip: "50.108.13.235"
    port: 2324
  - ip: "50.108.13.235"
    port: 2315
  - ip: "50.108.13.235"
    port: 27016
```

## Usage

Once installed, you can test with cURL

```bash
curl -s localhost:9100 | grep -v '#' | grep dayzsa_exporter
```

### Monitor with Bindplane

You can use [Bindplane](https://bindplane.com/solutions) to monitor
your exporter. The [Prometheus Source](https://bindplane.com/docs/resources/sources/prometheus)
can be used to collect metrics from the exporter.
