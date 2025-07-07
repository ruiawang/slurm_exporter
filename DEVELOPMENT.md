# Development

This project requires access to a node with the Slurm CLI (`sinfo`, `squeue`, `sdiag`, ...).

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.22 or higher recommended)
- Slurm CLI tools available in your `$PATH`

## Setup

Clone this repository:

```bash
git clone https://github.com/sckyzo/slurm_exporter.git
cd slurm_exporter
```

If you need a specific Go version, the Makefile will try to detect your installed version or use the default (`1.22.2`).

## Build

To build the exporter binary:

```bash
make build
```

The binary will be available in `bin/slurm_exporter`.

## Run tests

To run all tests:

```bash
make test
```

## Clean build artifacts

```bash
make clean
```

## Run the exporter

```bash
bin/slurm_exporter --web.listen-address=:8080
```

Or with GPU accounting enabled:

```bash
bin/slurm_exporter --web.listen-address=:8080 --gpus-acct
```

## Query metrics

```bash
curl http://localhost:8080/metrics
```

## Advanced

- You can override the Go version and architecture via environment variables:
  ```bash
  make build GO_VERSION=1.22.2 OS=linux ARCH=amd64
  ```

- The Makefile will automatically download and set up Go modules in a local `go/` directory if needed.

## References

* [Go client_golang documentation](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus)
* [Prometheus Metric Types](https://prometheus.io/docs/concepts/metric_types/)
* [Writing Exporters](https://prometheus.io/docs/instrumenting/writing_exporters/)
* [Available Exporters](https://prometheus.io/docs/instrumenting/exporters/)
