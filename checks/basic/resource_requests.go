/*
Copyright 2022 DigitalOcean

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
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&resourceRequirementsCheck{})
}

type resourceRequirementsCheck struct{}

// Name returns a unique name for this check.
func (r *resourceRequirementsCheck) Name() string {
	return "resource-requirements"
}

// Groups returns a list of group names this check should be part of.
func (r *resourceRequirementsCheck) Groups() []string {
	return []string{"basic", "doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (r *resourceRequirementsCheck) Description() string {
	return "Check if pods have resource requirements set"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (r *resourceRequirementsCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, pod := range objects.Pods.Items {
		d := r.checkResourceRequirements(pod.Spec.Containers, pod)
		diagnostics = append(diagnostics, d...)
		d = r.checkResourceRequirements(pod.Spec.InitContainers, pod)
		diagnostics = append(diagnostics, d...)
	}

	return diagnostics, nil
}

// checkImage checks if the image name is fully qualified
// Adds a warning if the container does not use a fully qualified image name
func (r *resourceRequirementsCheck) checkResourceRequirements(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	for _, container := range containers {
		if container.Resources.Size() == 0 {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Set resource requests and limits for container `%s` to prevent resource contention", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}
