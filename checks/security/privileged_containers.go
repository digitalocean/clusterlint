package security

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&privilegedContainerCheck{})
}

type privilegedContainerCheck struct{}

// Name returns a unique name for this check.
func (pc *privilegedContainerCheck) Name() string {
	return "privileged-containers"
}

// Groups returns a list of group names this check should be part of.
func (pc *privilegedContainerCheck) Groups() []string {
	return []string{"security"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (pc *privilegedContainerCheck) Description() string {
	return "Checks if there are pods with containers in privileged mode"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (pc *privilegedContainerCheck) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var w []error

	for _, pod := range objects.Pods.Items {
		podName := pod.GetName()
		namespace := pod.GetNamespace()
		w = append(w, checkPrivileged(pod.Spec.Containers, podName, namespace)...)
		w = append(w, checkPrivileged(pod.Spec.InitContainers, podName, namespace)...)
	}

	return w, nil, nil
}

// checkPrivileged checks if the container is running in privileged mode
// Adds a warning if it finds any privileged container
func checkPrivileged(containers []corev1.Container, podName string, namespace string) []error {
	var w []error
	for _, container := range containers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			w = append(w, fmt.Errorf("[Best Practice] Privileged container '%s' found in pod '%s', namespace '%s'.", container.Name, podName, namespace))
		}
	}
	return w
}
