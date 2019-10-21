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
	"context"
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&podStatusCheck{})
}

type podStatusCheck struct{}

// Name returns a unique name for this check.
func (p *podStatusCheck) Name() string {
	return "pod-state"
}

// Groups returns a list of group names this check should be part of.
func (p *podStatusCheck) Groups() []string {
	return []string{"workload-health"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (p *podStatusCheck) Description() string {
	return "Check if there are unhealthy pods in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (p *podStatusCheck) Run(_ context.Context, objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, pod := range objects.Pods.Items {
		if corev1.PodFailed == pod.Status.Phase || corev1.PodUnknown == pod.Status.Phase {
			d := checks.Diagnostic{
				Check:    p.Name(),
				Severity: checks.Warning,
				Message:  fmt.Sprintf("Unhealthy pod. State: `%s`. Pod state should be `Running`, `Pending` or `Succeeded`.", pod.Status.Phase),
				Kind:     checks.Pod,
				Object:   &pod.ObjectMeta,
				Owners:   pod.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
