package basic

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
)

func TestFullyQualifiedImageCheckMeta(t *testing.T) {
	fullyQualifiedImageCheck := fullyQualifiedImageCheck{}
	assert.Equal(t, "fully-qualified-image", fullyQualifiedImageCheck.Name())
	assert.Equal(t, "Checks if containers have fully qualified image names", fullyQualifiedImageCheck.Description())
	assert.Equal(t, []string{"basic"}, fullyQualifiedImageCheck.Groups())
}

func TestFullyQualifiedImageCheckRegistration(t *testing.T) {
	fullyQualifiedImageCheck := &fullyQualifiedImageCheck{}
	check, err := checks.Get("fully-qualified-image")
	assert.Equal(t, check, fullyQualifiedImageCheck)
	assert.Nil(t, err)
}

func TestFullyQualifiedImageWarning(t *testing.T) {
	const message string = "Use fully qualified image for container 'bar' in pod 'pod_foo' in namespace 'k8s'"
	const category string = "warning"

	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []kube.Diagnostic
	}{
		{
			name:     "no pods",
			arg:      initPod(),
			expected: nil,
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox:latest",
			arg:      container("k8s.gcr.io/busybox:1.2.3"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox:latest",
			arg:      container("busybox:latest"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox",
			arg:      container("k8s.gcr.io/busybox"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox",
			arg:      container("busybox"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox:latest",
			arg:      initContainer("k8s.gcr.io/busybox:1.2.3"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox:latest",
			arg:      initContainer("busybox:latest"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox",
			arg:      initContainer("k8s.gcr.io/busybox"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox",
			arg:      initContainer("busybox"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: issues(category, message),
		},
		{
			name:     "pod with container image - test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: issues(category, message),
		},
	}

	fullyQualifiedImageCheck := fullyQualifiedImageCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d, err := fullyQualifiedImageCheck.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
		})
	}
}

func TestMalformedImageError(t *testing.T) {
	const message string = "Malformed image name for container 'bar' in pod 'pod_foo' in namespace 'k8s'"
	const category string = "error"

	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []kube.Diagnostic
	}{
		{
			name:     "container with image : test:5000/repo/image@sha256:digest",
			arg:      container("test:5000/repo/image@sha256:digest"),
			expected: issues(category, message),
		},
		{
			name:     "init container with image : test:5000/repo/image@sha256:digest",
			arg:      initContainer("test:5000/repo/image@sha256:digest"),
			expected: issues(category, message),
		},
	}
	fullyQualifiedImageCheck := fullyQualifiedImageCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d, err := fullyQualifiedImageCheck.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
		})
	}
}
