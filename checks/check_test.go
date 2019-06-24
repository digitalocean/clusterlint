package checks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckIsDisabled(t *testing.T) {
	const name string = "pod_foo"
	pod := initPod(name)
	assert.False(t, IsEnabled(name, pod.ObjectMeta))
}

func TestCheckIsEnabled(t *testing.T) {
	const name string = "pod_foo"
	pod := initPod("")
	assert.True(t, IsEnabled(name, pod.ObjectMeta))
}

func TestNoClusterlintAnnotation(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod_foo", Annotations: map[string]string{"porg": "star wars"},
		},
	}
	assert.True(t, IsEnabled("pod_foo", pod.ObjectMeta))
}

func initPod(name string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod_foo", Annotations: map[string]string{"clusterlint.disable.checks": fmt.Sprintf("%s, bar, baz", name)},
		},
	}
	return pod
}
