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

package doks

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeLabels(t *testing.T) {
	tests := []struct {
		name                string
		nodeLabels          map[string]string
		expectedDiagnostics []checks.Diagnostic
	}{
		{
			name:                "no labels",
			nodeLabels:          nil,
			expectedDiagnostics: nil,
		},
		{
			name: "only doks labels",
			nodeLabels: map[string]string{
				"doks.digitalocean.com/foo": "bar",
				"doks.digitalocean.com/baz": "xyzzy",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "only built-in labels",
			nodeLabels: map[string]string{
				"kubernetes.io/hostname":                   "a-hostname",
				"beta.kubernetes.io/os":                    "linux",
				"failure-domain.beta.kubernetes.io/region": "tor1",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "only region label",
			nodeLabels: map[string]string{
				"region": "tor1",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "custom labels",
			nodeLabels: map[string]string{
				"doks.digitalocean.com/foo":                "bar",
				"doks.digitalocean.com/baz":                "xyzzy",
				"kubernetes.io/hostname":                   "a-hostname",
				"example.com/custom-label":                 "bad",
				"example.com/another-label":                "real-bad",
				"beta.kubernetes.io/os":                    "linux",
				"failure-domain.beta.kubernetes.io/region": "tor1",
				"region": "tor1",
			},
			expectedDiagnostics: []checks.Diagnostic{{
				Severity: checks.Warning,
				Message:  "Custom node labels will be lost if node is replaced or upgraded. Add custom labels on node pools instead.",
				Kind:     checks.Node,
				Details:  "Custom node labels: [example.com/another-label example.com/custom-label]",
				Object: &metav1.ObjectMeta{
					Labels: map[string]string{
						"doks.digitalocean.com/foo":                "bar",
						"doks.digitalocean.com/baz":                "xyzzy",
						"kubernetes.io/hostname":                   "a-hostname",
						"example.com/custom-label":                 "bad",
						"example.com/another-label":                "real-bad",
						"beta.kubernetes.io/os":                    "linux",
						"failure-domain.beta.kubernetes.io/region": "tor1",
						"region": "tor1",
					},
				},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objects := &kube.Objects{
				Nodes: &corev1.NodeList{
					Items: []corev1.Node{{
						ObjectMeta: metav1.ObjectMeta{
							Labels: test.nodeLabels,
						},
					}},
				},
			}

			check := &nodeLabelsTaintsCheck{}

			ds, err := check.Run(objects)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expectedDiagnostics, ds)
		})
	}
}

func TestNodeTaints(t *testing.T) {
	tests := []struct {
		name                string
		taints              []corev1.Taint
		expectedDiagnostics []checks.Diagnostic
	}{
		{
			name:                "no taints",
			taints:              nil,
			expectedDiagnostics: nil,
		},
		{
			name: "custom taints",
			taints: []corev1.Taint{{
				Key:    "example.com/my-taint",
				Value:  "foo",
				Effect: corev1.TaintEffectNoSchedule,
			}},
			expectedDiagnostics: []checks.Diagnostic{{
				Severity: checks.Warning,
				Details:  "Custom node taints: [example.com/my-taint]",
				Message:  "Custom node taints will be lost if node is replaced or upgraded.",
				Kind:     checks.Node,
				Object:   &metav1.ObjectMeta{},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objects := &kube.Objects{
				Nodes: &corev1.NodeList{
					Items: []corev1.Node{{
						Spec: corev1.NodeSpec{
							Taints: test.taints,
						},
					}},
				},
			}

			check := &nodeLabelsTaintsCheck{}

			ds, err := check.Run(objects)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expectedDiagnostics, ds)
		})
	}
}
