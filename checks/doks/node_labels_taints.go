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
	"fmt"
	"sort"
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
func (c *nodeLabelsTaintsCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, node := range objects.Nodes.Items {
		var customLabels, customTaints []string
		for labelKey := range node.Labels {
			if !isKubernetesLabel(labelKey) && !isDOKSLabel(labelKey) {
				customLabels = append(customLabels, labelKey)
			}
		}
		if len(customLabels) > 0 {
			// The order of the map iteration above is non-deterministic, so
			// sort the labels for stable output.
			sort.Strings(customLabels)
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Custom node labels will be lost if node is replaced or upgraded. Add custom labels on node pools instead.",
				Kind:     checks.Node,
				Object:   &node.ObjectMeta,
				Details:  fmt.Sprintf("Custom node labels: %s", customLabels),
			}
			diagnostics = append(diagnostics, d)
		}
		for _, taint := range node.Spec.Taints {
			if !isDOKSTaint(taint) {
				customTaints = append(customTaints, taint.Key)
			}
		}
		if len(customTaints) > 0 {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Custom node taints will be lost if node is replaced or upgraded.",
				Kind:     checks.Node,
				Object:   &node.ObjectMeta,
				Details:  fmt.Sprintf("Custom node taints: %s", customTaints),
			}
			diagnostics = append(diagnostics, d)
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
