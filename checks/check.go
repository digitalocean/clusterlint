package checks

import "github.com/digitalocean/clusterlint/kube"

// Check is a check that can run on Kubernetes objects.
type Check interface {
	// Name returns a unique name for this check.
	Name() string
	// Groups returns a list of group names this check should be part of. It is
	// valid to return nil or an empty list if a check does not belong in any
	// groups.
	Groups() []string
	// Description returns a detailed human-readable description of what this
	// check does.
	Description() string
	// Run runs this check on a set of Kubernetes objects. It can return
	// warnings (low-priority problems) and errors (high-priority problems) as
	// well as an error value indicating that the check failed to run.
	Run(*kube.Objects) (warnings []error, errors []error, err error)
}
