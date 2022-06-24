/*
Copyright 2022 DigitalOcean

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
	"fmt"

	batchv1beta1 "k8s.io/api/batch/v1beta1"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&cronJobConcurrencyCheck{})
}

type cronJobConcurrencyCheck struct{}

// Name returns a unique name for this check.
func (c *cronJobConcurrencyCheck) Name() string {
	return "cronjob-concurrency"
}

// Groups returns a list of group names this check should be part of.
func (c *cronJobConcurrencyCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (c *cronJobConcurrencyCheck) Description() string {
	return "Check if any cronjobs have a concurrency policy of 'Allow'"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (c *cronJobConcurrencyCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, cronjob := range objects.CronJobs.Items {
		if batchv1beta1.AllowConcurrent == cronjob.Spec.ConcurrencyPolicy {
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  fmt.Sprintf("CronJob has a concurrency policy of `%s`. Prefer to use `%s` or `%s`", cronjob.Spec.ConcurrencyPolicy, batchv1beta1.ForbidConcurrent, batchv1beta1.ReplaceConcurrent),
				Kind:     checks.CronJob,
				Object:   &cronjob.ObjectMeta,
				Owners:   cronjob.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}

	return diagnostics, nil
}
