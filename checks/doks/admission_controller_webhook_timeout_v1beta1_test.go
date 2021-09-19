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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	arv1 "k8s.io/api/admissionregistration/v1"
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBetaWebhookTimeoutCheckMeta(t *testing.T) {
	webhookCheck := betaWebhookTimeoutCheck{}
	assert.Equal(t, "admission-controller-webhook-timeout-v1beta1", webhookCheck.Name())
	assert.Equal(t, []string{"doks"}, webhookCheck.Groups())
	assert.NotEmpty(t, webhookCheck.Description())
}

func TestBetaWebhookTimeoutRegistration(t *testing.T) {
	webhookCheck := &betaWebhookTimeoutCheck{}
	check, err := checks.Get("admission-controller-webhook-timeout-v1beta1")
	assert.NoError(t, err)
	assert.Equal(t, check, webhookCheck)
}

func TestBetaWebhookTimeoutSkipWhenV1Exists(t *testing.T) {
	v1Objs := webhookTimeoutTestObjects(
		arv1.WebhookClientConfig{
			Service: &arv1.ServiceReference{
				Namespace: "webhook",
				Name:      "webhook-service",
			},
		},
		toIntP(31),
		2,
	)
	betaObjs := webhookTimeoutTestObjectsBeta(
		ar.WebhookClientConfig{
			Service: &ar.ServiceReference{
				Namespace: "webhook",
				Name:      "webhook-service",
			},
		},
		toIntP(31),
		2,
	)

	objs := &kube.Objects{
		MutatingWebhookConfigurations:       v1Objs.MutatingWebhookConfigurations,
		ValidatingWebhookConfigurations:     v1Objs.ValidatingWebhookConfigurations,
		MutatingWebhookConfigurationsBeta:   betaObjs.MutatingWebhookConfigurationsBeta,
		ValidatingWebhookConfigurationsBeta: betaObjs.ValidatingWebhookConfigurationsBeta,
	}

	webhookCheck := betaWebhookTimeoutCheck{}
	d, err := webhookCheck.Run(objs)
	assert.NoError(t, err)
	assert.Empty(t, d)
}

func TestBetaWebhookTimeoutError(t *testing.T) {
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
			},
			expected: nil,
		},
		{
			name: "TimeoutSeconds value is set to 10 seconds",
			objs: webhookTimeoutTestObjectsBeta(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(10),
				2,
			),
			expected: nil,
		},
		{
			name: "TimeoutSeconds value is set to 29 seconds",
			objs: webhookTimeoutTestObjectsBeta(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(29),
				2,
			),
			expected: nil,
		},
		{
			name: "TimeoutSeconds value is set to 30 seconds",
			objs: webhookTimeoutTestObjectsBeta(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(30),
				2,
			),
			expected: webhookTimeoutErrorsBeta(),
		},
		{
			name: "TimeoutSeconds value is set to 31 seconds",
			objs: webhookTimeoutTestObjectsBeta(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(31),
				2,
			),
			expected: webhookTimeoutErrorsBeta(),
		},
		{
			name: "TimeoutSeconds value is set to nil",
			objs: webhookTimeoutTestObjectsBeta(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				nil,
				2,
			),
			expected: nil,
		},
	}

	webhookCheck := betaWebhookTimeoutCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := webhookCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func webhookTimeoutTestObjectsBeta(
	clientConfig ar.WebhookClientConfig,
	timeoutSeconds *int32,
	numNodes int,
) *kube.Objects {
	objs := &kube.Objects{
		SystemNamespace: &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kube-system",
				Labels: map[string]string{"doks_test_key": "bar"},
			},
		},
		Namespaces: &corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "kube-system",
						Labels: map[string]string{"doks_test_key": "bar"},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "webhook",
						Labels: map[string]string{"doks_test_key": "xyzzy"},
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
							Name:           "mw_foo",
							ClientConfig:   clientConfig,
							TimeoutSeconds: timeoutSeconds,
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
							Name:           "vw_foo",
							ClientConfig:   clientConfig,
							TimeoutSeconds: timeoutSeconds,
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

func webhookTimeoutErrorsBeta() []checks.Diagnostic {
	objs := webhookTimeoutTestObjectsBeta(ar.WebhookClientConfig{}, nil, 0)
	validatingConfig := objs.ValidatingWebhookConfigurationsBeta.Items[0]
	mutatingConfig := objs.MutatingWebhookConfigurationsBeta.Items[0]

	diagnostics := []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  "Validating webhook with a TimeoutSeconds value greater than 29 seconds will block upgrades.",
			Kind:     checks.ValidatingWebhookConfiguration,
			Object:   &validatingConfig.ObjectMeta,
			Owners:   validatingConfig.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Error,
			Message:  "Mutating webhook with a TimeoutSeconds value greater than 29 seconds will block upgrades.",
			Kind:     checks.MutatingWebhookConfiguration,
			Object:   &mutatingConfig.ObjectMeta,
			Owners:   mutatingConfig.ObjectMeta.GetOwnerReferences(),
		},
	}
	return diagnostics
}
