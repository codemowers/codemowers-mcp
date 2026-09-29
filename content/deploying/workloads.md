# Kubernetes workload design

- Separate HTTP servers and background workers so their replicas and resource budgets scale independently. Reuse an image when they share code, selecting distinct startup commands. Workers without inbound network traffic need no Service. Process and shutdown rules belong to `building/general` and `developing/general`.
- Do not build an orchestrator inside an application container: no forked worker managers, supervisord/s6 process trees, embedded message brokers used to dispatch child workers, or custom restart/replication loops. Use Kubernetes workloads for process lifecycle and independently provisioned messaging services when durable queues are needed. An application may consume a queue and schedule bounded in-process tasks; it must not supervise a fleet of child services. The narrowly scoped legacy exception is described in `building/general`.
- Use Kubernetes Lease objects for leader election, CronJobs for schedules with minute-level resolution, and Jobs for finite work. Keep scheduling and supervision out of long-lived application containers.
- Package migrations and maintenance commands in the application image and run them as Jobs with the release's configuration. Coordinate execution so replicas cannot race migrations; support overlapping application versions during rollouts.
- Set termination grace periods to cover bounded request draining and job completion/return. Set job retry and concurrency limits for the operation's failure and duplicate-delivery semantics.
- Size CPU and memory per component, including heavier media workers. Inspect namespace quotas and defaults using `deploying/discovery`; supply values required by the workload under `deploying/general`.
- Discover messaging APIs before declaring resources. Choose partitions, retention, replay and deletion behavior deliberately. Compacted topics represent keyed state and do not preserve every historical event.

## Mount packaged data with image volumes

For the separate model/map/data images in `building/general`, prefer [Kubernetes image volumes](https://kubernetes.io/docs/concepts/storage/volumes/#image) when supported by the target cluster and container runtime. Mount the artifact's content into the application as read-only data; do not start a sidecar just to hold files. Keep mutable outputs on separate writable storage.

Expose the data-image reference and mount path through deployment configuration, pin production artifacts by digest, and verify registry access and the required image-volume support before use. Follow the official [image-volume walkthrough](https://kubernetes.io/docs/tasks/configure-pod-container/image-volumes/). If support is unavailable, agree on an available data-delivery mechanism while retaining separate, versioned artifacts; do not silently bundle the data back into the executable image.
