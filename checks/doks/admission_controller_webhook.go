package doks

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
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
func (w *webhookCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, config := range objects.ValidatingWebhookConfigurations.Items {
		for _, validatingWebhook := range config.Webhooks {
			if *validatingWebhook.FailurePolicy == ar.Fail && doesSelectorIncludeKubeSystem(validatingWebhook.NamespaceSelector, objects.SystemNamespace) {
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
			if *mutatingWebhook.FailurePolicy == ar.Fail && doesSelectorIncludeKubeSystem(mutatingWebhook.NamespaceSelector, objects.SystemNamespace) {
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

func doesSelectorIncludeKubeSystem(selector *metav1.LabelSelector, namespace *corev1.Namespace) bool {
	if selector.Size() == 0 {
		return true
	}
	labels := namespace.GetLabels()
	for key, value := range selector.MatchLabels {
		if v, ok := labels[key]; ok && v == value {
			continue
		}
		return false
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
