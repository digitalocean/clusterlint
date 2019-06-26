###### Default Namespace

Name:         `default-namespace`

Group:        `basic`

Description:  Namespace is used to limit the scope of the Kubernetes resources created by multiple sets of users within a team. Even though there is a default namespace, dumping all the created resources into one namespace is not recommended. It can lead to privilege escalation, resource name collisions, latency in operations as resources scale up and mismanagement of kubernetes objects. Having namespaces ensures that resource quotas can be enabled to keep track node, cpu and memory usage for individual teams.

Example:

```yaml
# Don't do this
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

How to fix:

```yaml
# Explicitly specify namespace in the object config
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

###### Latest Tag

Name: `latest-tag`

Group: `basic`

Description: Using container images with `latest` tag or not specifying a tag in the image (implies `latest` tag) is not recommended. It leads to confusion around the version of image used. With a dynamic environment in a Kubernetes cluster, pods get rescheduled often. Upon a reschedule, you may find that the images' versions have changed. This can break the application and make it difficult to debug the errors in the application. You can update segments of the application individually if the images are pinned to specific versions.

Example:

```yaml
# Don't do this
spec:
  containers:
  - name: mypod
    image: nginx
  - name: redis
    image: redis:latest
```

How to fix:

```yaml
# Explicitly specify tag or digest in the object config
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
  - name: redis
    image: redis@sha256:dca057ffa2337682333a3aba69cc0e7809819b3cd7fc78f3741d9de8c2a4f08b
```

###### Privileged Containers

Name: `privileged-containers`

Group: `security`

Description: Use the `privileged` mode for trusted containers only. Because the privileged mode allows container processes to access the host, malicious containers can extensively damage the host and bring down many services on the cluster. Explicitly add additional capabilities for the container if that helps with getting more privileges than default. Sometimes, however, containers need to be run in privileged mode. Make sure to test the container before using in production. For more information about the risks of running containers in privileged mode, please refer to this [article](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/).

Example:

```yaml
# Don't do this
spec:
  containers:
  - name: mypod
    image: nginx
    securityContext:
      privileged: true
```

How to fix:

```yaml
# Explicitly add capabilities to the container if that helps with getting more privileges than default
spec:
  containers:
  - name: mypod
    image: nginx
    securityContext:
      capabilities:
        add:
        - NET_ADMIN
```

###### Fully Qualified Image

Name: `fully-qualified-image`

Group: `basic`

Description: Docker is the most popular runtime for Kubernetes. However, Kubernetes supports other container runtimes as well: containerd, CRI-O, etc. If the registry is not prepended to the image name, docker assumes `docker.io` and pulls it from DockerHub. However, the other runtimes will result in errors while pulling images. In order to maintain portability, we recommend to provide a fully qualified image name. If the underlying runtime is changed and the object configs are deployed to a new cluster, having fully qualified image names ensures that the applications don't break.

Example:

```yaml
# Don't do this
spec:
  containers:
  - name: mypod
    image: nginx:1.17.0
```

How to fix:

```yaml
# Provide the registry name in the image.
spec:
  containers:
  - name: mypod
    image: docker.io/nginx:1.17.0
```

###### Node name selector

Name: `node-name-pod-selector`

Group: `doks`

Description: On upgrade of a cluster on DOKS, the worker nodes' hostname changes. So, if a user's pod spec relies on the hostname to schedule pods on specific nodes, pod scheduling will fail after upgrade.

Example:

```yaml
# Don't do this
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

How to fix:

```yaml
# Use a custom label or a DOKS specific label
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

###### Pod State

Name: `pod-state`

Group: `workload-health`

Description: This check is done so users can find out if they have unhealthy pods in their cluster before upgrade. If there are suspicious failed pods, this check will indicate the same.

This check is not run by default. Specify group name or check name in order to run this check.


###### HostPath Volume

Name: `hostpath-volume`

Group: `basic`

Description: Using hostPath volumes is best avoided because:

- Pods with identical configuration (such as created from a podTemplate) may behave differently on different nodes due to different files on the nodes.
- When Kubernetes adds resource-aware scheduling, as is planned, it will not be able to account for resources used by a hostPath
the files or directories created on the underlying hosts are only writable by root.
- You either need to run your process as root in a privileged Container or modify the file permissions on the host to be able to write to a hostPath volume

For more details about hostpath, please refer to the Kubernetes [documentation](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)

Example:

```yaml
# Don't do this
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

How to fix:

```yaml
# Use other volume sources. See https://kubernetes.io/docs/concepts/storage/volumes/
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
###### Unused Persistent Volume

Name: `unused-pv`

Group: `basic`

Description: This check reports all the persistent volumes in the cluster that are not claimed by persistent volume claims in any namespace. The cluster can be cleaned up based on this information and there will be fewer objects to manage.

How to fix:

```bash
kubectl delete pv <unused pv>
```

###### Unused Persistent Volume Claims

Name: `unused-pvc`

Group: `basic`

Description: This check reports all the persistent volume claims in the cluster that are not referenced by pods in the respective namespaces. The cluster can be cleaned up based on this information and there will be fewer objects to manage.

How to fix:

```bash
<<<<<<< HEAD
kubectl delete pvc <unused pvc>
```

###### Unused Config Maps

Name: `unused-config-map`

Group: `basic`

Description: This check reports all the config maps in the cluster that are not referenced by pods in the respective namespaces. The cluster can be cleaned up based on this information and there will be fewer objects to manage.

How to fix:

```bash
kubectl delete configmap <unused config map>
```
