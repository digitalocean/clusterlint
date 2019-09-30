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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBarePodCheckMeta(t *testing.T) {
	barePodCheck := barePodCheck{}
	assert.Equal(t, "bare-pods", barePodCheck.Name())
	assert.Equal(t, []string{"basic", "doks"}, barePodCheck.Groups())
	assert.NotEmpty(t, barePodCheck.Description())
}

func TestBarePodCheckRegistration(t *testing.T) {
	barePodCheck := &barePodCheck{}
	check, err := checks.Get(barePodCheck.Name())
	assert.NoError(t, err, "Should not fail registration")
	assert.Equal(t, check, barePodCheck)
}

func TestBarePodError(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pods",
			objs:     &kube.Objects{Pods: &corev1.PodList{}},
			expected: nil,
		},
		{
			name:     "pod has owner ref",
			objs:     initRefs(initPod()),
			expected: nil,
		},
		{
			name:     "multiple pods with owner refs",
			objs:     initRefs(initMultiplePods()),
			expected: nil,
		},
		{
			name: "pod has no owner ref",
			objs: initPod(),
			expected: []checks.Diagnostic{
				{
					Severity: "warning",
					Check:    "bare-pods",
					Kind:     checks.Pod,
					Message:  "Avoid using bare pods in clusters",
					Object:   GetObjectMeta(),
					Owners:   nil,
				},
			},
		},
		{
			name: "multiple pods with no owner ref",
			objs: initMultiplePods(),
			expected: []checks.Diagnostic{
				{
					Severity: "warning",
					Check:    "bare-pods",
					Kind:     checks.Pod,
					Message:  "Avoid using bare pods in clusters",
					Object:   &metav1.ObjectMeta{Name: "pod_1", Namespace: "k8s"},
					Owners:   nil,
				},
				{
					Severity: "warning",
					Check:    "bare-pods",
					Kind:     checks.Pod,
					Message:  "Avoid using bare pods in clusters",
					Object:   &metav1.ObjectMeta{Name: "pod_2", Namespace: "k8s"},
					Owners:   nil,
				},
			},
		},
	}

	barePodCheck := &barePodCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := barePodCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initRefs(objs *kube.Objects) *kube.Objects {
	for index, _ := range objs.Pods.Items {
		objs.Pods.Items[index].ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				Name:       "Deployment",
				APIVersion: "apps/v1",
			},
		}
	}
	return objs
}
