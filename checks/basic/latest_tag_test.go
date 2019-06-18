package basic

import (
	"fmt"
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLatestTagCheckMeta(t *testing.T) {
	latestTagCheck := latestTagCheck{}
	assert.Equal(t, "latest-tag", latestTagCheck.Name())
	assert.Equal(t, "Checks if there are pods with container images having latest tag", latestTagCheck.Description())
	assert.Equal(t, []string{"basic"}, latestTagCheck.Groups())
}

func TestLatestTagCheckRegistration(t *testing.T) {
	latestTagCheck := &latestTagCheck{}
	check, err := checks.Get("latest-tag")
	assert.Equal(t, check, latestTagCheck)
	assert.Nil(t, err)
}

func TestLatestTagWarning(t *testing.T) {
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
			name:     "pod with container image - k8s.gcr.io/busybox:latest",
			arg:      container("k8s.gcr.io/busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - busybox:latest",
			arg:      container("busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox",
			arg:      container("k8s.gcr.io/busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - busybox",
			arg:      container("busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - private:5000/repo/busybox",
			arg:      container("http://private:5000/repo/busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - private:5000/repo/busybox:latest",
			arg:      container("http://private:5000/repo/busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:digest",
			arg:      container("test:5000/repo@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:digest",
			arg:      container("repo@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:digest",
			arg:      container("test:5000/repo:ignore-tag@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox:v1.2.3",
			arg:      container("busybox:v1.2.3"),
			expected: nil,
		},

		{
			name:     "pod with init container image - k8s.gcr.io/busybox:latest",
			arg:      initContainer("k8s.gcr.io/busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with init container image - busybox:latest",
			arg:      initContainer("busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with init container image - k8s.gcr.io/busybox",
			arg:      initContainer("k8s.gcr.io/busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with init container image - busybox",
			arg:      initContainer("busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - http://private:5000/repo/busybox",
			arg:      container("http://private:5000/repo/busybox"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - http://private:5000/repo/busybox:latest",
			arg:      container("http://private:5000/repo/busybox:latest"),
			expected: warnings(),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:digest",
			arg:      initContainer("test:5000/repo@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:digest",
			arg:      initContainer("test:5000/repo:ignore-tag@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:digest",
			arg:      initContainer("repo@sha256:digest"),
			expected: nil,
		},
		{
			name:     "pod with init container image - busybox:v1.2.3",
			arg:      initContainer("busybox:v1.2.3"),
			expected: nil,
		},
	}

	latestTagCheck := latestTagCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			w, e, err := latestTagCheck.Run(scenario.arg)
			assert.ElementsMatch(t, scenario.expected, w)
			assert.Empty(t, e)
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

func container(image string) *kube.Objects {
	objs := initPod()
	objs.Pods = &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "bar", Image: image}}},
			},
		},
	}
	return objs
}

func initContainer(image string) *kube.Objects {
	objs := initPod()
	objs.Pods = &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				Spec:       corev1.PodSpec{InitContainers: []corev1.Container{{Name: "bar", Image: image}}},
			},
		},
	}
	return objs
}

func warnings() []error {
	w := []error{
		fmt.Errorf("[Best Practice] Use specific tags instead of latest for container 'bar' in pod 'pod_foo' in namespace 'k8s'"),
	}
	return w
}
