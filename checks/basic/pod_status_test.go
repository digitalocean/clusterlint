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
)

func TestMeta(t *testing.T) {
	podStatusCheck := podStatusCheck{}
	assert.Equal(t, "pod-state", podStatusCheck.Name())
	assert.Equal(t, []string{"workload-health"}, podStatusCheck.Groups())
	assert.NotEmpty(t, podStatusCheck.Description())
}

func TestPodStateCheckRegistration(t *testing.T) {
	podStatusCheck := &podStatusCheck{}
	check, err := checks.Get("pod-state")
	assert.NoError(t, err)
	assert.Equal(t, check, podStatusCheck)
}

func TestPodStateError(t *testing.T) {
	podStatusCheck := podStatusCheck{}
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
			name:     "pod with running status",
			objs:     status(corev1.PodRunning),
			expected: nil,
		},
		{
			name:     "pod with pending status",
			objs:     status(corev1.PodPending),
			expected: nil,
		},
		{
			name:     "pod with succeeded status",
			objs:     status(corev1.PodSucceeded),
			expected: nil,
		},
		{
			name: "pod with failed status",
			objs: status(corev1.PodFailed),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Unhealthy pod. State: `Failed`. Pod state should be `Running`, `Pending` or `Succeeded`.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
		{
			name: "pod with unknown status",
			objs: status(corev1.PodUnknown),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Unhealthy pod. State: `Unknown`. Pod state should be `Running`, `Pending` or `Succeeded`.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := podStatusCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func status(status corev1.PodPhase) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Status = corev1.PodStatus{
		Phase: status,
	}
	return objs
}
