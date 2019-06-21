package basic

import (
	"fmt"
	"strings"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&latestTagCheck{})
}

type latestTagCheck struct{}

// Name returns a unique name for this check.
func (l *latestTagCheck) Name() string {
	return "latest-tag"
}

// Groups returns a list of group names this check should be part of.
func (l *latestTagCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (l *latestTagCheck) Description() string {
	return "Checks if there are pods with container images having latest tag"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (l *latestTagCheck) Run(objects *kube.Objects) ([]kube.Diagnostic, error) {
	var diagnostics []kube.Diagnostic
	for _, pod := range objects.Pods.Items {
		podName := pod.GetName()
		namespace := pod.GetNamespace()
		diagnostics = append(diagnostics, checkTags(pod.Spec.Containers, podName, namespace)...)
		diagnostics = append(diagnostics, checkTags(pod.Spec.InitContainers, podName, namespace)...)
	}

	return diagnostics, nil
}

// checkTags checks if the image name conforms to pattern `image:latest` or `image`
// Adds a warning if it finds any image that uses the latest tag
func checkTags(containers []corev1.Container, podName string, namespace string) []kube.Diagnostic {
	var d []kube.Diagnostic
	for _, container := range containers {
		namedRef, _ := reference.ParseNormalizedNamed(container.Image)
		tagNameOnly := reference.TagNameOnly(namedRef)
		if strings.HasSuffix(tagNameOnly.String(), ":latest") {
			d = append(d, kube.Diagnostic{Category: "warning", Message: fmt.Sprintf("[Best Practice] Use specific tags instead of latest for container '%s' in pod '%s' in namespace '%s'", container.Name, podName, namespace)})
		}
	}
	return d
}
