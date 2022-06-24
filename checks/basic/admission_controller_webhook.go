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

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	v1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&webhookCheck{})
}

type webhookCheck struct{}

// Name returns a unique name for this check.
func (w *webhookCheck) Name() string {
	return "admission-controller-webhook"
}

// Groups returns a list of group names this check should be part of.
func (w *webhookCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *webhookCheck) Description() string {
	return "Check for admission control webhooks"
}

// Run runs this check on a set of Kubernetes objects.
func (w *webhookCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	const apiserverServiceName = "kubernetes"

	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurations.Items {
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

	for _, config := range objects.MutatingWebhookConfigurations.Items {
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

func namespaceExists(namespaceList *v1.NamespaceList, namespace string) bool {
	for _, ns := range namespaceList.Items {
		if ns.Name == namespace {
			return true
		}
	}
	return false
}

func serviceExists(serviceList *v1.ServiceList, service, namespace string) bool {
	for _, svc := range serviceList.Items {
		if svc.Name == service && svc.Namespace == namespace {
			return true
		}
	}
	return false
}
