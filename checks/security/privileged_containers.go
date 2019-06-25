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
func (pc *privilegedContainerCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, pod := range objects.Pods.Items {
		diagnostics = append(diagnostics, checkPrivileged(pod.Spec.Containers, pod)...)
		diagnostics = append(diagnostics, checkPrivileged(pod.Spec.InitContainers, pod)...)
	}

	return diagnostics, nil
}

// checkPrivileged checks if the container is running in privileged mode
// Adds a warning if it finds any privileged container
func checkPrivileged(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	for _, container := range containers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("[Best Practice] Privileged container '%s' found in pod '%s'", container.Name, pod.GetName()),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}
