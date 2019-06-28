package basic

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&unusedPVCheck{})
}

type unusedPVCheck struct{}

// Name returns a unique name for this check.
func (pv *unusedPVCheck) Name() string {
	return "unused-pv"
}

// Groups returns a list of group names this check should be part of.
func (pv *unusedPVCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (pv *unusedPVCheck) Description() string {
	return "Check if there are unused persistent volumes in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (pv *unusedPVCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pv := range objects.PersistentVolumes.Items {
		if pv.Spec.ClaimRef == nil {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Unused Persistent Volume '%s'.", pv.GetName()),
				Kind:     checks.PersistentVolume,
				Object:   &pv.ObjectMeta,
				Owners:   pv.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
