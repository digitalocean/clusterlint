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
)

func init() {
	checks.Register(&unusedPVCheck{})
}

type unusedPVCheck struct{}

// Name returns a unique name for this check.
func (pv *unusedPVCheck) Name() string {
	return "unused-pv"
}

// Groups returns a list of group names this check should be part of.
func (pv *unusedPVCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (pv *unusedPVCheck) Description() string {
	return "Check if there are unused persistent volumes in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (pv *unusedPVCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pv := range objects.PersistentVolumes.Items {
		if pv.Spec.ClaimRef == nil {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Unused Persistent Volume '%s'.", pv.GetName()),
				Kind:     checks.PersistentVolume,
				Object:   &pv.ObjectMeta,
				Owners:   pv.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
