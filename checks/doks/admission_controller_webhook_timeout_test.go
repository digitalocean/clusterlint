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
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWebhookTimeoutCheckMeta(t *testing.T) {
	webhookCheck := webhookTimeoutCheck{}
	assert.Equal(t, "admission-controller-webhook-timeout", webhookCheck.Name())
	assert.Equal(t, []string{"doks"}, webhookCheck.Groups())
	assert.NotEmpty(t, webhookCheck.Description())
}

func TestWebhookTimeoutRegistration(t *testing.T) {
	webhookCheck := &webhookTimeoutCheck{}
	check, err := checks.Get("admission-controller-webhook-timeout")
	assert.NoError(t, err)
	assert.Equal(t, check, webhookCheck)
}

func TestWebhookTimeoutError(t *testing.T) {
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
			},
			expected: nil,
		},
		{
			name: "TimeoutSeconds value is set to 10 seconds",
			objs: webhookTimeoutTestObjects(
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
			objs: webhookTimeoutTestObjects(
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
			objs: webhookTimeoutTestObjects(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(30),
				2,
			),
			expected: webhookTimeoutErrors(),
		},
		{
			name: "TimeoutSeconds value is set to 31 seconds",
			objs: webhookTimeoutTestObjects(
				ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: "webhook",
						Name:      "webhook-service",
					},
				},
				toIntP(31),
				2,
			),
			expected: webhookTimeoutErrors(),
		},
		{
			name: "TimeoutSeconds value is set to nil",
			objs: webhookTimeoutTestObjects(
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

	webhookCheck := webhookTimeoutCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := webhookCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func webhookTimeoutTestObjects(
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
		MutatingWebhookConfigurations: &ar.MutatingWebhookConfigurationList{
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
		ValidatingWebhookConfigurations: &ar.ValidatingWebhookConfigurationList{
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

func webhookTimeoutErrors() []checks.Diagnostic {
	objs := webhookTimeoutTestObjects(ar.WebhookClientConfig{}, nil, 0)
	validatingConfig := objs.ValidatingWebhookConfigurations.Items[0]
	mutatingConfig := objs.MutatingWebhookConfigurations.Items[0]

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

// converts an int to an int32 and returns a pointer
func toIntP(i int) *int32 {
	num := int32(i)
	return &num
}
