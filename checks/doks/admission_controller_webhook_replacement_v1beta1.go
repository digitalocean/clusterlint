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
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	ar "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	checks.Register(&betaWebhookReplacementCheck{})
}

type betaWebhookReplacementCheck struct{}

// Name returns a unique name for this check.
func (w *betaWebhookReplacementCheck) Name() string {
	return "admission-controller-webhook-replacement-v1beta1"
}

// Groups returns a list of group names this check should be part of.
func (w *betaWebhookReplacementCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *betaWebhookReplacementCheck) Description() string {
	return "Check for admission control webhooks that could cause problems during upgrades or node replacement"
}

// Run runs this check on a set of Kubernetes objects.
func (w *betaWebhookReplacementCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
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
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
			if !applicableBeta(wh.Rules) {
				// Webhooks that do not apply to core/v1, apps/v1, apps/v1beta1, apps/v1beta2 resources are fine
				continue
			}
			if *wh.FailurePolicy == ar.Ignore {
				// Webhooks with failurePolicy: Ignore are fine.
				continue
			}
			if wh.ClientConfig.Service == nil {
				// Webhooks whose targets are external to the cluster are fine.
				continue
			}
			if wh.ClientConfig.Service.Namespace == metav1.NamespaceDefault &&
				wh.ClientConfig.Service.Name == apiserverServiceName {
				// Webhooks that target the kube-apiserver are fine.
				continue
			}
			if !selectorMatchesNamespace(wh.NamespaceSelector, objects.SystemNamespace) {
				// Webhooks that don't apply to kube-system are fine.
				continue
			}
			var svcNamespace *v1.Namespace
			for _, ns := range objects.Namespaces.Items {
				ns := ns
				if ns.Name == wh.ClientConfig.Service.Namespace {
					svcNamespace = &ns
				}
			}
			if svcNamespace != nil &&
				!selectorMatchesNamespace(wh.NamespaceSelector, svcNamespace) &&
				len(objects.Nodes.Items) > 1 {
				// Webhooks that don't apply to their own namespace are fine, as
				// long as there's more than one node in the cluster.
				continue
			}

			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  "Validating webhook is configured in such a way that it may be problematic during upgrades.",
				Kind:     checks.ValidatingWebhookConfiguration,
				Object:   &config.ObjectMeta,
				Owners:   config.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)

			// We don't want to produce diagnostics for multiple webhooks in the
			// same webhook configuration, so break out of the inner loop if we
			// get here.
			break
		}
	}

	for _, config := range objects.MutatingWebhookConfigurationsBeta.Items {
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
			if !applicableBeta(wh.Rules) {
				// Webhooks that do not apply to core/v1, apps/v1, apps/v1beta1, apps/v1beta2 resources are fine
				continue
			}
			if *wh.FailurePolicy == ar.Ignore {
				// Webhooks with failurePolicy: Ignore are fine.
				continue
			}
			if wh.ClientConfig.Service == nil {
				// Webhooks whose targets are external to the cluster are fine.
				continue
			}
			if wh.ClientConfig.Service.Namespace == metav1.NamespaceDefault &&
				wh.ClientConfig.Service.Name == apiserverServiceName {
				// Webhooks that target the kube-apiserver are fine.
				continue
			}
			if !selectorMatchesNamespace(wh.NamespaceSelector, objects.SystemNamespace) {
				// Webhooks that don't apply to kube-system are fine.
				continue
			}
			var svcNamespace *v1.Namespace
			for _, ns := range objects.Namespaces.Items {
				ns := ns
				if ns.Name == wh.ClientConfig.Service.Namespace {
					svcNamespace = &ns
				}
			}
			if svcNamespace != nil &&
				!selectorMatchesNamespace(wh.NamespaceSelector, svcNamespace) &&
				len(objects.Nodes.Items) > 1 {
				// Webhooks that don't apply to their own namespace are fine, as
				// long as there's more than one node in the cluster.
				continue
			}

			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  "Mutating webhook is configured in such a way that it may be problematic during upgrades.",
				Kind:     checks.MutatingWebhookConfiguration,
				Object:   &config.ObjectMeta,
				Owners:   config.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)

			// We don't want to produce diagnostics for multiple webhooks in the
			// same webhook configuration, so break out of the inner loop if we
			// get here.
			break
		}
	}
	return diagnostics, nil
}

func applicableBeta(rules []ar.RuleWithOperations) bool {
	for _, r := range rules {
		if apiVersions(r.APIVersions) {
			// applies to "apiVersions: v1"
			if len(r.APIGroups) == 0 {
				return true
			}
			// applies to "apiVersion: v1", "apiVersion: apps/v1", "apiVersion: apps/v1beta1", "apiVersion: apps/v1beta2"
			for _, g := range r.APIGroups {
				if g == "" || g == "*" || g == "apps" {
					return true
				}
			}
		}
	}
	return false
}
