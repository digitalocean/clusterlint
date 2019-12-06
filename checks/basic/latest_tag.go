/*
Copyright 2019 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
func (l *latestTagCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pod := range objects.Pods.Items {
		diagnostics = append(diagnostics, l.checkTags(pod.Spec.Containers, pod)...)
		diagnostics = append(diagnostics, l.checkTags(pod.Spec.InitContainers, pod)...)
	}

	return diagnostics, nil
}

// checkTags checks if the image name conforms to pattern `image:latest` or `image`
// Adds a warning if it finds any image that uses the latest tag
func (l *latestTagCheck) checkTags(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	for _, container := range containers {
		namedRef, err := reference.ParseNormalizedNamed(container.Image)
		if err != nil {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Image name for container '%s' could not be parsed", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
			continue
		}
		tagNameOnly := reference.TagNameOnly(namedRef)
		if strings.HasSuffix(tagNameOnly.String(), ":latest") {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Avoid using latest tag for container '%s'", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}
