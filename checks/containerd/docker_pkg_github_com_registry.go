/*
Copyright 2021 DigitalOcean

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

package containerd

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&domainNameCheck{})
}

type domainNameCheck struct{}

// Name returns a unique name for this check.
func (l *domainNameCheck) Name() string {
	return "docker-pkg-github-com-registry"
}

// Groups returns a list of group names this check should be part of.
func (l *domainNameCheck) Groups() []string {
	return []string{"containerd", "doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (l *domainNameCheck) Description() string {
	return "Checks if there are pods with container images that are hosted at the docker.pkg.github.com registry"
}

// Run runs this check on a set of Kubernetes objects. It can return errors
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (l *domainNameCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pod := range objects.Pods.Items {
		diagnostics = append(diagnostics, l.checkTags(pod.Spec.Containers, pod)...)
		diagnostics = append(diagnostics, l.checkTags(pod.Spec.InitContainers, pod)...)
	}

	return diagnostics, nil
}

// checkTags checks if the image registry is `docker.pkg.github.com`
// Adds an error if it finds any image that comes from that registry
func (l *domainNameCheck) checkTags(containers []corev1.Container, pod corev1.Pod) []checks.Diagnostic {
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
		domainNameOnly := reference.Domain(namedRef)
		if domainNameOnly == "docker.pkg.github.com" {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  fmt.Sprintf("containerd can't pull images from docker.pkg.github.com, used by container '%s'", container.Name),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}
