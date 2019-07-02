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

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&fullyQualifiedImageCheck{})
}

type fullyQualifiedImageCheck struct{}

// Name returns a unique name for this check.
func (fq *fullyQualifiedImageCheck) Name() string {
	return "fully-qualified-image"
}

// Groups returns a list of group names this check should be part of.
func (fq *fullyQualifiedImageCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (fq *fullyQualifiedImageCheck) Description() string {
	return "Checks if containers have fully qualified image names"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (fq *fullyQualifiedImageCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, pod := range objects.Pods.Items {
		d := checkImage(pod.Spec.Containers, pod)
		diagnostics = append(diagnostics, d...)
		d = checkImage(pod.Spec.InitContainers, pod)
		diagnostics = append(diagnostics, d...)
	}

	return diagnostics, nil
}

// checkImage checks if the image name is fully qualified
// Adds a warning if the container does not use a fully qualified image name
func checkImage(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	for _, container := range containers {
		value, err := reference.ParseAnyReference(container.Image)
		if err != nil {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  fmt.Sprintf("Malformed image name for container '%s'", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		} else {
			if value.String() != container.Image {
				d := checks.Diagnostic{
					Severity: checks.Warning,
					Message:  fmt.Sprintf("Use fully qualified image for container '%s'", container.Name),
					Kind:     checks.Pod,
					Object:   &pod.ObjectMeta,
					Owners:   pod.ObjectMeta.GetOwnerReferences(),
				}
				diagnostics = append(diagnostics, d)
			}
		}
	}
	return diagnostics
}
