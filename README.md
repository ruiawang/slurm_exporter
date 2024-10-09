<script type='text/javascript' src='https://storage.ko-fi.com/cdn/widget/Widget_2.js'></script>
<script type='text/javascript'>
  kofiwidget2.init('Buy me a coffee ❤️☕', '#29abe0', 'C0C514I8WG');
  kofiwidget2.draw();
</script>

# Prometheus Slurm Exporter

Prometheus collector and exporter for metrics extracted from the [Slurm](https://slurm.schedmd.com/overview.html) resource scheduling system.

## Exported Metrics

### State of the CPUs

* **Allocated**: CPUs which have been allocated to a job.
* **Idle**: CPUs not allocated to a job and thus available for use.
* **Other**: CPUs which are unavailable for use at the moment.
* **Total**: total number of CPUs.

- Information extracted from the SLURM [**sinfo**](https://slurm.schedmd.com/sinfo.html) command.
- [Slurm CPU Management User and Administrator Guide](https://slurm.schedmd.com/cpu_management.html)

### State of the GPUs

* **Allocated**: GPUs which have been allocated to a job.
* **Other**: GPUs which are unavailable for use at the moment.
* **Total**: total number of GPUs.
* **Utilization**: total GPU utilization on the cluster.

- Information extracted from the SLURM [**sinfo**](https://slurm.schedmd.com/sinfo.html) and [**sacct**](https://slurm.schedmd.com/sacct.html) commands.
- [Slurm GRES scheduling](https://slurm.schedmd.com/gres.html)

**IMPORTANT**: GPU accounting is **disabled by default**. To enable GPU accounting, you must pass the `--gpus-acct` flag when running the exporter. Without this flag, GPU-related metrics will not be collected or exported.

**NOTE**: Since version **0.19**, you must explicitly enable GPU accounting by adding the `--gpus-acct` option.


### State of the Nodes

* **Allocated**: nodes which have been allocated to one or more jobs.
* **Completing**: all jobs associated with these nodes are in the process of being completed.
* **Down**: nodes which are unavailable for use.
* **Drain**: nodes in a ``drained`` or ``draining`` state.
* **Fail**: nodes expected to fail soon and unavailable for use.
* **Error**: nodes in an error state and incapable of running jobs.
* **Idle**: nodes not allocated to any jobs.
* **Maint**: nodes marked for maintenance.
* **Mixed**: nodes with some CPUs allocated and others idle.
* **Planned**: nodes held for a multi-node job launch.
* **Resv**: nodes in an advanced reservation.

- Information extracted from the SLURM [**sinfo**](https://slurm.schedmd.com/sinfo.html) command.

#### Additional node usage info

Since version **0.18**, information about CPUs and memory (allocated, idle, and total) is also extracted for every node known by Slurm. This includes node labels like hostname and status.

### Status of the Jobs

* **PENDING**: Jobs awaiting resource allocation.
* **PENDING_DEPENDENCY**: Jobs awaiting job dependency resolution.
* **RUNNING**: Jobs currently allocated resources.
* **SUSPENDED**: Jobs with suspended execution.
* **CANCELLED**: Jobs cancelled by a user or administrator.
* **COMPLETING**: Jobs in the process of completion.
* **COMPLETED**: Jobs that terminated with an exit code of zero.
* **CONFIGURING**: Jobs waiting for resources to become ready.
* **FAILED**: Jobs that terminated with a non-zero exit code.
* **TIMEOUT**: Jobs terminated upon reaching their time limit.
* **PREEMPTED**: Jobs terminated due to preemption.
* **NODE_FAIL**: Jobs terminated due to node failure.

- Information extracted from the SLURM [**squeue**](https://slurm.schedmd.com/squeue.html) command.

### State of the Partitions

* Running/suspended jobs per partition, divided by Slurm account and user.
* Total/allocated/idle CPUs per partition and per user ID.

### Jobs Information per Account and User

Information about running, pending, and suspended jobs per Slurm account and user are also extracted using [**squeue**](https://slurm.schedmd.com/squeue.html).

### Scheduler Information

* **Server Thread count**: Number of active `slurmctld` threads.
* **Queue size**: Length of the scheduler queue.
* **DBD Agent queue size**: Length of the SlurmDBD message queue.
* **Last cycle**: Time for the last scheduling cycle (microseconds).
* **Mean cycle**: Mean scheduling cycle time since last reset.
* **Cycles per minute**: Number of scheduling executions per minute.
* **Backfill metrics**: Metrics related to backfilling jobs, including cycle times, depth mean, and total backfilled jobs.

- Information extracted from the SLURM [**sdiag**](https://slurm.schedmd.com/sdiag.html) command.

### TLS and Basic Authentication

The Prometheus Slurm Exporter supports TLS and Basic Authentication by using a configuration file. To enable these features, you need to specify the path to a configuration file via the `--web.config.file` flag. For more information on how to configure TLS or Basic Auth, refer to the [Exporter Toolkit documentation](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

Example:

```bash
./slurm_exporter --web.config.file=/path/to/web-config.yml
```

An example `web-config.yml` file:

```yaml
tls_server_config:
  cert_file: /path/to/cert.crt
  key_file: /path/to/cert.key
basic_auth_users:
  admin: $2y$12$EXAMPLE_ENCRYPTED_PASSWORD_HASH
```

For more details, see the [Exporter Toolkit documentation](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

## Installation

1. Build the exporter as described in [DEVELOPMENT.md](DEVELOPMENT.md) and copy the executable `bin/slurm_exporter` to a node with access to the Slurm CLI.
2. A Systemd unit file is provided in [lib/systemd/prometheus-slurm-exporter.service](lib/systemd/prometheus-slurm-exporter.service).
3. Optionally, you can package the exporter as a Snap. See [packages/snap/README.md](packages/snap/README.md) for details.

## Commands to Start the Exporter

Here are the different ways to start the exporter based on your needs:

1. **Basic launch without GPU accounting:**

```bash
./slurm_exporter --web.listen-address=:8080
```

2. Launch with GPU accounting enabled:

```bash
./slurm_exporter --web.listen-address=:8080 --gpus-acct
```

Launch with TLS and Basic Authentication:

```bash
./slurm_exporter --web.listen-address=:8080 --web.config.file=/path/to/web-config.yml
```

For more details on TLS and Basic Authentication configuration, refer to the [Exporter Toolkit documentation](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

## Prometheus Configuration

Configure Prometheus to scrape the Slurm exporter:

```yaml
scrape_configs:
  - job_name: 'my_slurm_exporter'
    scrape_interval: 30s
    scrape_timeout: 30s
    static_configs:
      - targets: ['slurm_host.fqdn:8080']
```

* **scrape_interval**: Set to 30 seconds to prevent overloading the Slurm master.
* **scrape_timeout**: Ensure a reasonable timeout to avoid `context_deadline_exceeded` errors on a busy Slurm master.

Check the Prometheus configuration before reloading:

```bash
$ promtool check-config prometheus.yml
```

## Grafana Dashboard

A [Grafana dashboard](https://grafana.com/dashboards/4323) is available to visualize the exported metrics:

![Node Status](images/Node_Status.png)

![Job Status](images/Job_Status.png)

![Scheduler Info](images/Scheduler_Info.png)

## License

This project is licensed under the GNU General Public License, version 3 or later.

