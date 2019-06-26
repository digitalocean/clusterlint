package doks

import (
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
			expected: errors(invalidPod()),
		},
	}

	podSelectorCheck := podSelectorCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := podSelectorCheck.Run(test.objs)
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

func errors(objs *kube.Objects) []checks.Diagnostic {
	pod := objs.Pods.Items[0]
	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  "Avoid node name label for node selector.",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
