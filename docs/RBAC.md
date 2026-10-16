The snippet below is an example to show how to run clusterlint in-cluster with RBAC enabled.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: clusterlint
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: clusterlint-role
rules:
- apiGroups: [""]
  resources:
   - configmaps
   - cronjobs
   - deployments
   - jobs
   - limitranges
   - namespaces
   - nodes
   - persistentvolumeclaims
   - persistentvolumes
   - pods
   - podtemplates
   - resourcequotas
   - secrets
   - serviceaccounts
   - services
   - statefulsets
   - volumes
  verbs: ["get", "watch", "list"]
- apiGroups: ["snapshot.storage.k8s.io"]
  resources:
  - volumesnapshotcontents
  - volumesnapshots
  verbs: ["get", "watch", "list"]
- apiGroups: ["batch"]
  resources:
  - cronjobs
  verbs: ["get", "watch", "list"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources:
  - validatingwebhookconfigurations
  - mutatingwebhookconfigurations
  - validatingadmissionpolicies
  - validatingadmissionpolicybindings
  verbs: ["get", "watch", "list"]
- apiGroups: ["storage.k8s.io"]
  resources:
  - storageclasses
  - defaultstorageclass
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: clusterlint-role-binding
subjects:
  - kind: ServiceAccount
    name: clusterlint
    namespace: clusterlint
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
          automountServiceAccountToken: true
          containers:
            - name: clusterlint
              image: digitalocean/clusterlint:latest
              command: ['/clusterlint', '--in-cluster', 'run']
              imagePullPolicy: IfNotPresent
          restartPolicy: Never
```
