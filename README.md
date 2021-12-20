# Clusterlint

[![CircleCI](https://circleci.com/gh/digitalocean/clusterlint.svg?style=svg)](https://circleci.com/gh/digitalocean/clusterlint)

As clusters scale and become increasingly difficult to maintain, clusterlint helps operators conform to Kubernetes best practices around resources, security and reliability to avoid common problems while operating or upgrading the clusters.

Clusterlint queries live Kubernetes clusters for resources, executes common and platform specific checks against these resources and provides actionable feedback to cluster operators.  It is a non invasive tool that is run externally. Clusterlint does not alter the resource configurations.

### Background

Kubernetes resources can be configured and applied in many ways. This flexibility often makes it difficult to identify problems across the cluster at the time of configuration. Clusterlint looks at live clusters to analyze all its resources and report problems, if any.

There are some common best practices to follow while applying configurations to a cluster like:

- Namespace is used to limit the scope of the Kubernetes resources created by multiple sets of users within a team. Even though there is a default namespace, dumping all the created resources into one namespace is not recommended. It can lead to privilege escalation, resource name collisions, latency in operations as resources scale up and mismanagement of kubernetes objects. Having namespaces ensures that resource quotas can be enabled to keep track node, cpu and memory usage for individual teams.

- Always specify resource requests and limits on pods: When containers have resource requests specified, the scheduler can make better decisions about which nodes to place pods on. And when containers have their limits specified, contention for resources on a node can be handled in a specified manner.

While there are problems that are common to clusters irrespective of the environment they are running in, the fact that different Kubernetes configurations (VMs, managed solutions, etc.) have different subtleties affect how workloads run. Clusterlint provides platform specific checks to identify issues with resources that cluster operators can fix to run in a specific environment.

Some examples of such checks are:

- On upgrade of a cluster on [DOKS](https://www.digitalocean.com/products/kubernetes/), the worker nodes' hostname changes. So, if a user's pod spec relies on the hostname to schedule pods on specific nodes, pod scheduling will fail after upgrade.

*Please refer to [checks.md](https://github.com/digitalocean/clusterlint/blob/master/checks.md) to get some background on every check that clusterlint performs.*

### Install

```bash
go get github.com/digitalocean/clusterlint/cmd/clusterlint
```

The above command creates the `clusterlint` binary in `$GOPATH/bin`

### Usage

```bash
clusterlint list [options]  // list all checks available
clusterlint run [options]  // run all or specific checks
```

### Running in-cluster

If you're running clusterlint from within a Pod, you can use the `--in-cluster` flag to access the Kubernetes API from the Pod.

```
clusterlint --in-cluster run
```

Here's a simple example of CronJob definition to run clusterlint in the default namespace without RBAC : 

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: clusterlint-cron
spec:
  schedule: "0 */1 * * *"
  concurrencyPolicy: Replace
  failedJobsHistoryLimit: 3
  successfulJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: clusterlint
              image: docker.io/clusterlint:latest
              imagePullPolicy: IfNotPresent
          restartPolicy: Never
```

If you're using RBAC, see [docs/RBAC.md](docs/RBAC.md).

### Specific checks and groups

All checks that clusterlint performs are categorized into groups. A check can belong to multiple groups. This framework allows one to only run specific checks on a cluster. For instance, if a cluster is running on DOKS, then, running checks specific to AWS does not make sense. Clusterlint can blacklist aws related checks, if any while running against a DOKS cluster.

```bash
clusterlint run -g basic                // runs only checks that are part of the basic group
clusterlint run -G security            // runs all checks that are not part of the security group
clusterlint run -c default-namespace  // runs only the default-namespace check
clusterlint run -C default-namespace // exclude default-namespace check
```

### Disabling checks via Annotations

Clusterlint provides a way to ignore some special objects in the cluster from being checked. For example, resources in the kube-system namespace often use privileged containers. This can create a lot of noise in the output when a cluster operator is looking for feedback to improve the cluster configurations. In order to avoid such a situation where objects that are exempt from being checked, the annotation `clusterlint.digitalocean.com/disabled-checks` can be added in the resource configuration. The annotation takes in a comma separated list of check names that should be excluded while running clusterlint.

```json
"metadata": {
  "annotations": {
    "clusterlint.digitalocean.com/disabled-checks" : "noop,bare-pods"
  }
}
```

### Building local checks

Some individuals and organizations have Kubernetes best practices that are not
applicable to the general community, but which they would like to check with
clusterlint. If your check may be useful for *anyone* else, we encourage you to
submit it to clusterlint rather than keeping it local. However, if you have a
truly specific check that is not appropriate for sharing with the broader
community, you can implement it using Go plugins.

See the [example plugin](example-plugin) for documentation on how to build a
plugin. Please be sure to read the [caveats](example-plugin/README.md#caveats)
and consider whether you really want to maintain a plugin.

To use your plugin with clusterlint, pass its path on the commandline:

```console
$ clusterlint --plugins=/path/to/plugin.so list
$ clusterlint --plugins=/path/to/plugin.so run -c my-plugin-check
```

## Contributing

Contributions are welcome, in the form of either issues or pull requests. Please
see the [contribution guidelines](CONTRIBUTING.md) for details.

## License

Copyright 2019 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
