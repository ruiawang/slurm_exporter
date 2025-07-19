# Slurm Commands Used in slurm_exporter

This file documents all the Slurm shell commands executed by the `slurm_exporter` application to collect metrics.

## `collector/cpus.go`

- `sinfo -h -o %C`: Retrieves the state of CPUs (allocated/idle/other/total).

## `collector/gpus.go`

- `sinfo -a -h --Format=Nodes: ,GresUsed: --state=allocated`: Retrieves allocated GPUs.
- `sinfo -a -h --Format=Nodes: ,Gres: ,GresUsed: --state=idle,allocated`: Retrieves idle GPUs.
- `sinfo -a -h --Format=Nodes: ,Gres:`: Retrieves the total number of GPUs.
- `sacct -a -X --format=Allocgres --state=RUNNING --noheader --parsable2`: (Commented out in the code) Retrieves allocated GRES for running jobs.

## `collector/node.go`

- `sinfo -h -N -O NodeList,AllocMem,Memory,CPUsState,StateLong,Partition`: Retrieves detailed information for each node, including memory usage, CPU state, and partition.

## `collector/nodes.go`

- `sinfo -h -o %D|%T|%b -p <partition> | sort | uniq`: Retrieves the number of nodes by state and feature set for a given partition.
- `scontrol show nodes -o | grep -c NodeName=[a-z]*[0-9]*`: Counts the total number of nodes.
- `sinfo -h -o %R | sort | uniq`: Retrieves the list of all partitions.

## `collector/partitions.go`

- `sinfo -h -o%R,%C`: Retrieves the CPU state for each partition.
- `squeue -a -r -h -o%P --states=PENDING`: Retrieves the list of pending jobs by partition.

## `collector/queue.go`

- `/usr/bin/squeue -h -o %P,%T,%C,%r,%u`: Retrieves information about jobs in the queue (partition, state, cores, reason, user).

## `collector/scheduler.go`

- `/usr/bin/sdiag`: Retrieves statistics from the Slurm scheduler.

## `collector/slurm_binary_info.go`

- `sinfo --version`: Checks the version of `sinfo`.
- `squeue --version`: Checks the version of `squeue`.
- `sdiag --version`: Checks the version of `sdiag`.
- `scontrol --version`: Checks the version of `scontrol`.
- `sacct --version`: Checks the version of `sacct`.
- `sbatch --version`: Checks the version of `sbatch`.
- `salloc --version`: Checks the version of `salloc`.
- `srun --version`: Checks the version of `srun`.

## `collector/sshare.go`

- `sshare -n -P -o account,fairshare`: Retrieves fair share information by account.

## `collector/users.go`

- `squeue -a -r -h -o %A|%u|%T|%C`: Retrieves information about jobs by user (job ID, user, state, cores).