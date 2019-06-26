package security

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPrivilegedContainersCheckMeta(t *testing.T) {
	privilegedContainerCheck := privilegedContainerCheck{}
	assert.Equal(t, "privileged-containers", privilegedContainerCheck.Name())
	assert.Equal(t, []string{"security"}, privilegedContainerCheck.Groups())
	assert.NotEmpty(t, privilegedContainerCheck.Description())
}

func TestPrivilegedContainersCheckRegistration(t *testing.T) {
	privilegedContainerCheck := &privilegedContainerCheck{}
	check, err := checks.Get("privileged-containers")
	assert.NoError(t, err)
	assert.Equal(t, check, privilegedContainerCheck)
}

func TestPrivilegedContainerWarning(t *testing.T) {
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
			name:     "pod with container in privileged mode",
			objs:     container(true),
			expected: warnings(container(true)),
		},
		{
			name:     "pod with container.SecurityContext = nil",
			objs:     containerSecurityContextNil(),
			expected: nil,
		},
		{
			name:     "pod with container.SecurityContext.Privileged = nil",
			objs:     containerPrivilegedNil(),
			expected: nil,
		},
		{
			name:     "pod with container in regular mode",
			objs:     container(false),
			expected: nil,
		},
		{
			name:     "pod with init container in privileged mode",
			objs:     initContainer(true),
			expected: warnings(initContainer(true)),
		},
		{
			name:     "pod with initContainer.SecurityContext = nil",
			objs:     initContainerSecurityContextNil(),
			expected: nil,
		},
		{
			name:     "pod with initContainer.SecurityContext.Privileged = nil",
			objs:     initContainerPrivilegedNil(),
			expected: nil,
		},
		{
			name:     "pod with init container in regular mode",
			objs:     initContainer(false),
			expected: nil,
		},
	}

	privilegedContainerCheck := privilegedContainerCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := privilegedContainerCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initPod() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func container(privileged bool) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
			}},
	}
	return objs
}

func containerSecurityContextNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "bar",
			}},
	}
	return objs
}

func containerPrivilegedNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{},
			}},
	}
	return objs
}

func initContainer(privileged bool) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
			}},
	}
	return objs
}

func initContainerSecurityContextNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "bar",
			}},
	}
	return objs
}

func initContainerPrivilegedNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{},
			}},
	}
	return objs
}

func warnings(objs *kube.Objects) []checks.Diagnostic {
	pod := objs.Pods.Items[0]
	d := []checks.Diagnostic{
		{
			Severity: checks.Warning,
			Message:  "Privileged container 'bar' found. Please ensure that the image is from a trusted source.",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
