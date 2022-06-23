package doks

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&podSelectorCheck{})
}

type invalidSnapshotCheck struct{}

// Name returns a unique name for this check.
func (i *invalidSnapshotCheck) Name() string {
	return "invalid-volume-snapshot"
}

// Groups returns a list of group names this check should be part of.
func (i *invalidSnapshotCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (i *invalidSnapshotCheck) Description() string {
	return "Checks if there are invalid volume snapshots that would fail webhook validation"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (i *invalidSnapshotCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	return diagnostics, nil
}
