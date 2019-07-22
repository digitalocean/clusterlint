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

package security

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&nonRootUserCheck{})
}

type nonRootUserCheck struct{}

// Name returns a unique name for this check.
func (nr *nonRootUserCheck) Name() string {
	return "non-root-user"
}

// Groups returns a list of group names this check should be part of.
func (nr *nonRootUserCheck) Groups() []string {
	return []string{"security"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nr *nonRootUserCheck) Description() string {
	return "Checks if there are pods which run as root user"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nr *nonRootUserCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, pod := range objects.Pods.Items {
		diagnostics = append(diagnostics, nr.checkRootUser(pod.Spec.Containers, pod)...)
		diagnostics = append(diagnostics, nr.checkRootUser(pod.Spec.InitContainers, pod)...)
	}

	return diagnostics, nil
}

func (nr *nonRootUserCheck) checkRootUser(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	for _, container := range containers {
		podRunAsRoot := pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsNonRoot == nil || !*pod.Spec.SecurityContext.RunAsNonRoot
		containerRunAsRoot := container.SecurityContext == nil || container.SecurityContext.RunAsNonRoot == nil || !*container.SecurityContext.RunAsNonRoot

		if containerRunAsRoot && podRunAsRoot {
			d := checks.Diagnostic{
				Check:    nr.Name(),
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Container `%s` can run as root user. Please ensure that the image is from a trusted source.", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}
