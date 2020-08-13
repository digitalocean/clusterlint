/*
Copyright 2020 DigitalOcean

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
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&webhookTimeoutCheck{})
}

type webhookTimeoutCheck struct{}

// Name returns a unique name for this check.
func (w *webhookTimeoutCheck) Name() string {
	return "admission-controller-webhook-timeout"
}

// Groups returns a list of group names this check should be part of.
func (w *webhookTimeoutCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *webhookTimeoutCheck) Description() string {
	return "Check for admission control webhooks that have exceeded a timeout of 30 seconds."
}

// Run runs this check on a set of Kubernetes objects.
func (w *webhookTimeoutCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurations.Items {
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
			if wh.TimeoutSeconds == nil {
				// TimeoutSeconds value should be set to a non-nil value (greater than or equal to 1 and less than 30).
				// If the TimeoutSeconds value is set to nil and the cluster version is 1.13.*, users are
				// unable to configure the TimeoutSeconds value and this value will stay at nil, breaking
				// upgrades. It's only for versions >= 1.14 that the value will default to 30 seconds.
				continue
			} else if *wh.TimeoutSeconds < int32(1) || *wh.TimeoutSeconds >= int32(30) {
				// Webhooks with TimeoutSeconds set: less than 1 or greater than or equal to 30 is bad.
				d := checks.Diagnostic{
					Severity: checks.Error,
					Message:  "Validating webhook with a TimeoutSeconds value greater than 29 seconds will block upgrades.",
					Kind:     checks.ValidatingWebhookConfiguration,
					Object:   &config.ObjectMeta,
					Owners:   config.ObjectMeta.GetOwnerReferences(),
				}
				diagnostics = append(diagnostics, d)
			}
		}
	}

	for _, config := range objects.MutatingWebhookConfigurations.Items {
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
			if wh.TimeoutSeconds == nil {
				// TimeoutSeconds value should be set to a non-nil value (greater than or equal to 1 and less than 30).
				// If the TimeoutSeconds value is set to nil and the cluster version is 1.13.*, users are
				// unable to configure the TimeoutSeconds value and this value will stay at nil, breaking
				// upgrades. It's only for versions >= 1.14 that the value will default to 30 seconds.
				continue
			} else if *wh.TimeoutSeconds < int32(1) || *wh.TimeoutSeconds >= int32(30) {
				// Webhooks with TimeoutSeconds set: less than 1 or greater than or equal to 30 is bad.
				d := checks.Diagnostic{
					Severity: checks.Error,
					Message:  "Mutating webhook with a TimeoutSeconds value greater than 29 seconds will block upgrades.",
					Kind:     checks.MutatingWebhookConfiguration,
					Object:   &config.ObjectMeta,
					Owners:   config.ObjectMeta.GetOwnerReferences(),
				}
				diagnostics = append(diagnostics, d)
			}
		}
	}
	return diagnostics, nil
}
