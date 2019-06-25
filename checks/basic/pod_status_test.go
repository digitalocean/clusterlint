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
	assert.Equal(t, "Check if there are unhealthy pods in the cluster", podStatusCheck.Description())
	assert.Equal(t, []string{"workload-health"}, podStatusCheck.Groups())
}

func TestPodStateError(t *testing.T) {
	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []checks.Diagnostic
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
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod 'pod_foo' in namespace 'k8s' has state: Failed. Pod state should be `Running`, `Pending` or `Succeeded`.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
		{
			name: "pod with unknown status",
			arg:  status(corev1.PodUnknown),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod 'pod_foo' in namespace 'k8s' has state: Unknown. Pod state should be `Running`, `Pending` or `Succeeded`.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
	}

	podStatusCheck := podStatusCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d, err := podStatusCheck.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
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
