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
	resource "k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceRequestsCheckMeta(t *testing.T) {
	resourceRequirementsCheck := resourceRequirementsCheck{}
	assert.Equal(t, "resource-requirements", resourceRequirementsCheck.Name())
	assert.Equal(t, []string{"basic", "doks"}, resourceRequirementsCheck.Groups())
	assert.NotEmpty(t, resourceRequirementsCheck.Description())
}

func TestResourceRequestsCheckRegistration(t *testing.T) {
	resourceRequirementsCheck := &resourceRequirementsCheck{}
	check, err := checks.Get("resource-requirements")
	assert.NoError(t, err)
	assert.Equal(t, check, resourceRequirementsCheck)
}

func TestResourceRequestsWarning(t *testing.T) {
	const message = "Set resource requests and limits for container `bar` to prevent resource contention"

	resourceRequirementsCheck := resourceRequirementsCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pods",
			objs:     initPod(),
			expected: nil,
		},
		{
			name: "container with no resource requests or limits",
			objs: container("alpine"),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  message,
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
		{
			name: "init container with no resource requests or limits",
			objs: initContainer("alpine"),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  message,
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
		{
			name:     "resource requests set",
			objs:     resources(),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := resourceRequirementsCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func resources() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "bar",
				Image: "alpine",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(500, "m"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(1000, "m"),
					},
				},
			}},
		InitContainers: []corev1.Container{
			{
				Name:  "bar",
				Image: "alpine",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(500, "m"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(1000, "m"),
					},
				},
			},
		},
	}
	return objs

}
