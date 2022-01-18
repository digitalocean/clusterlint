The snippet below is an example to show how to run clusterlint in-cluster with RBAC enabled.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: clusterlint-role
rules:
- apiGroups: [""]
  resources:
    - pods
    - volumes
    - deployments
    - services
    - cronjobs
    - namespaces
    - jobs
    - persistentvolumeclaims
    - persistentvolumes
    - statefulsets
    - storageclasses
    - configmaps
    - defaultstorageclass
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: clusterlint-role-binding
  namespace: clusterlint
subjects:
  - kind: ServiceAccount
    name: clusterlint
    namespace: clusterlint
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: clusterlint-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: clusterlint
  namespace: clusterlint
automountServiceAccountToken: false
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: clusterlint-cron
  namespace: clusterlint
spec:
  schedule: "0 */1 * * *"
  concurrencyPolicy: Replace
  failedJobsHistoryLimit: 3
  successfulJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: clusterlint
          containers:
            - name: clusterlint
              image: docker.io/clusterlint:latest
              imagePullPolicy: IfNotPresent
          restartPolicy: Never
```
