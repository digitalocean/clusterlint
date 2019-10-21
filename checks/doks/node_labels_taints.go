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
	"context"
	"strings"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&nodeLabelsTaintsCheck{})
}

type nodeLabelsTaintsCheck struct{}

// Name returns the name of the check.
func (*nodeLabelsTaintsCheck) Name() string {
	return "node-labels-and-taints"
}

// Groups returns groups for this check.
func (*nodeLabelsTaintsCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a description of the check.
func (*nodeLabelsTaintsCheck) Description() string {
	return "Checks that nodes do not have custom labels or taints configured."
}

// Run runs the check.
func (c *nodeLabelsTaintsCheck) Run(_ context.Context, objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, node := range objects.Nodes.Items {
		for labelKey := range node.Labels {
			if !isKubernetesLabel(labelKey) && !isDOKSLabel(labelKey) {
				d := checks.Diagnostic{
					Check:    c.Name(),
					Severity: checks.Warning,
					Message:  "Custom node labels will be lost if node is replaced or upgraded.",
					Kind:     checks.Node,
					Object:   &node.ObjectMeta,
				}
				diagnostics = append(diagnostics, d)
				// Produce only one label diagnostic per node.
				break
			}
		}
		for _, taint := range node.Spec.Taints {
			if !isDOKSTaint(taint) {
				d := checks.Diagnostic{
					Check:    c.Name(),
					Severity: checks.Warning,
					Message:  "Custom node taints will be lost if node is replaced or upgraded.",
					Kind:     checks.Node,
					Object:   &node.ObjectMeta,
				}
				diagnostics = append(diagnostics, d)
				// Produce only one taint diagnostic per node.
				break
			}
		}
	}

	return diagnostics, nil
}

func isKubernetesLabel(key string) bool {
	// Built-in Kubernetes labels are in various subdomains of
	// kubernetes.io. Assume all such labels are built in.
	return strings.Contains(key, corev1.ResourceDefaultNamespacePrefix)
}

func isDOKSLabel(key string) bool {
	// DOKS labels use the doks.digitalocean.com namespace. Assume all such
	// labels are set by DOKS.
	if strings.HasPrefix(key, "doks.digitalocean.com/") {
		return true
	}

	// CCM also sets a region label.
	if key == "region" {
		return true
	}

	return false
}

func isDOKSTaint(taint corev1.Taint) bool {
	// Currently DOKS never sets taints.
	return false
}
