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

package security

import (
	"context"
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
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
	privilegedContainerCheck := privilegedContainerCheck{}

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
			objs:     containerPrivileged(true),
			expected: warnings(containerPrivileged(true), privilegedContainerCheck.Name()),
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
			objs:     containerPrivileged(false),
			expected: nil,
		},
		{
			name:     "pod with init container in privileged mode",
			objs:     initContainerPrivileged(true),
			expected: warnings(initContainerPrivileged(true), privilegedContainerCheck.Name()),
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
			objs:     initContainerPrivileged(false),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := privilegedContainerCheck.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func warnings(objs *kube.Objects, name string) []checks.Diagnostic {
	pod := objs.Pods.Items[0]
	d := []checks.Diagnostic{
		{
			Check:    name,
			Severity: checks.Warning,
			Message:  "Privileged container 'bar' found. Please ensure that the image is from a trusted source.",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
