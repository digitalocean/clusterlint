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
	Register(&alwaysFail{})

	filter := CheckFilter{
		IncludeChecks: []string{"always-fail"},
	}
	client := &kube.Client{
		KubeClient: fake.NewSimpleClientset(),
	}
	client.KubeClient.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	})

	alwaysFailCheck, err := Get("always-fail")
	assert.NoError(t, err)

	result, err := Run(context.Background(), client, filter, DiagnosticFilter{},kube.ObjectsFilter{})
	assert.NoError(t, err)
	assert.Len(t, result.Diagnostics, 1)
	assert.Equal(t, alwaysFailCheck.Name(), result.Diagnostics[0].Check)
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
