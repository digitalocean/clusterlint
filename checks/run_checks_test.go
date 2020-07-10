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

package checks

import (
	"context"
	"testing"

	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRun(t *testing.T) {
	tests := []struct{
		name string
		check string
		expectedErr string
		expectedDiagnostics int
	}{
		{
			name: "test failure",
			check: "always-fail",
			expectedDiagnostics: 1,
		},
		{
			name: "test panic",
			check: "panic-check",
			expectedErr: "Recovered from panic in check 'panic-check':",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Register(&alwaysFail{})
			Register(&panicCheck{})
			filter := CheckFilter{
				IncludeChecks: []string{test.check},
			}

			client := &kube.Client{
				KubeClient: fake.NewSimpleClientset(),
			}
			client.KubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
				},
			}, metav1.CreateOptions{})

			check, err := Get(test.check)
			assert.NoError(t, err)

			result, err := Run(context.Background(), client, filter, DiagnosticFilter{}, kube.ObjectFilter{})
			if test.expectedErr == "" {
				assert.NoError(t, err)
				assert.Len(t, result.Diagnostics, 1)
				assert.Equal(t, check.Name(), result.Diagnostics[0].Check)
			} else {
				assert.Contains(t, err.Error(), test.expectedErr)
				assert.Nil(t, result)
			}
		})
	}
}

type alwaysFail struct{}

// Name returns a unique name for this check.
func (nc *alwaysFail) Name() string {
	return "always-fail"
}

// Groups returns a list of group names this check should be part of.
func (nc *alwaysFail) Groups() []string {
	return nil
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *alwaysFail) Description() string {
	return "Does not check anything. Always returns an error.."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *alwaysFail) Run(*kube.Objects) ([]Diagnostic, error) {
	return []Diagnostic{{
		Message:  "This check always produces an error.",
		Severity: Error,
		Kind:     Pod,
		Object:   &metav1.ObjectMeta{},
	}}, nil
}

type panicCheck struct {}

// Name returns a unique name for this check.
func (nc *panicCheck) Name() string {
	return "panic-check"
}

// Groups returns a list of group names this check should be part of.
func (nc *panicCheck) Groups() []string {
	return nil
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *panicCheck) Description() string {
	return "Does not check anything. Panics.."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *panicCheck) Run(*kube.Objects) ([]Diagnostic, error) {
	type some struct {
		x int
	}
	var s *some
	_ = s.x
	return nil, nil
}
