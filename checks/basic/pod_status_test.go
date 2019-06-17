package basic

import (
	"fmt"
	"testing"

	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMeta(t *testing.T) {
	podStatusCheck := podStatusCheck{}
	assert.Equal(t, "pod-state", podStatusCheck.Name())
	assert.Equal(t, "Check if there are unhealthy pods in the cluster", podStatusCheck.Description())
	assert.Equal(t, []string{"basic"}, podStatusCheck.Groups())
}

func TestPodStateError(t *testing.T) {
	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []error
	}{
		{
			name:     "no pods",
			arg:      initPod(),
			expected: nil,
		},
		{
			name:     "pod with running status",
			arg:      status(corev1.PodRunning),
			expected: nil,
		},
		{
			name:     "pod with pending status",
			arg:      status(corev1.PodPending),
			expected: nil,
		},
		{
			name:     "pod with succeeded status",
			arg:      status(corev1.PodSucceeded),
			expected: nil,
		},
		{
			name: "pod with failed status",
			arg:  status(corev1.PodFailed),
			expected: []error{
				fmt.Errorf("Pod 'pod_foo' in namespace 'k8s' has state: Failed. Pod state should be `Running`, `Pending` or `Succeeded`."),
			},
		},
		{
			name: "pod with unknown status",
			arg:  status(corev1.PodUnknown),
			expected: []error{
				fmt.Errorf("Pod 'pod_foo' in namespace 'k8s' has state: Unknown. Pod state should be `Running`, `Pending` or `Succeeded`."),
			},
		},
	}

	podStatusCheck := podStatusCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			w, e, err := podStatusCheck.Run(scenario.arg)
			assert.ElementsMatch(t, scenario.expected, e)
			assert.Empty(t, w)
			assert.Nil(t, err)
		})
	}
}

func initPod() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{},
	}
	return objs
}

func status(status corev1.PodPhase) *kube.Objects {
	objs := initPod()
	objs.Pods = &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				Status:     corev1.PodStatus{Phase: status},
			},
		},
	}
	return objs
}
