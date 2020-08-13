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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	checks.Register(&webhookReplacementCheck{})
}

type webhookReplacementCheck struct{}

// Name returns a unique name for this check.
func (w *webhookReplacementCheck) Name() string {
	return "admission-controller-webhook-replacement"
}

// Groups returns a list of group names this check should be part of.
func (w *webhookReplacementCheck) Groups() []string {
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *webhookReplacementCheck) Description() string {
	return "Check for admission control webhooks that could cause problems during upgrades or node replacement"
}

// Run runs this check on a set of Kubernetes objects.
func (w *webhookReplacementCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	const apiserverServiceName = "kubernetes"

	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurations.Items {
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
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

	for _, config := range objects.MutatingWebhookConfigurations.Items {
		config := config
		for _, wh := range config.Webhooks {
			wh := wh
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

func selectorMatchesNamespace(selector *metav1.LabelSelector, namespace *corev1.Namespace) bool {
	if selector.Size() == 0 {
		return true
	}
	labels := namespace.GetLabels()
	for key, value := range selector.MatchLabels {
		if v, ok := labels[key]; !ok || v != value {
			return false
		}
	}
	for _, lbr := range selector.MatchExpressions {
		if !match(labels, lbr) {
			return false
		}
	}
	return true
}

func match(labels map[string]string, lbr metav1.LabelSelectorRequirement) bool {
	switch lbr.Operator {
	case metav1.LabelSelectorOpExists:
		if _, ok := labels[lbr.Key]; ok {
			return true
		}
		return false
	case metav1.LabelSelectorOpDoesNotExist:
		if _, ok := labels[lbr.Key]; !ok {
			return true
		}
		return false
	case metav1.LabelSelectorOpIn:
		if v, ok := labels[lbr.Key]; ok && contains(lbr.Values, v) {
			return true
		}
		return false
	case metav1.LabelSelectorOpNotIn:
		if v, ok := labels[lbr.Key]; !ok || !contains(lbr.Values, v) {
			return true
		}
		return false
	}
	return false
}

func contains(list []string, name string) bool {
	for _, l := range list {
		if l == name {
			return true
		}
	}
	return false
}
