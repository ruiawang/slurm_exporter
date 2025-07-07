[![Buy Me a Coffee](https://storage.ko-fi.com/cdn/kofi2.png?v=3)](https://ko-fi.com/C0C514I8WG)

# Prometheus Slurm Exporter

Prometheus collector and exporter for metrics extracted from the [Slurm](https://slurm.schedmd.com/overview.html) resource scheduling system.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Metrics](#metrics)
- [Prometheus Configuration](#prometheus-configuration)
- [Grafana Dashboard](#grafana-dashboard)
- [License](#license)

---

## Features

- Export CPU, GPU, node, job, and partition metrics from Slurm
- Optional GPU accounting (`--gpus-acct`)
- TLS and Basic Auth support via [Exporter Toolkit](https://github.com/prometheus/exporter-toolkit)
- Ready-to-use Grafana dashboard

---

## Installation

1. Build as described in [DEVELOPMENT.md](DEVELOPMENT.md) and copy `bin/slurm_exporter` to a node with Slurm CLI access.
2. (Optional) Use the Systemd unit: [lib/systemd/prometheus-slurm-exporter.service](lib/systemd/prometheus-slurm-exporter.service)  
   *Note: I have not personally tested this unit; it is a leftover from the original fork.*
3. (Optional) Snap package: see [packages/snap/README.md](packages/snap/README.md)  
   *Note: I have not personally tested the Snap packaging; it is a leftover from the original fork.*

---

## Usage

**Basic:**
```bash
./slurm_exporter --web.listen-address=:8080
```

**With GPU accounting:**
```bash
./slurm_exporter --web.listen-address=:8080 --gpus-acct
```

**With TLS/Basic Auth:**
```bash
./slurm_exporter --web.listen-address=:8080 --web.config.file=/path/to/web-config.yml
```

> **Note:** GPU accounting is **disabled by default**. Use `--gpus-acct` to enable.

**Sample `web-config.yml`:**
```yaml
tls_server_config:
  cert_file: /path/to/cert.crt
  key_file: /path/to/cert.key
basic_auth_users:
  admin: $2y$12$EXAMPLE_ENCRYPTED_PASSWORD_HASH
```
See [Exporter Toolkit documentation](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md) for details.

---

## Metrics

### CPU Metrics

| Metric      | Description                                      |
|-------------|--------------------------------------------------|
| Allocated   | CPUs allocated to a job                          |
| Idle        | CPUs available for use                           |
| Other       | CPUs unavailable for use                         |
| Total       | Total number of CPUs                             |

- Extracted from: [`sinfo`](https://slurm.schedmd.com/sinfo.html)
- [CPU Management Guide](https://slurm.schedmd.com/cpu_management.html)

---

### GPU Metrics

| Metric      | Description                                      |
|-------------|--------------------------------------------------|
| Allocated   | GPUs allocated to a job                          |
| Other       | GPUs unavailable for use                         |
| Total       | Total number of GPUs                             |
| Utilization | Total GPU utilization on the cluster             |

- Extracted from: [`sinfo`](https://slurm.schedmd.com/sinfo.html), [`sacct`](https://slurm.schedmd.com/sacct.html)
- [GRES scheduling](https://slurm.schedmd.com/gres.html)

> **IMPORTANT:** GPU accounting is **disabled by default**. Enable it with `--gpus-acct`.

---

### Node Metrics

| State       | Description                                      |
|-------------|--------------------------------------------------|
| allocated   | Allocated to one or more jobs                    |
| completing  | Jobs on node are completing                      |
| down        | Node unavailable                                 |
| drain       | Node drained/draining                            |
| fail        | Node expected to fail soon                       |
| error       | Node in error state                              |
| idle        | Not allocated to any jobs                        |
| maint       | Marked for maintenance                           |
| mixed       | Some CPUs allocated, others idle                 |
| planned     | Held for multi-node job launch                   |
| resv        | In advanced reservation                          |

- Extracted from: [`sinfo`](https://slurm.schedmd.com/sinfo.html)

#### Additional Node Usage Info

Since v0.18: CPUs and memory (allocated, idle, total) per node, with labels (hostname, status).

---

### Job Metrics

| State                | Description                                 |
|----------------------|---------------------------------------------|
| PENDING              | Awaiting resource allocation                |
| PENDING_DEPENDENCY   | Awaiting dependency resolution              |
| RUNNING              | Currently allocated resources               |
| SUSPENDED            | Execution suspended                         |
| CANCELLED            | Cancelled by user/admin                     |
| COMPLETING           | In process of completion                    |
| COMPLETED            | Terminated with exit code 0                 |
| CONFIGURING          | Waiting for resources to be ready           |
| FAILED               | Terminated with non-zero exit code          |
| TIMEOUT              | Terminated on time limit                    |
| PREEMPTED            | Terminated due to preemption                |
| NODE_FAIL            | Terminated due to node failure              |

- Extracted from: [`squeue`](https://slurm.schedmd.com/squeue.html)

---

### Partition Metrics

| Metric                | Description                                |
|-----------------------|--------------------------------------------|
| Running/Suspended Jobs| Per partition, by account and user         |
| CPUs                  | Total/allocated/idle per partition/user ID |

---

### Jobs per Account and User

- Running, pending, and suspended jobs per Slurm account and user (from [`squeue`](https://slurm.schedmd.com/squeue.html)).

---

### Scheduler Metrics

| Metric                 | Description                               |
|------------------------|-------------------------------------------|
| Server Thread count    | Active `slurmctld` threads                |
| Queue size             | Scheduler queue length                     |
| DBD Agent queue size   | SlurmDBD message queue length              |
| Last cycle             | Time for last scheduling cycle (µs)        |
| Mean cycle             | Mean scheduling cycle time                 |
| Cycles per minute      | Scheduling executions per minute           |
| Backfill metrics       | Backfilling jobs: cycle times, depth, etc. |

- Extracted from: [`sdiag`](https://slurm.schedmd.com/sdiag.html)

---

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'slurm_exporter'
    scrape_interval: 30s
    scrape_timeout: 30s
    static_configs:
      - targets: ['slurm_host.fqdn:8080']
```

- **scrape_interval**: 30s to avoid overloading Slurm master.
- **scrape_timeout**: Set to avoid `context_deadline_exceeded` errors.

Check config:
```bash
promtool check-config prometheus.yml
```

---

## Grafana Dashboard

A [Grafana dashboard](https://grafana.com/dashboards/4323) is available:

![Node Status](images/Node_Status.png)
![Job Status](images/Job_Status.png)
![Scheduler Info](images/Scheduler_Info.png)

---

## License

This project is licensed under the GNU General Public License, version 3 or later.

