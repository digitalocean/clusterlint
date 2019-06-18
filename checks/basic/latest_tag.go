package basic

import (
	"fmt"
	"strings"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
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
func (l *latestTagCheck) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var w []error
	for _, pod := range objects.Pods.Items {
		podName := pod.GetName()
		namespace := pod.GetNamespace()
		w = append(w, checkTags(pod.Spec.Containers, podName, namespace)...)
		w = append(w, checkTags(pod.Spec.InitContainers, podName, namespace)...)
	}

	return w, nil, nil
}

// checkTags checks if the image name conforms to pattern `image:latest` or `image`
// Adds a warning if it finds any image that uses the latest tag
func checkTags(containers []corev1.Container, podName string, namespace string) []error {
	var w []error
	for _, container := range containers {
		image := container.Image[strings.LastIndex(container.Image, "/")+1:]
		if strings.Contains(image, ":latest") || !strings.Contains(image, ":") {
			w = append(w, fmt.Errorf("[Best Practice] Use specific tags instead of latest for container '%s' in pod '%s' in namespace '%s'", container.Name, podName, namespace))
		}
	}
	return w
}
