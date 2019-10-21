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

package doks

import (
	"context"
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodSelectorCheckMeta(t *testing.T) {
	podSelectorCheck := podSelectorCheck{}
	assert.Equal(t, "node-name-pod-selector", podSelectorCheck.Name())
	assert.Equal(t, []string{"doks"}, podSelectorCheck.Groups())
	assert.NotEmpty(t, podSelectorCheck.Description())
}

func TestPodSelectorCheckRegistration(t *testing.T) {
	podSelectorCheck := &podSelectorCheck{}
	check, err := checks.Get("node-name-pod-selector")
	assert.NoError(t, err)
	assert.Equal(t, check, podSelectorCheck)
}

func TestNodeNameError(t *testing.T) {
	podSelectorCheck := podSelectorCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no node name selector",
			objs:     empty(),
			expected: nil,
		},
		{
			name:     "node name used in node selector",
			objs:     invalidPod(),
			expected: expectedWarnings(invalidPod(), podSelectorCheck.Name()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := podSelectorCheck.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func empty() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{},
	}
	return objs
}

func invalidPod() *kube.Objects {
	objs := empty()
	objs.Pods = &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				Spec:       corev1.PodSpec{NodeSelector: map[string]string{corev1.LabelHostname: "foo"}},
			},
		},
	}
	return objs
}

func expectedWarnings(objs *kube.Objects, name string) []checks.Diagnostic {
	pod := objs.Pods.Items[0]
	diagnostics := []checks.Diagnostic{
		{
			Check:    name,
			Severity: checks.Warning,
			Message:  "Avoid node name label for node selector.",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
