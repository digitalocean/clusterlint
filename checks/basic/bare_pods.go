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
	"strings"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&barePodCheck{})
}

type barePodCheck struct{}

// Name returns a unique name for this check.
func (b *barePodCheck) Name() string {
	return "bare-pods"
}

// Groups returns a list of group names this check should be part of.
func (b *barePodCheck) Groups() []string {
	return []string{"basic", "doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (b *barePodCheck) Description() string {
	return "Check if there are bare pods in the cluster"
}

// Run runs this check on a set of Kubernetes objects.
func (b *barePodCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	for _, pod := range objects.Pods.Items {
		pod := pod
		if len(pod.ObjectMeta.OwnerReferences) == 0 {
			// skip static pod
			if objects.Nodes != nil {
				if isStaticPod(pod, objects.Nodes.Items) {
					continue
				}
			}
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Avoid using bare pods in clusters",
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}

func isStaticPod(pod corev1.Pod, nodeList []corev1.Node) bool {
	for _, node := range nodeList {
		// https://github.com/kubernetes/kubernetes/blob/b409073e99695ea35642a8194b9285ac12fd0cf8/pkg/kubelet/config/common.go#L51
		if strings.HasSuffix(pod.Name, "-"+strings.ToLower(node.Name)) {
			return true
		}
	}
	return false
}
