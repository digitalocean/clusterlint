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
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&betaWebhookCheck{})
}

type betaWebhookCheck struct{}

// Name returns a unique name for this check.
func (w *betaWebhookCheck) Name() string {
	return "admission-controller-webhook-v1beta1"
}

// Groups returns a list of group names this check should be part of.
func (w *betaWebhookCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *betaWebhookCheck) Description() string {
	return "Check for admission control webhooks"
}

// Run runs this check on a set of Kubernetes objects.
func (w *betaWebhookCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	if len(objects.ValidatingWebhookConfigurations.Items) > 0 ||
		len(objects.MutatingWebhookConfigurations.Items) > 0 {
		// Skip this check if there are v1 webhook configurations. On clusters
		// that support both v1beta1 and v1 admission control, the same webhook
		// configurations will be returned for both versions.
		return nil, nil
	}

	const apiserverServiceName = "kubernetes"

	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurationsBeta.Items {
		for _, wh := range config.Webhooks {
			if wh.ClientConfig.Service != nil {
				// Ensure that the service (and its namespace) that is configure actually exists.

				if !namespaceExists(objects.Namespaces, wh.ClientConfig.Service.Namespace) {
					diagnostics = append(diagnostics, checks.Diagnostic{
						Severity: checks.Error,
						Message:  fmt.Sprintf("Validating webhook %s is configured against a service in a namespace that does not exist.", wh.Name),
						Kind:     checks.ValidatingWebhookConfiguration,
						Object:   &config.ObjectMeta,
						Owners:   config.ObjectMeta.GetOwnerReferences(),
					})
					continue
				}

				if !serviceExists(objects.Services, wh.ClientConfig.Service.Name, wh.ClientConfig.Service.Namespace) {
					diagnostics = append(diagnostics, checks.Diagnostic{
						Severity: checks.Error,
						Message:  fmt.Sprintf("Validating webhook %s is configured against a service that does not exist.", wh.Name),
						Kind:     checks.ValidatingWebhookConfiguration,
						Object:   &config.ObjectMeta,
						Owners:   config.ObjectMeta.GetOwnerReferences(),
					})
				}
			}
		}
	}

	for _, config := range objects.MutatingWebhookConfigurationsBeta.Items {
		for _, wh := range config.Webhooks {
			if wh.ClientConfig.Service != nil {
				// Ensure that the service (and its namespace) that is configure actually exists.

				if !namespaceExists(objects.Namespaces, wh.ClientConfig.Service.Namespace) {
					diagnostics = append(diagnostics, checks.Diagnostic{
						Severity: checks.Error,
						Message:  fmt.Sprintf("Mutating webhook %s is configured against a service in a namespace that does not exist.", wh.Name),
						Kind:     checks.MutatingWebhookConfiguration,
						Object:   &config.ObjectMeta,
						Owners:   config.ObjectMeta.GetOwnerReferences(),
					})
					continue
				}

				if !serviceExists(objects.Services, wh.ClientConfig.Service.Name, wh.ClientConfig.Service.Namespace) {
					diagnostics = append(diagnostics, checks.Diagnostic{
						Severity: checks.Error,
						Message:  fmt.Sprintf("Mutating webhook %s is configured against a service that does not exist.", wh.Name),
						Kind:     checks.MutatingWebhookConfiguration,
						Object:   &config.ObjectMeta,
						Owners:   config.ObjectMeta.GetOwnerReferences(),
					})
				}
			}
		}
	}
	return diagnostics, nil
}
