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

package doks

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&invalidSnapshotContentCheck{})
}

type invalidSnapshotContentCheck struct{}

// Name returns a unique name for this check.
func (i *invalidSnapshotContentCheck) Name() string {
	return "invalid-volume-snapshot-content"
}

// Groups returns a list of group names this check should be part of.
func (i *invalidSnapshotContentCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (i *invalidSnapshotContentCheck) Description() string {
	return "Checks if there are invalid volume snapshot contents that would fail webhook validation"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (i *invalidSnapshotContentCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	errMsg := "Snapshot content has been marked as invalid by CSI validation - check volumeHandle and snapshotHandle are not both set"
	ssLabelKey := "snapshot.storage.sigs.k8s.io/invalid-snapshot-content-resource"
	for _, snapshot := range objects.VolumeSnapshotsV1Content.Items {
		snapshotLabels := snapshot.Labels
		if _, ok := snapshotLabels[ssLabelKey]; ok {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  errMsg,
				Kind:     checks.VolumeSnapshotContent,
				Object:   &snapshot.ObjectMeta,
				Owners:   snapshot.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	for _, snapshot := range objects.VolumeSnapshotsBetaContent.Items {
		snapshotLabels := snapshot.Labels
		if _, ok := snapshotLabels[ssLabelKey]; ok {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  errMsg,
				Kind:     checks.VolumeSnapshotContent,
				Object:   &snapshot.ObjectMeta,
				Owners:   snapshot.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}
