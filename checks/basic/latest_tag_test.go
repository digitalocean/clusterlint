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
	assert.Equal(t, []string{"basic"}, latestTagCheck.Groups())
	assert.NotEmpty(t, latestTagCheck.Description())
}

func TestLatestTagCheckRegistration(t *testing.T) {
	latestTagCheck := &latestTagCheck{}
	check, err := checks.Get("latest-tag")
	assert.NoError(t, err)
	assert.Equal(t, check, latestTagCheck)
}

func TestLatestTagWarning(t *testing.T) {
	const message = "Avoid using latest tag for container 'bar'"
	const severity = checks.Warning
	const name = "latest-tag"

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
			name:     "pod with container image - k8s.gcr.io/busybox:latest",
			objs:     container("k8s.gcr.io/busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - busybox:latest",
			objs:     container("busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - k8s.gcr.io/busybox",
			objs:     container("k8s.gcr.io/busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - busybox",
			objs:     container("busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - private:5000/busybox",
			objs:     container("private:5000/repo/busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - private:5000/busybox:latest",
			objs:     container("private:5000/repo/busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     container("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     container("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     container("test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - busybox:v1.2.3",
			objs:     container("busybox:v1.2.3"),
			expected: nil,
		},

		{
			name:     "pod with init container image - k8s.gcr.io/busybox:latest",
			objs:     initContainer("k8s.gcr.io/busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with init container image - busybox:latest",
			objs:     initContainer("busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with init container image - k8s.gcr.io/busybox",
			objs:     initContainer("k8s.gcr.io/busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with init container image - busybox",
			objs:     initContainer("busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - private:5000/busybox",
			objs:     container("private:5000/repo/busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - private:5000/busybox:latest",
			objs:     container("private:5000/repo/busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     initContainer("test:5000/repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - test:5000/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     initContainer("test:5000/repo/image:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     initContainer("repo/image@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with init container image - busybox:v1.2.3",
			objs:     initContainer("busybox:v1.2.3"),
			expected: nil,
		},
	}

	latestTagCheck := latestTagCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := latestTagCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}
