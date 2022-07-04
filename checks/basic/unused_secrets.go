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
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&unusedSecretCheck{})
}

type unusedSecretCheck struct{}

type identifier struct {
	Name      string
	Namespace string
}

// Name returns a unique name for this check.
func (s *unusedSecretCheck) Name() string {
	return "unused-secret"
}

// Groups returns a list of group names this check should be part of.
func (s *unusedSecretCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (s *unusedSecretCheck) Description() string {
	return "Checks if there are unused secrets in the cluster. Ignores service account tokens"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (s *unusedSecretCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	used, err := checkReferences(objects)
	if err != nil {
		return nil, err
	}

	for _, secret := range filter(objects.Secrets.Items) {
		if _, ok := used[kube.Identifier{Name: secret.GetName(), Namespace: secret.GetNamespace()}]; !ok {
			secret := secret
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused secret",
				Kind:     checks.Secret,
				Object:   &secret.ObjectMeta,
				Owners:   secret.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}

//checkReferences checks each pod for config map references in volumes and environment variables
func checkReferences(objects *kube.Objects) (map[kube.Identifier]struct{}, error) {
	used := make(map[kube.Identifier]struct{})
	var empty struct{}
	var mu sync.Mutex
	var g errgroup.Group
	for _, pod := range objects.Pods.Items {
		pod := pod
		namespace := pod.GetNamespace()
		g.Go(func() error {
			for _, volume := range pod.Spec.Volumes {
				s := volume.VolumeSource.Secret
				if s != nil {
					mu.Lock()
					used[kube.Identifier{Name: s.SecretName, Namespace: namespace}] = empty
					mu.Unlock()
				}
				if volume.VolumeSource.Projected != nil {
					for _, source := range volume.VolumeSource.Projected.Sources {
						s := source.Secret
						if s != nil {
							mu.Lock()
							used[kube.Identifier{Name: s.LocalObjectReference.Name, Namespace: namespace}] = empty
							mu.Unlock()
						}
					}
				}
			}
			for _, imageSecret := range pod.Spec.ImagePullSecrets {
				mu.Lock()
				used[kube.Identifier{Name: imageSecret.Name, Namespace: namespace}] = empty
				mu.Unlock()
			}
			identifiers := envVarsSecretRefs(pod.Spec.Containers, namespace)
			identifiers = append(identifiers, envVarsSecretRefs(pod.Spec.InitContainers, namespace)...)
			mu.Lock()
			for _, i := range identifiers {
				used[i] = empty
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, sa := range objects.ServiceAccounts.Items {
		sa := sa
		namespace := sa.Namespace

		g.Go(func() error {
			for _, imageSecret := range sa.ImagePullSecrets {
				mu.Lock()
				used[kube.Identifier{Name: imageSecret.Name, Namespace: namespace}] = empty
				mu.Unlock()
			}

			for _, secret := range sa.Secrets {
				mu.Lock()
				used[kube.Identifier{Name: secret.Name, Namespace: namespace}] = empty
				mu.Unlock()
			}
			return nil
		})
	}

	return used, g.Wait()
}

// envVarsSecretRefs checks for secret references in container environment variables
func envVarsSecretRefs(containers []corev1.Container, namespace string) []kube.Identifier {
	var refs []kube.Identifier
	for _, container := range containers {
		for _, env := range container.EnvFrom {
			if env.SecretRef != nil {
				refs = append(refs, kube.Identifier{Name: env.SecretRef.LocalObjectReference.Name, Namespace: namespace})
			}
		}
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				refs = append(refs, kube.Identifier{Name: env.ValueFrom.SecretKeyRef.LocalObjectReference.Name, Namespace: namespace})
			}
		}
	}
	return refs
}

// filter returns Secrets that are not of type `kubernetes.io/service-account-token`
func filter(secrets []corev1.Secret) []corev1.Secret {
	var filtered []corev1.Secret
	for _, secret := range secrets {
		if secret.Type != corev1.SecretTypeServiceAccountToken {
			filtered = append(filtered, secret)
		}
	}
	return filtered
}
