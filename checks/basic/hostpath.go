package basic

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&hostPathCheck{})
}

type hostPathCheck struct{}

// Name returns a unique name for this check.
func (h *hostPathCheck) Name() string {
	return "hostpath-volume"
}

// Groups returns a list of group names this check should be part of.
func (h *hostPathCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (h *hostPathCheck) Description() string {
	return "Check if there are pods using hostpath volumes"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (h *hostPathCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pod := range objects.Pods.Items {
		for _, volume := range pod.Spec.Volumes {
			pod := pod
			if volume.VolumeSource.HostPath != nil {
				d := checks.Diagnostic{
					Severity: checks.Error,
					Message:  "Avoid using hostpath for volume.",
					Kind:     checks.Pod,
					Object:   &pod.ObjectMeta,
					Owners:   pod.ObjectMeta.GetOwnerReferences(),
				}
				diagnostics = append(diagnostics, d)
				break
			}
		}
	}

	return diagnostics, nil
}
