package doks

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&podSelectorCheck{})
}

type podSelectorCheck struct{}

// Name returns a unique name for this check.
func (nc *podSelectorCheck) Name() string {
	return "node-name-pod-selector"
}

// Groups returns a list of group names this check should be part of.
func (nc *podSelectorCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *podSelectorCheck) Description() string {
	return "Checks if there are pods which use kubernetes.io/hostname label in the node selector."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *podSelectorCheck) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var e []error
	for _, pod := range objects.Pods.Items {
		nodeSelectorMap := pod.Spec.NodeSelector
		if _, ok := nodeSelectorMap[corev1.LabelHostname]; ok {
			e = append(e, fmt.Errorf("pod '%s' in namespace '%s' uses the node name for node selector", pod.GetName(), pod.GetNamespace()))
		}
	}
	return nil, e, nil
}
