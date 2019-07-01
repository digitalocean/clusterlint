package basic

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&unusedClaimCheck{})
}

type unusedClaimCheck struct{}

// Name returns a unique name for this check.
func (c *unusedClaimCheck) Name() string {
	return "unused-pvc"
}

// Groups returns a list of group names this check should be part of.
func (c *unusedClaimCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (c *unusedClaimCheck) Description() string {
	return "Check if there are unused persistent volume claims in the cluster"
}

type identifier struct {
	Name      string
	Namespace string
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (c *unusedClaimCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	used := make(map[identifier]bool)
	for _, pod := range objects.Pods.Items {
		for _, volume := range pod.Spec.Volumes {
			claim := volume.VolumeSource.PersistentVolumeClaim
			if claim != nil {
				used[identifier{Name: claim.ClaimName, Namespace: pod.GetNamespace()}] = true
			}
		}
	}

	for _, claim := range objects.PersistentVolumeClaims.Items {
		if _, ok := used[identifier{Name: claim.GetName(), Namespace: claim.GetNamespace()}]; !ok {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused persistent volume claim",
				Kind:     checks.PersistentVolumeClaim,
				Object:   &claim.ObjectMeta,
				Owners:   claim.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
