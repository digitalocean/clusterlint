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
	arv1 "k8s.io/api/admissionregistration/v1"
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBetaWebhookCheckMeta(t *testing.T) {
	webhookCheck := betaWebhookReplacementCheck{}
	assert.Equal(t, "admission-controller-webhook-replacement-v1beta1", webhookCheck.Name())
	assert.Equal(t, []string{"doks"}, webhookCheck.Groups())
	assert.NotEmpty(t, webhookCheck.Description())
}

func TestBetaWebhookCheckRegistration(t *testing.T) {
	webhookCheck := &betaWebhookReplacementCheck{}
	check, err := checks.Get("admission-controller-webhook-replacement-v1beta1")
	assert.NoError(t, err)
	assert.Equal(t, check, webhookCheck)
}

func TestBetaWebhookSkipWhenV1Exists(t *testing.T) {
	v1Objs := webhookTestObjects(
		arv1.Fail,
		&metav1.LabelSelector{},
		arv1.WebhookClientConfig{
			Service: &arv1.ServiceReference{
				Namespace: "webhook",
				Name:      "webhook-service",
			},
		},
		2,
		[]string{"*"},
		[]string{"*"},
	)
	betaObjs := webhookTestObjectsBeta(
		ar.Fail,
		&metav1.LabelSelector{},
		ar.WebhookClientConfig{
			Service: &ar.ServiceReference{
				Namespace: "webhook",
				Name:      "webhook-service",
			},
		},
		2,
		[]string{"*"},
		[]string{"*"},
	)

	objs := &kube.Objects{
		MutatingWebhookConfigurations:       v1Objs.MutatingWebhookConfigurations,
		ValidatingWebhookConfigurations:     v1Objs.ValidatingWebhookConfigurations,
		MutatingWebhookConfigurationsBeta:   betaObjs.MutatingWebhookConfigurationsBeta,
		ValidatingWebhookConfigurationsBeta: betaObjs.ValidatingWebhookConfigurationsBeta,
	}

	webhookCheck := betaWebhookReplacementCheck{}
	d, err := webhookCheck.Run(objs)
	assert.NoError(t, err)
	assert.Empty(t, d)
}

func TestBetaWebhookError(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name: "no webhook configurations",
			objs: &kube.Objects{
				MutatingWebhookConfigurationsBeta:   &ar.MutatingWebhookConfigurationList{},
				ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{},
				MutatingWebhookConfigurations:       &arv1.MutatingWebhookConfigurationList{},
				ValidatingWebhookConfigurations:     &arv1.ValidatingWebhookConfigurationList{},
				SystemNamespace:                     &corev1.Namespace{},
			},
			expected: nil,
		},
		{
			name: "failure policy is ignore",
			objs: webhookTestObjectsBeta(
				ar.Ignore,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "webook does not use service",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					URL: &webhookURL,
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "webook service is apiserver",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "default",
						Name:      "kubernetes",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace label selector does not match kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"non-existent-label-on-namespace": "bar"},
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace OpExists expression selector does not match kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchExpressions: expr("non-existent", []string{}, metav1.LabelSelectorOpExists),
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace OpDoesNotExist expression selector does not match kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchExpressions: expr("doks_key", []string{}, metav1.LabelSelectorOpDoesNotExist),
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace OpIn expression selector does not match kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchExpressions: expr("doks_key", []string{"non-existent"}, metav1.LabelSelectorOpIn),
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace OpNotIn expression selector does not match kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchExpressions: expr("doks_key", []string{"bar"}, metav1.LabelSelectorOpNotIn),
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "namespace label selector does not match own namespace",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"doks_key": "bar"},
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: nil,
		},
		{
			name: "single-node cluster",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"doks_key": "bar"},
				},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				1,
				[]string{"*"},
				[]string{"*"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "webhook applies to its own namespace and kube-system",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "error: webhook applies to core/v1 group",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{""},
				[]string{"v1"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "error: webhook applies to apps/v1 group",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"apps"},
				[]string{"v1"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "error: webhook applies to apps/v1beta1 group",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"apps"},
				[]string{"v1beta1"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "error: webhook applies to apps/v1beta2 group",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"apps"},
				[]string{"v1beta2"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "error: webhook applies to *",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"*"},
				[]string{"*"},
			),
			expected: webhookErrorsBeta(),
		},
		{
			name: "no error: webhook applies to batch/v1",
			objs: webhookTestObjectsBeta(
				ar.Fail,
				&metav1.LabelSelector{},
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				2,
				[]string{"batch"},
				[]string{"*"},
			),
			expected: nil,
		},
	}

	webhookCheck := betaWebhookReplacementCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := webhookCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func webhookTestObjectsBeta(
	failurePolicyType ar.FailurePolicyType,
	nsSelector *metav1.LabelSelector,
	clientConfig ar.WebhookClientConfig,
	numNodes int,
	groups []string,
	versions []string,
) *kube.Objects {
	objs := &kube.Objects{
		SystemNamespace: &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kube-system",
				Labels: map[string]string{"doks_key": "bar"},
			},
		},
		Namespaces: &corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "kube-system",
						Labels: map[string]string{"doks_key": "bar"},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "webhook",
						Labels: map[string]string{"doks_key": "xyzzy"},
					},
				},
			},
		},
		MutatingWebhookConfigurations:   &arv1.MutatingWebhookConfigurationList{},
		ValidatingWebhookConfigurations: &arv1.ValidatingWebhookConfigurationList{},
		MutatingWebhookConfigurationsBeta: &ar.MutatingWebhookConfigurationList{
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
							NamespaceSelector: nsSelector,
							ClientConfig:      clientConfig,
							Rules: []ar.RuleWithOperations{
								{
									Rule: ar.Rule{
										APIGroups:   groups,
										APIVersions: versions,
									},
								},
							},
						},
					},
				},
			},
		},
		ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{
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
							NamespaceSelector: nsSelector,
							ClientConfig:      clientConfig,
							Rules: []ar.RuleWithOperations{
								{
									Rule: ar.Rule{
										APIGroups:   groups,
										APIVersions: versions,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	objs.Nodes = &corev1.NodeList{}
	for i := 0; i < numNodes; i++ {
		objs.Nodes.Items = append(objs.Nodes.Items, corev1.Node{})
	}
	return objs
}

func webhookErrorsBeta() []checks.Diagnostic {
	objs := webhookTestObjectsBeta(ar.Fail, nil, ar.WebhookClientConfig{}, 0, []string{"*"}, []string{"*"})
	validatingConfig := objs.ValidatingWebhookConfigurationsBeta.Items[0]
	mutatingConfig := objs.MutatingWebhookConfigurationsBeta.Items[0]

	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  "Validating webhook is configured in such a way that it may be problematic during upgrades.",
			Kind:     checks.ValidatingWebhookConfiguration,
			Object:   &validatingConfig.ObjectMeta,
			Owners:   validatingConfig.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Error,
			Message:  "Mutating webhook is configured in such a way that it may be problematic during upgrades.",
			Kind:     checks.MutatingWebhookConfiguration,
			Object:   &mutatingConfig.ObjectMeta,
			Owners:   mutatingConfig.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
