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

func TestInvalidSnapshotContentCheckMeta(t *testing.T) {
	invalidSnapshotContentCheck := invalidSnapshotContentCheck{}
	assert.Equal(t, "invalid-volume-snapshot-content", invalidSnapshotContentCheck.Name())
	assert.Equal(t, []string{"doks"}, invalidSnapshotContentCheck.Groups())
	assert.NotEmpty(t, invalidSnapshotContentCheck.Description())
}

func TestInvalidSnapshotContentCheckRegistration(t *testing.T) {
	invalidSnapshotContentCheck := &invalidSnapshotContentCheck{}
	check, err := checks.Get("invalid-volume-snapshot-content")
	assert.NoError(t, err)
	assert.Equal(t, check, invalidSnapshotContentCheck)
}

func TestInvalidSnapshotContents(t *testing.T) {
	invalidSnapshotContentCheck := invalidSnapshotContentCheck{}
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "valid snapshot contents",
			objs:     validSnapshotContent(),
			expected: nil,
		},
		{
			name:     "invalid snapshots contents",
			objs:     invalidSnapshotContents(),
			expected: expectedSnapshotContentErrors(invalidSnapshotContents(), invalidSnapshotContentCheck.Name()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := invalidSnapshotContentCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func validSnapshotContent() *kube.Objects {
	objs := &kube.Objects{
		VolumeSnapshotsV1Content: &csitypes.VolumeSnapshotContentList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypes.VolumeSnapshotContent{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       csitypes.VolumeSnapshotContentSpec{},
				},
			},
		},
		VolumeSnapshotsBetaContent: &csitypesbeta.VolumeSnapshotContentList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypesbeta.VolumeSnapshotContent{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       csitypesbeta.VolumeSnapshotContentSpec{},
				},
			},
		},
	}
	return objs
}

func invalidSnapshotContents() *kube.Objects {
	objs := &kube.Objects{
		VolumeSnapshotsV1Content: &csitypes.VolumeSnapshotContentList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypes.VolumeSnapshotContent{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"snapshot.storage.sigs.k8s.io/invalid-snapshot-content-resource": "",
						},
					},
					Spec: csitypes.VolumeSnapshotContentSpec{},
				},
			},
		},
		VolumeSnapshotsBetaContent: &csitypesbeta.VolumeSnapshotContentList{
			TypeMeta: metav1.TypeMeta{},
			Items: []csitypesbeta.VolumeSnapshotContent{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"snapshot.storage.sigs.k8s.io/invalid-snapshot-content-resource": "",
						},
					},
					Spec: csitypesbeta.VolumeSnapshotContentSpec{},
				},
			},
		},
	}
	return objs
}

func expectedSnapshotContentErrors(objs *kube.Objects, name string) []checks.Diagnostic {
	v1snap := objs.VolumeSnapshotsV1Content.Items[0]
	betasnap := objs.VolumeSnapshotsBetaContent.Items[0]
	errMsg := "Snapshot content has been marked as invalid by CSI validation - check volumeHandle and snapshotHandle are not both set"
	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  errMsg,
			Kind:     checks.VolumeSnapshotContent,
			Object:   &v1snap.ObjectMeta,
			Owners:   v1snap.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Error,
			Message:  errMsg,
			Kind:     checks.VolumeSnapshotContent,
			Object:   &betasnap.ObjectMeta,
			Owners:   betasnap.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
