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
)

func TestHostpathCheckMeta(t *testing.T) {
	hostPathCheck := hostPathCheck{}
	assert.Equal(t, "hostpath-volume", hostPathCheck.Name())
	assert.Equal(t, []string{"basic", "doks"}, hostPathCheck.Groups())
	assert.NotEmpty(t, hostPathCheck.Description())
}

func TestHostpathCheckRegistration(t *testing.T) {
	hostPathCheck := &hostPathCheck{}
	check, err := checks.Get("hostpath-volume")
	assert.NoError(t, err)
	assert.Equal(t, check, hostPathCheck)
}

func TestHostpathVolumeError(t *testing.T) {
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
			name:     "pod with no volumes",
			objs:     container("docker.io/nginx:foo"),
			expected: nil,
		},
		{
			name: "pod with other volume",
			objs: volume(corev1.VolumeSource{
				GitRepo: &corev1.GitRepoVolumeSource{Repository: "boo"},
			}),
			expected: nil,
		},
		{
			name: "pod with hostpath volume",
			objs: volume(corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/tmp"},
			}),
			expected: []checks.Diagnostic{
				{
					Check:    "hostpath-volume",
					Severity: checks.Warning,
					Message:  "Avoid using hostpath for volume 'bar'.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
	}

	hostPathCheck := hostPathCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := hostPathCheck.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func volume(volumeSrc corev1.VolumeSource) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name:         "bar",
				VolumeSource: volumeSrc,
			}},
	}
	return objs
}
