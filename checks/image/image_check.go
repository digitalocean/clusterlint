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

package image

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&containerImageCheck{})
}

type containerImageCheck struct{}

// Name returns a unique name for this check.
func (pc *containerImageCheck) Name() string {
	return "container-images"
}

// Groups returns a list of group names this check should be part of.
func (pc *containerImageCheck) Groups() []string {
	return []string{"image"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (pc *containerImageCheck) Description() string {
	return "Checks if there are pods with containers in privileged mode"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (pc *containerImageCheck) Run(ctx context.Context, objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	repos := ctx.Value(checks.Repositories)
	validRepos := make(map[string]struct{})
	if repos != nil {
		repoArr, ok := repos.([]string)
		if ok {
			for _, r := range repoArr {
				validRepos[r] = struct{}{}
			}
		}
	}
	for _, pod := range objects.Pods.Items {
		diagnostics = append(diagnostics, pc.checkPod(validRepos, pod)...)
	}

	return diagnostics, nil
}

// checkPod checks if all pod images are well formed and
// do not have any errors, or obvious mistakes.
func (pc *containerImageCheck) checkPod(validRepos map[string]struct{}, pod corev1.Pod) []checks.Diagnostic {
	var diagnostics []checks.Diagnostic
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	for _, container := range containers {
		named, err := checkImageError(container.Name, container.Image)
		if err != nil {
			diagnostics = append(diagnostics, checks.Diagnostic{
				Check:    pc.Name(),
				Severity: checks.Error,
				Message:  err.Error(),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			})
		} else {
			for _, e := range lintImage(named, validRepos, container.Name, container.Image) {
				diagnostics = append(diagnostics, checks.Diagnostic{
					Check:    pc.Name(),
					Severity: checks.Warning,
					Message:  e.Error(),
					Kind:     checks.Pod,
					Object:   &pod.ObjectMeta,
					Owners:   pod.ObjectMeta.GetOwnerReferences(),
				})
			}
		}
	}
	return diagnostics
}

func checkImageError(containerName, image string) (reference.Named, error) {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, fmt.Errorf("%q's container image, %q, is malformed: %w", containerName, image, err)
	}
	return named, nil
}

const dockerIO = "docker.io"

var (
	errEmptyDomain   = errors.New("empty domain")
	errInvalidDomain = errors.New("invalid domain")
	errNoSha         = errors.New("no sha")
)

func lintImage(named reference.Named, validRepos map[string]struct{}, containerName, image string) (warnings []error) {
	domain := reference.Domain(named)
	// the reference lib inserts "docker.io" into the image if it is not present,
	// we remove it if we find that it isn't specified.
	if domain == dockerIO && strings.Index(image, dockerIO) != 0 {
		domain = ""
	}
	if domain == "" {
		warnings = append(warnings, fmt.Errorf("%q's container image, %q, does not specify a domain, the runtime may not be able to resolve this image: %w", containerName, image, errEmptyDomain))
	} else if validRepos != nil {
		if _, ok := validRepos[domain]; len(validRepos) > 0 && !ok {
			var repos []string
			for k := range validRepos {
				repos = append(repos, k)
			}
			warnings = append(warnings, fmt.Errorf("%q's container image, %q, does not have a domain that is in the list of accepted domains (%q): %w", containerName, image, strings.Join(repos, "\", \""), errInvalidDomain))
		}
	}

	_, ok := named.(reference.Digested)
	if !ok {
		warnings = append(warnings, fmt.Errorf("%q's container image, %q, does not have a valid sha digest, this makes the container runtime vulnerable to MIM attacks: %w", containerName, image, errNoSha))
	}

	return
}
