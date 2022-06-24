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
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&unusedClaimCheck{})
}

type unusedClaimCheck struct{}

// Name returns a unique name for this check.
func (c *unusedClaimCheck) Name() string {
	return "unused-pvc"
}

// Groups returns a list of group names this check should be part of.
func (c *unusedClaimCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (c *unusedClaimCheck) Description() string {
	return "Check if there are unused persistent volume claims in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (c *unusedClaimCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	used := make(map[kube.Identifier]struct{})
	var empty struct{}
	for _, pod := range objects.Pods.Items {
		for _, volume := range pod.Spec.Volumes {
			claim := volume.VolumeSource.PersistentVolumeClaim
			if claim != nil {
				used[kube.Identifier{Name: claim.ClaimName, Namespace: pod.GetNamespace()}] = empty
			}
		}
	}

	for _, claim := range objects.PersistentVolumeClaims.Items {
		if _, ok := used[kube.Identifier{Name: claim.GetName(), Namespace: claim.GetNamespace()}]; !ok {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused persistent volume claim",
				Kind:     checks.PersistentVolumeClaim,
				Object:   &claim.ObjectMeta,
				Owners:   claim.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
