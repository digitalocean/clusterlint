package basic

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
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
	const message string = "Avoid using latest tag for container 'bar' in pod 'pod_foo'"
	const severity checks.Severity = checks.Warning

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
			name:     "pod with container image - k8s.gcr.io/busybox:latest",
			arg:      container("k8s.gcr.io/busybox:latest"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - busybox:latest",
			arg:      container("busybox:latest"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox",
			arg:      container("k8s.gcr.io/busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - busybox",
			arg:      container("busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - private:5000/busybox",
			arg:      container("private:5000/repo/busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - private:5000/busybox:latest",
			arg:      container("private:5000/repo/busybox:latest"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      container("test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
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
			expected: issues(severity, message),
		},
		{
			name:     "pod with init container image - busybox:latest",
			arg:      initContainer("busybox:latest"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with init container image - k8s.gcr.io/busybox",
			arg:      initContainer("k8s.gcr.io/busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with init container image - busybox",
			arg:      initContainer("busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - private:5000/busybox",
			arg:      container("private:5000/repo/busybox"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - private:5000/busybox:latest",
			arg:      container("private:5000/repo/busybox:latest"),
			expected: issues(severity, message),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			arg:      initContainer("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
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
			d, err := latestTagCheck.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
		})
	}
}
