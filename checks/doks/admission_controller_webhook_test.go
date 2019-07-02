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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWebhookCheckMeta(t *testing.T) {
	webhookCheck := webhookCheck{}
	assert.Equal(t, "admission-controller-webhook", webhookCheck.Name())
	assert.Equal(t, []string{"doks"}, webhookCheck.Groups())
	assert.NotEmpty(t, webhookCheck.Description())
}

func TestWebhookCheckRegistration(t *testing.T) {
	webhookCheck := &webhookCheck{}
	check, err := checks.Get("admission-controller-webhook")
	assert.NoError(t, err)
	assert.Equal(t, check, webhookCheck)
}

func TestWebhookError(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name: "no webhook configurations",
			objs: &kube.Objects{
				MutatingWebhookConfigurations:   &ar.MutatingWebhookConfigurationList{},
				ValidatingWebhookConfigurations: &ar.ValidatingWebhookConfigurationList{},
				SystemNamespace:                 &corev1.Namespace{},
			},
			expected: nil,
		},
		{
			name:     "failure policy is ignore",
			objs:     initObjects(ar.Ignore),
			expected: nil,
		},
		{
			name:     "webook does not use service",
			objs:     webhookURL(),
			expected: nil,
		},
		{
			name:     "namespace selector is empty",
			objs:     initObjects(ar.Fail),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector matches label",
			objs:     label(map[string]string{"doks_key": "bar"}),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector does not match label",
			objs:     label(map[string]string{"non-existent-label-on-namespace": "bar"}),
			expected: nil,
		},
		{
			name:     "namespace selector matches OpExists expression",
			objs:     expr("doks_key", []string{}, metav1.LabelSelectorOpExists),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector matches OpDoesNotExist expression",
			objs:     expr("random", []string{}, metav1.LabelSelectorOpDoesNotExist),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector matches OpIn expression",
			objs:     expr("doks_key", []string{"bar"}, metav1.LabelSelectorOpIn),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector matches OpNotIn expression",
			objs:     expr("doks_key", []string{"non-existent"}, metav1.LabelSelectorOpNotIn),
			expected: webhookErrors(),
		},
		{
			name:     "namespace selector does not match OpExists expression",
			objs:     expr("non-existent", []string{}, metav1.LabelSelectorOpExists),
			expected: nil,
		},
		{
			name:     "namespace selector does not match OpDoesNotExist expression",
			objs:     expr("doks_key", []string{}, metav1.LabelSelectorOpDoesNotExist),
			expected: nil,
		},
		{
			name:     "namespace selector does not match OpIn expression",
			objs:     expr("doks_key", []string{"non-existent"}, metav1.LabelSelectorOpIn),
			expected: nil,
		},
		{
			name:     "namespace selector does not match OpNotIn expression",
			objs:     expr("doks_key", []string{"bar"}, metav1.LabelSelectorOpNotIn),
			expected: nil,
		},
	}

	webhookCheck := webhookCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := webhookCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initObjects(failurePolicyType ar.FailurePolicyType) *kube.Objects {
	objs := &kube.Objects{
		SystemNamespace: &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kube-system",
				Labels: map[string]string{"doks_key": "bar"}},
		},
		MutatingWebhookConfigurations: &ar.MutatingWebhookConfigurationList{
			Items: []ar.MutatingWebhookConfiguration{
				{
					TypeMeta: metav1.TypeMeta{Kind: "MutatingWebhookConfiguration", APIVersion: "v1beta1"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "mwc_foo",
					},
					Webhooks: []ar.MutatingWebhook{
						{
							Name:              "mw_foo",
							FailurePolicy:     &failurePolicyType,
							NamespaceSelector: &metav1.LabelSelector{},
							ClientConfig: ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Name:      "some-svc",
									Namespace: "k8s",
								},
							},
						},
					},
				},
			},
		},
		ValidatingWebhookConfigurations: &ar.ValidatingWebhookConfigurationList{
			Items: []ar.ValidatingWebhookConfiguration{
				{
					TypeMeta: metav1.TypeMeta{Kind: "ValidatingWebhookConfiguration", APIVersion: "v1beta1"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "vwc_foo",
					},
					Webhooks: []ar.ValidatingWebhook{
						{
							Name:              "vw_foo",
							FailurePolicy:     &failurePolicyType,
							NamespaceSelector: &metav1.LabelSelector{},
							ClientConfig: ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Name:      "some-svc",
									Namespace: "k8s",
								},
							},
						},
					},
				},
			},
		},
	}
	return objs
}

func webhookURL() *kube.Objects {
	var url = "https://example.com/webhook/action"
	objs := initObjects(ar.Fail)
	objs.ValidatingWebhookConfigurations.Items[0].Webhooks[0].ClientConfig = ar.WebhookClientConfig{
		URL: &url,
	}
	objs.MutatingWebhookConfigurations.Items[0].Webhooks[0].ClientConfig = ar.WebhookClientConfig{
		URL: &url,
	}
	return objs
}

func label(label map[string]string) *kube.Objects {
	objs := initObjects(ar.Fail)
	objs.ValidatingWebhookConfigurations.Items[0].Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
		MatchLabels: label,
	}
	objs.MutatingWebhookConfigurations.Items[0].Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
		MatchLabels: label,
	}
	return objs
}

func expr(key string, values []string, labelOperator metav1.LabelSelectorOperator) *kube.Objects {
	objs := initObjects(ar.Fail)
	objs.ValidatingWebhookConfigurations.Items[0].Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      key,
				Operator: labelOperator,
				Values:   values,
			},
		},
	}
	objs.MutatingWebhookConfigurations.Items[0].Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      key,
				Operator: labelOperator,
				Values:   values,
			},
		},
	}
	return objs
}

func webhookErrors() []checks.Diagnostic {
	objs := initObjects(ar.Fail)
	validatingConfig := objs.ValidatingWebhookConfigurations.Items[0]
	mutatingConfig := objs.MutatingWebhookConfigurations.Items[0]
	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  "Webhook matches objects in the kube-system namespace. This can cause problems when upgrading the cluster.",
			Kind:     checks.ValidatingWebhookConfiguration,
			Object:   &validatingConfig.ObjectMeta,
			Owners:   validatingConfig.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Error,
			Message:  "Webhook matches objects in the kube-system namespace. This can cause problems when upgrading the cluster.",
			Kind:     checks.MutatingWebhookConfiguration,
			Object:   &mutatingConfig.ObjectMeta,
			Owners:   mutatingConfig.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
