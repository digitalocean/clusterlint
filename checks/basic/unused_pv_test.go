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
	"context"
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnusedPVCheckMeta(t *testing.T) {
	unusedPVCheck := unusedPVCheck{}
	assert.Equal(t, "unused-pv", unusedPVCheck.Name())
	assert.Equal(t, []string{"basic"}, unusedPVCheck.Groups())
	assert.NotEmpty(t, unusedPVCheck.Description())
}

func TestUnusedPVCheckRegistration(t *testing.T) {
	unusedPVCheck := &unusedPVCheck{}
	check, err := checks.Get("unused-pv")
	assert.NoError(t, err)
	assert.Equal(t, check, unusedPVCheck)
}

func TestUnusedPVWarning(t *testing.T) {
	unusedPVCheck := unusedPVCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pvs",
			objs:     &kube.Objects{PersistentVolumes: &corev1.PersistentVolumeList{}},
			expected: nil,
		},
		{
			name:     "bound pv",
			objs:     bound(),
			expected: nil,
		},
		{
			name: "unused pv",
			objs: unused(),
			expected: []checks.Diagnostic{
				{
					Check:    unusedPVCheck.Name(),
					Severity: checks.Warning,
					Message:  "Unused Persistent Volume 'pv_foo'.",
					Kind:     checks.PersistentVolume,
					Object:   &metav1.ObjectMeta{Name: "pv_foo"},
					Owners:   GetOwners(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := unusedPVCheck.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initPV() *kube.Objects {
	objs := &kube.Objects{
		PersistentVolumes: &corev1.PersistentVolumeList{
			Items: []corev1.PersistentVolume{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pv_foo"},
				},
			},
		},
	}
	return objs
}

func bound() *kube.Objects {
	objs := initPV()
	objs.PersistentVolumes.Items[0].Spec = corev1.PersistentVolumeSpec{
		ClaimRef: &corev1.ObjectReference{
			Kind:      "PersistentVolumeClaim",
			Name:      "foo",
			Namespace: "k8s",
		},
	}
	return objs
}

func unused() *kube.Objects {
	objs := initPV()
	objs.PersistentVolumes.Items[0].Spec = corev1.PersistentVolumeSpec{}
	return objs
}
