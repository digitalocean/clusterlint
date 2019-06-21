package noop

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&check{})
}

type check struct{}

// Name returns a unique name for this check.
func (nc *check) Name() string {
	return "noop"
}

// Groups returns a list of group names this check should be part of.
func (nc *check) Groups() []string {
	return nil
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *check) Description() string {
	return "Does not check anything. Returns no errors."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *check) Run(*kube.Objects) ([]kube.Diagnostic, error) {
	return nil, nil
}
