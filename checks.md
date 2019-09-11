[Clusterlint](https://github.com/digitalocean/clusterlint) flags issues with a cluster that might prevent or complicate its upgrade to a new Kubernetes version.

## Default Namespace

- Name: `default-namespace`
- Group: `basic`

Namespaces are a way to limit the scope of the resources that subsets of users within a team can create. While a default namespace is created for every Kubernetes cluster, we don't recommend adding all created resources into the default namespace because of the risk of privilege escalation, resource name collisions, latency in operations as resources scale up, and mismanagement of Kubernetes objects. Having namespaces lets you enable resource quotas can be enabled to track node, CPU and memory usage for individual teams.

### Example

```yaml
# Not recommended: Defining resources with no namespace, which adds them to the default.
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  labels:
    name: mypod
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
```

### How to Fix

```yaml
# Recommended: Explicitly specify a namespace in the object config
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  namespace: test
  labels:
    name: mypod
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
```

## Latest Tag

- Name: `latest-tag`
- Group: `basic`

We don't recommend using container images with the `latest` tag or not specifying a tag in the image (which defaults to `latest`), as this leads to confusion around the version of image used. Pods get rescheduled often as conditions inside a cluster change, and upon a reschedule, you may find that the images' versions have changed to use the latest release, which can break the application and make it difficult to debug errors. Instead, update segments of the application individually using images pinned to specific versions.

### Example

```yaml
# Not recommended: Not specifying an image tag, or using "latest"
spec:
  containers:
  - name: mypod
    image: nginx
  - name: redis
    image: redis:latest
```

### How to Fix

```yaml
# Recommended: Explicitly specify a tag or digest
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
  - name: redis
    image: redis@sha256:dca057ffa2337682333a3aba69cc0e7809819b3cd7fc78f3741d9de8c2a4f08b
```

## Privileged Containers

- Name: `privileged-containers`
- Group: `security`

Use the `privileged` mode for trusted containers only. Because the privileged mode allows container processes to access the host, malicious containers can extensively damage the host and bring down services on the cluster. If you need to run containers in privileged mode, test the container before using it in production. For more information about the risks of running containers in privileged mode, please refer to the [Kubernetes security context documentation](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/).

### Example

```yaml
# Not recommended: Using privileged mode instead of granting capabilities when it's not necessary
spec:
  containers:
  - name: mypod
    image: nginx
    securityContext:
      privileged: true
```

### How to Fix

```yaml
# Recommended: Explicitly add only the needed capabilities to the container
spec:
  containers:
  - name: mypod
    image: nginx
    securityContext:
      capabilities:
        add:
        - NET_ADMIN
```

## Run As Non-Root

- Name: `run-as-non-root`
- Group: `security`

If containers within a pod are allowed to run with the process ID (PID) `0`, then the host can be subjected to malicious activity. We recommend using a user identifier (UID) other than `0` in your container image for running applications. You can also enforce this in the Kubernetes pod configuration as shown below.

### Example

```yaml
# Not recommended: Doing nothing to prevent containers from running under UID 0
spec:
  containers:
  - name: mypod
    image: nginx
```

### How to Fix

```yaml
# Recommended: Ensure containers do not run as root
spec:
  securityContext:
    runAsNonRoot: true
  containers:
  - name: mypod
    image: nginx

```

## Fully Qualified Image

- Name: `fully-qualified-image`
- Group: `basic`

Docker is the most popular runtime for Kubernetes. However, Kubernetes supports other container runtimes as well, such as containerd and CRI-O. If the registry is not prepended to the image name, docker assumes `docker.io` and pulls it from DockerHub. However, the other runtimes will result in errors while pulling images. To maintain portability, we recommend using a fully qualified image name. If the underlying runtime is changed and the object configs are deployed to a new cluster, having fully qualified image names ensures that the applications don't break.

### Example

```yaml
# Not recommended: Failing to specify the registry in the image name
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
```

### How to Fix

```yaml
# Recommended: Provide the registry name in the image
spec:
  containers:
  - name: mypod
    image: docker.io/nginx:1.17.0
```

## Node Name Selector

- Name: `node-name-pod-selector`
- Group: `doks`

On upgrade of a cluster on DOKS, the worker nodes' hostname changes. So, if a user's pod spec relies on the hostname to schedule pods on specific nodes, pod scheduling will fail after the upgrade.

### Example

```yaml
# Not recommended: Using a raw DigitalOcean resource name in the nodeSelector
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
  nodeSelector:
    kubernetes.io/hostname: pool-y25ag12r1-xxxx
```

### How to Fix

```yaml
# Recommended: Use a custom label or a DOKS specific label
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
  nodeSelector:
    doks.digitalocean.com/node-pool: pool-y25ag12r1
```

## Admission Controller Webhook

- Name: `admission-controller-webhook`
- Group: `doks`

When you configure admission controllers with webhooks that have a `failurePolicy` set to `Fail`, it prevents managed components like Cilium, kube-proxy, and CoreDNS from starting on a new node during an upgrade. This will result in the cluster upgrade failing.

### Example

```yaml
# Not recommended: Configure a webhook with a failurePolicy set to "Fail"
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: sample-webhook.example.com
webhooks:
- name: sample-webhook.example.com
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
    scope: "Namespaced"
  clientConfig:
    service:
      namespace: kube-system
      name: webhook-server
      path: /pods
  admissionReviewVersions:
  - v1beta1
  timeoutSeconds: 1
  failurePolicy: Fail
```

### How to Fix

```yaml
# Recommended: Exclude objects in the `kube-system` namespace by explicitly specifying a namespaceSelector or objectSelector
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: sample-webhook.example.com
webhooks:
- name: sample-webhook.example.com
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
    scope: "Namespaced"
  clientConfig:
    service:
      namespace: kube-system
      name: webhook-server
      path: /pods
  admissionReviewVersions:
  - v1beta1
  timeoutSeconds: 1
  failurePolicy: Fail
  namespaceSelector:
    matchExpressions:
      - key: "doks.digitalocean.com/webhook"
        operator: "Exists"
```

## Pod State

- Name: `pod-state`
- Group: `workload-health`

This checks for unhealthy pods in a cluster. This check is not run by default. Specify a group name or a check name to run this check.

## HostPath Volume

- Name: `hostpath-volume`
- Group: `basic`

Using `hostPath` volumes is best avoided because:

- Pods with an identical configuration (such as those created from a `podTemplate`) intended to behave identically to one another regardless of their deployment will in fact behave differently from node to node due to differences in the files present on the nodes themselves.
- Resource-aware scheduling is not be able to account for resources used by a `hostPath` volume.
- The files created on the hosts are only writable by root; you will need to run your process as root in a privileged container or modify the file permissions on the host to be able to write to a `hostPath` volume.

For more details about `hostPath` volumes, please refer to [the Kubernetes documentation](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)

### Example

```yaml
# Not recommended: Using a hostPath volume
apiVersion: v1
kind: Pod
metadata:
  name: test-pd
spec:
  containers:
  - image: docker.io/nginx:1.17.0
    name: test-container
    volumeMounts:
    - mountPath: /test-pd
      name: test-volume
  volumes:
  - name: test-volume
    hostPath:
      path: /data
      type: Directory
```

### How to Fix

```yaml
# Recommended: Use other volume sources. See https://kubernetes.io/docs/concepts/storage/volumes/
apiVersion: v1
kind: Pod
metadata:
  name: test-pd
spec:
  containers:
  - image: docker.io/nginx:1.17.0
    name: test-container
    volumeMounts:
    - mountPath: /test-pd
      name: test-volume
  volumes:
  - name: test-volume
    cephfs:
      monitors:
        - 10.16.154.78:6789
      user: admin
      secretFile: "/etc/ceph/admin.secret"
      readOnly: true
```

## Unused Persistent Volume

- Name: `unused-pv`
- Group: `basic`

This check reports all the persistent volumes in the cluster that are not claimed by a `PersistentVolumeClaim` (PVC) in any namespace. You can clean up the cluster based on this information and there will be fewer objects to manage.

### How to Fix

```bash
kubectl delete pv <unused pv>
```

## Unused Persistent Volume Claims

- Name: `unused-pvc`
- Group: `basic`

This check reports all the PVCs in the cluster that are not referenced by pods in the respective namespaces. You can clean up the cluster based on this information.

### How to Fix

```bash
kubectl delete pvc <unused pvc>
```

## Unused Config Maps

- Name: `unused-config-map`
- Group: `basic`

This check reports all the config maps in the cluster that are not referenced by pods in the respective namespaces. You can clean up the cluster based on this information.

### How to Fix

```bash
kubectl delete configmap <unused config map>
```

## Unused Secrets

- Name: `unused-secret`
- Group: `basic`

This check reports all the secret names in the cluster that are not referenced by pods in the respective namespaces. You can clean up the cluster based on this information.

### How to Fix

```bash
kubectl delete secret <unused secret name>
```

## Resource Requests and Limits

- Name: `resource-requirements`
- Group: `basic`

When you specify resource limits for containers, the scheduler can make better decisions about which nodes to place pods on, and handle contention for resources on a node in a specified manner.

### Example

```yaml
# Not recommended: Scheduling pods without specifying any resource limits
apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - image: docker.io/nginx:1.17.0
    name: test-container
```

### How to Fix

```yaml
# Recommended: Specify resource requests and limits
apiVersion: v1
kind: Pod
metadata:
  name: test-pd
spec:
  containers:
  - image: docker.io/nginx:1.17.0
    name: test-container
    resources:
      limits:
        cpu: 102m
      requests:
        cpu: 102m
```

## Bare Pods

- Name: `bare-pods`
- Group: `basic`

When the node that a pod is running on reboots or fails, the pod is terminated and will not be restarted. However, a job will create new pods to replace terminated ones. For this reason, we recommend that you use a job, deployment, or `StatefulSet` rather than a bare pod, even if your application requires only a single pod.

### Example

```yaml
# Not recommended: Deploying a bare pod without any deployment parameters
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  namespace: test
  labels:
    name: mypod
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
```

### How to Fix

```yaml
# Recommended: Configure pods as part of a deployment, job, or StatefulSet
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: test
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
```
