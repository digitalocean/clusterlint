/*
Copyright 2020 DigitalOcean

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
	checks.Register(&unusedCMCheck{})
}

type unusedCMCheck struct{}

// Name returns a unique name for this check.
func (c *unusedCMCheck) Name() string {
	return "unused-config-map"
}

// Groups returns a list of group names this check should be part of.
func (c *unusedCMCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (c *unusedCMCheck) Description() string {
	return "Checks if there are unused config maps in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (c *unusedCMCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	used, err := checkPodReferences(objects)
	if err != nil {
		return nil, err
	}

	nodeRefs, err := checkNodeReferences(objects)
	if err != nil {
		return nil, err
	}

	for k, v := range nodeRefs {
		used[k] = v
	}

	for _, cm := range objects.ConfigMaps.Items {
		if _, ok := used[kube.Identifier{Name: cm.GetName(), Namespace: cm.GetNamespace()}]; !ok {
			cm := cm
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused config map",
				Kind:     checks.ConfigMap,
				Object:   &cm.ObjectMeta,
				Owners:   cm.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}

func checkNodeReferences(objects *kube.Objects) (map[kube.Identifier]struct{}, error) {
	used := make(map[kube.Identifier]struct{})
	var empty struct{}
	var mu sync.Mutex
	var g errgroup.Group
	for _, node := range objects.Nodes.Items {
		node := node
		g.Go(func() error {
			source := node.Spec.ConfigSource
			if source != nil {
				mu.Lock()
				used[kube.Identifier{Name: source.ConfigMap.Name, Namespace: source.ConfigMap.Namespace}] = empty
				mu.Unlock()
			}
			return nil
		})
	}
	return used, g.Wait()
}

//checkPodReferences checks each pod for config map references in volumes and environment variables
func checkPodReferences(objects *kube.Objects) (map[kube.Identifier]struct{}, error) {
	used := make(map[kube.Identifier]struct{})
	var empty struct{}
	var mu sync.Mutex
	var g errgroup.Group
	for _, pod := range objects.Pods.Items {
		pod := pod
		namespace := pod.GetNamespace()
		g.Go(func() error {
			for _, volume := range pod.Spec.Volumes {
				cm := volume.VolumeSource.ConfigMap
				if cm != nil {
					mu.Lock()
					used[kube.Identifier{Name: cm.LocalObjectReference.Name, Namespace: namespace}] = empty
					mu.Unlock()
				}
				if volume.VolumeSource.Projected != nil {
					for _, source := range volume.VolumeSource.Projected.Sources {
						cm := source.ConfigMap
						if cm != nil {
							mu.Lock()
							used[kube.Identifier{Name: cm.LocalObjectReference.Name, Namespace: namespace}] = empty
							mu.Unlock()
						}
					}
				}
			}
			identifiers := checkEnvVars(pod.Spec.Containers, namespace)
			identifiers = append(identifiers, checkEnvVars(pod.Spec.InitContainers, namespace)...)
			mu.Lock()
			for _, i := range identifiers {
				used[i] = empty
			}
			mu.Unlock()

			return nil
		})
	}

	return used, g.Wait()
}

// checkEnvVars checks for config map references in container environment variables
func checkEnvVars(containers []corev1.Container, namespace string) []kube.Identifier {
	var refs []kube.Identifier
	for _, container := range containers {
		for _, env := range container.EnvFrom {
			if env.ConfigMapRef != nil {
				refs = append(refs, kube.Identifier{Name: env.ConfigMapRef.LocalObjectReference.Name, Namespace: namespace})
			}
		}
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
				refs = append(refs, kube.Identifier{Name: env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name, Namespace: namespace})
			}
		}
	}
	return refs
}
