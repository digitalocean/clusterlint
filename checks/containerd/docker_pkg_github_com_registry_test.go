/*
Copyright 2021 DigitalOcean

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

package containerd

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
)

func TestDomainNameCheckMeta(t *testing.T) {
	domainNameCheck := domainNameCheck{}
	assert.Equal(t, "docker-pkg-github-com-registry", domainNameCheck.Name())
	assert.Equal(t, []string{"containerd", "doks"}, domainNameCheck.Groups())
	assert.NotEmpty(t, domainNameCheck.Description())
}

func TestDomainNameCheckRegistration(t *testing.T) {
	domainNameCheck := &domainNameCheck{}
	check, err := checks.Get("docker-pkg-github-com-registry")
	assert.NoError(t, err)
	assert.Equal(t, check, domainNameCheck)
}

func TestDockerPkgGithubComRegistry(t *testing.T) {
	const message = "containerd can't pull images from docker.pkg.github.com, used by container 'bar'"
	const invalidMessage = "Image name for container 'bar' could not be parsed"
	const severity = checks.Error
	const name = "docker-pkg-github-com-registry"

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
			name:     "pod with container image - docker.pkg.github.com/busybox:latest",
			objs:     container("docker.pkg.github.com/busybox:latest"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - docker.pkg.github.com/busybox",
			objs:     container("docker.pkg.github.com/busybox"),
			expected: issues(severity, message, checks.Pod, name),
		},
		{
			name:     "pod with container image - test:5000/repo",
			objs:     container("test:5000/repo/image"),
			expected: nil,
		},
		{
			name:     "pod with container image - ghcr.io/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			objs:     container("ghcr.io/repo:ignore-tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: nil,
		},
		{
			name:     "pod with container image - ghcr.io/busybox:v1.2.3",
			objs:     container("ghcr.io/busybox:v1.2.3"),
			expected: nil,
		},
		{
			name:     "pod with init container with invalid image name",
			objs:     initContainer(""),
			expected: issues(checks.Warning, invalidMessage, checks.Pod, name),
		},
		{
			name:     "pod with container with invalid image name",
			objs:     container(""),
			expected: issues(checks.Warning, invalidMessage, checks.Pod, name),
		},
	}

	domainNameCheck := domainNameCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := domainNameCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}
