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
	"testing"

	csitypes "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	csitypesbeta "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func TestInvalidSnapshotCheckMeta(t *testing.T) {
	invalidSnapshotCheck := invalidSnapshotCheck{}
	assert.Equal(t, "invalid-volume-snapshot", invalidSnapshotCheck.Name())
	assert.Equal(t, []string{"doks"}, invalidSnapshotCheck.Groups())
	assert.NotEmpty(t, invalidSnapshotCheck.Description())
}

func TestInvalidSnapshotCheckRegistration(t *testing.T) {
	invalidSnapshotCheck := &invalidSnapshotCheck{}
	check, err := checks.Get("invalid-volume-snapshot")
	assert.NoError(t, err)
	assert.Equal(t, check, invalidSnapshotCheck)
}

func TestInvalidSnapshots(t *testing.T) {
	invalidSnapshotCheck := invalidSnapshotCheck{}
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "valid snapshot",
			objs:     emptyValidSnapshots(),
			expected: nil,
		},
		{
			name:     "invalid snapshots",
			objs:     emptyInvalidSnapshots(),
			expected: expectedSnapshotErrors(emptyInvalidSnapshots(), invalidSnapshotCheck.Name()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := invalidSnapshotCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func emptyValidSnapshots() *kube.Objects {
	objs := &kube.Objects{
		VolumeSnapshotsV1: &csitypes.VolumeSnapshotList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypes.VolumeSnapshot{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       csitypes.VolumeSnapshotSpec{},
				},
			},
		},
		VolumeSnapshotsBeta: &csitypesbeta.VolumeSnapshotList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypesbeta.VolumeSnapshot{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       csitypesbeta.VolumeSnapshotSpec{},
				},
			},
		},
	}
	return objs
}

func emptyInvalidSnapshots() *kube.Objects {
	objs := &kube.Objects{
		VolumeSnapshotsV1: &csitypes.VolumeSnapshotList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypes.VolumeSnapshot{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"snapshot.storage.kubernetes.io/invalid-snapshot-resource": "",
						},
					},
					Spec: csitypes.VolumeSnapshotSpec{},
				},
			},
		},
		VolumeSnapshotsBeta: &csitypesbeta.VolumeSnapshotList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypesbeta.VolumeSnapshot{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"snapshot.storage.kubernetes.io/invalid-snapshot-resource": "",
						},
					},
					Spec: csitypesbeta.VolumeSnapshotSpec{},
				},
			},
		},
	}
	return objs
}

func expectedSnapshotErrors(objs *kube.Objects, name string) []checks.Diagnostic {
	v1snap := objs.VolumeSnapshotsV1.Items[0]
	betasnap := objs.VolumeSnapshotsBeta.Items[0]
	errMsg := "Snapshot has been marked as invalid by CSI validation - check persistentVolumeClaimName and volumeSnapshotContentName are not both set"
	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  errMsg,
			Kind:     checks.VolumeSnapshot,
			Object:   &v1snap.ObjectMeta,
			Owners:   v1snap.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Error,
			Message:  errMsg,
			Kind:     checks.VolumeSnapshot,
			Object:   &betasnap.ObjectMeta,
			Owners:   betasnap.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
