package doks

import (
	"fmt"
	"testing"

	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroup(t *testing.T) {
	podSelectorCheck := podSelectorCheck{}
	assert.Equal(t, []string{"doks"}, podSelectorCheck.Groups())
}

func TestNodeNameError(t *testing.T) {
	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []error
	}{
		{
			name:     "no node name selector",
			arg:      empty(),
			expected: nil,
		},
		{
			name:     "node name used in node selector",
			arg:      invalidPod(),
			expected: errors(),
		},
	}

	podSelectorCheck := podSelectorCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			_, e, _ := podSelectorCheck.Run(scenario.arg)
			assert.ElementsMatch(t, scenario.expected, e)
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

func errors() []error {
	e := []error{
		fmt.Errorf("pod 'pod_foo' in namespace 'k8s' uses the node name for node selector"),
	}
	return e
}
