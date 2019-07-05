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
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return []string{"doks"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (w *webhookCheck) Description() string {
	return "Check for admission controllers that could prevent managed components from starting"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (w *webhookCheck) Run(data *checks.CheckData) ([]checks.Diagnostic, error) {
	objects := data.Objects
	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurations.Items {
		for _, validatingWebhook := range config.Webhooks {
			if *validatingWebhook.FailurePolicy == ar.Fail &&
				validatingWebhook.ClientConfig.Service != nil &&
				selectorMatchesNamespace(validatingWebhook.NamespaceSelector, objects.SystemNamespace) {
				d := checks.Diagnostic{
					Severity: checks.Error,
					Message:  "Webhook matches objects in the kube-system namespace. This can cause problems when upgrading the cluster.",
					Kind:     checks.ValidatingWebhookConfiguration,
					Object:   &config.ObjectMeta,
					Owners:   config.ObjectMeta.GetOwnerReferences(),
				}
				diagnostics = append(diagnostics, d)
			}
		}
	}

	for _, config := range objects.MutatingWebhookConfigurations.Items {
		for _, mutatingWebhook := range config.Webhooks {
			if *mutatingWebhook.FailurePolicy == ar.Fail &&
				mutatingWebhook.ClientConfig.Service != nil &&
				selectorMatchesNamespace(mutatingWebhook.NamespaceSelector, objects.SystemNamespace) {
				d := checks.Diagnostic{
					Severity: checks.Error,
					Message:  "Webhook matches objects in the kube-system namespace. This can cause problems when upgrading the cluster.",
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
