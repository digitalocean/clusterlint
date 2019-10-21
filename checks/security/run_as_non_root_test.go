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
	corev1 "k8s.io/api/core/v1"
)

func TestNonRootUserCheckMeta(t *testing.T) {
	nonRootUserCheck := nonRootUserCheck{}
	assert.Equal(t, "non-root-user", nonRootUserCheck.Name())
	assert.Equal(t, []string{"security"}, nonRootUserCheck.Groups())
	assert.NotEmpty(t, nonRootUserCheck.Description())
}

func TestNonRootUserCheckRegistration(t *testing.T) {
	nonRootUserCheck := &nonRootUserCheck{}
	check, err := checks.Get("non-root-user")
	assert.NoError(t, err)
	assert.Equal(t, check, nonRootUserCheck)
}

func TestNonRootUserWarning(t *testing.T) {
	nonRootUserCheck := nonRootUserCheck{}
	trueVar := true
	falseVar := false

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pods",
			objs:     &kube.Objects{Pods: &corev1.PodList{}},
			expected: nil,
		},
		{
			name:     "pod security context and container security context unset",
			objs:     containerSecurityContextNil(),
			expected: diagnostic(),
		},
		{
			name:     "pod security context unset, container with run as non root set to true",
			objs:     containerNonRoot(nil, &trueVar),
			expected: nil,
		},
		{
			name:     "pod security context unset, container with run as non root set to false",
			objs:     containerNonRoot(nil, &falseVar),
			expected: diagnostic(),
		},
		{
			name:     "pod run as non root true, container run as non root true",
			objs:     containerNonRoot(&trueVar, &trueVar),
			expected: nil,
		},
		{
			name:     "pod run as non root true, container run as non root false",
			objs:     containerNonRoot(&trueVar, &falseVar),
			expected: nil,
		},
		{
			name:     "pod run as non root false, container run as non root true",
			objs:     containerNonRoot(&falseVar, &trueVar),
			expected: nil,
		},
		{
			name:     "pod run as non root false, container run as non root false",
			objs:     containerNonRoot(&falseVar, &falseVar),
			expected: diagnostic(),
		},
		{
			name:     "pod run as non root true, container security context unset",
			objs:     containerNonRoot(&trueVar, nil),
			expected: nil,
		},
		{
			name:     "pod run as non root false, container security context unset",
			objs:     containerNonRoot(&falseVar, nil),
			expected: diagnostic(),
		},
		// init container tests

		{
			name:     "pod security context and init container security context unset",
			objs:     initContainerSecurityContextNil(),
			expected: diagnostic(),
		},
		{
			name:     "pod security context unset, init container with run as non root set to true",
			objs:     initContainerNonRoot(nil, &trueVar),
			expected: nil,
		},
		{
			name:     "pod security context unset, init container with run as non root set to false",
			objs:     initContainerNonRoot(nil, &falseVar),
			expected: diagnostic(),
		},
		{
			name:     "pod run as non root true, init container run as non root true",
			objs:     initContainerNonRoot(&trueVar, &trueVar),
			expected: nil,
		},
		{
			name:     "pod run as non root true, init container run as non root false",
			objs:     initContainerNonRoot(&trueVar, &falseVar),
			expected: nil,
		},
		{
			name:     "pod run as non root false, init container run as non root true",
			objs:     initContainerNonRoot(&falseVar, &trueVar),
			expected: nil,
		},
		{
			name:     "pod run as non root false, init container run as non root false",
			objs:     initContainerNonRoot(&falseVar, &falseVar),
			expected: diagnostic(),
		},
		{
			name:     "pod run as non root true, init container security context unset",
			objs:     initContainerNonRoot(&trueVar, nil),
			expected: nil,
		},
		{
			name:     "pod run as non root false, init container security context unset",
			objs:     initContainerNonRoot(&falseVar, nil),
			expected: diagnostic(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := nonRootUserCheck.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func diagnostic() []checks.Diagnostic {
	pod := initPod().Pods.Items[0]
	d := []checks.Diagnostic{
		{
			Check:    "non-root-user",
			Severity: checks.Warning,
			Message:  "Container `bar` can run as root user. Please ensure that the image is from a trusted source.",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
