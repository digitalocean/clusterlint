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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBetaWebhookCheckMeta(t *testing.T) {
	betaWebhookCheck := betaWebhookCheck{}
	assert.Equal(t, "admission-controller-webhook-v1beta1", betaWebhookCheck.Name())
	assert.Equal(t, []string{"basic"}, betaWebhookCheck.Groups())
	assert.NotEmpty(t, betaWebhookCheck.Description())
}

func TestBetaWebhookCheckRegistration(t *testing.T) {
	betaWebhookCheck := &betaWebhookCheck{}
	check, err := checks.Get("admission-controller-webhook-v1beta1")
	assert.NoError(t, err)
	assert.Equal(t, check, betaWebhookCheck)
}

func TestBetaWebHookRun(t *testing.T) {
	emptyNamespaceList := &corev1.NamespaceList{
		Items: []corev1.Namespace{},
	}
	emptyServiceList := &corev1.ServiceList{
		Items: []corev1.Service{},
	}

	baseMWC := ar.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{Kind: "MutatingWebhookConfiguration", APIVersion: "v1beta1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mwc_foo",
		},
		Webhooks: []ar.MutatingWebhook{},
	}
	baseMW := ar.MutatingWebhook{
		Name: "mw_foo",
	}

	baseVWC := ar.ValidatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{Kind: "ValidatingWebhookConfiguration", APIVersion: "v1beta1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vwc_foo",
		},
		Webhooks: []ar.ValidatingWebhook{},
	}
	baseVW := ar.ValidatingWebhook{
		Name: "vw_foo",
	}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name: "no webhook configuration",
			objs: &kube.Objects{
				MutatingWebhookConfigurationsBeta:   &ar.MutatingWebhookConfigurationList{},
				ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{},
				SystemNamespace:                     &corev1.Namespace{},
			},
			expected: nil,
		},
		{
			name: "direct url webhooks",
			objs: &kube.Objects{
				Namespaces: emptyNamespaceList,
				MutatingWebhookConfigurationsBeta: &ar.MutatingWebhookConfigurationList{
					Items: []ar.MutatingWebhookConfiguration{
						func() ar.MutatingWebhookConfiguration {
							mwc := baseMWC
							mw := baseMW
							mw.ClientConfig = ar.WebhookClientConfig{
								URL: strPtr("http://webhook.com"),
							}
							mwc.Webhooks = append(mwc.Webhooks, mw)
							return mwc
						}(),
					},
				},
				ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{
					Items: []ar.ValidatingWebhookConfiguration{
						func() ar.ValidatingWebhookConfiguration {
							vwc := baseVWC
							vw := baseVW
							vw.ClientConfig = ar.WebhookClientConfig{
								URL: strPtr("http://webhook.com"),
							}
							vwc.Webhooks = append(vwc.Webhooks, vw)
							return vwc
						}(),
					},
				},
				SystemNamespace: &corev1.Namespace{},
			},
			expected: nil,
		},
		{
			name: "namespace does not exist",
			objs: &kube.Objects{
				Namespaces: emptyNamespaceList,
				Services: &corev1.ServiceList{
					Items: []corev1.Service{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "service",
							},
						},
					},
				},
				MutatingWebhookConfigurationsBeta: &ar.MutatingWebhookConfigurationList{
					Items: []ar.MutatingWebhookConfiguration{
						func() ar.MutatingWebhookConfiguration {
							mwc := baseMWC
							mw := baseMW
							mw.ClientConfig = ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Namespace: "missing",
									Name:      "service",
								},
							}
							mwc.Webhooks = append(mwc.Webhooks, mw)
							return mwc
						}(),
					},
				},
				ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{
					Items: []ar.ValidatingWebhookConfiguration{
						func() ar.ValidatingWebhookConfiguration {
							vwc := baseVWC
							vw := baseVW
							vw.ClientConfig = ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Namespace: "missing",
									Name:      "service",
								},
							}
							vwc.Webhooks = append(vwc.Webhooks, vw)
							return vwc
						}(),
					},
				},
				SystemNamespace: &corev1.Namespace{},
			},
			expected: []checks.Diagnostic{
				{
					Severity: checks.Error,
					Message:  "Validating webhook vw_foo is configured against a service in a namespace that does not exist.",
					Kind:     checks.ValidatingWebhookConfiguration,
				},
				{
					Severity: checks.Error,
					Message:  "Mutating webhook mw_foo is configured against a service in a namespace that does not exist.",
					Kind:     checks.MutatingWebhookConfiguration,
				},
			},
		},
		{
			name: "service does not exist",
			objs: &kube.Objects{
				Namespaces: &corev1.NamespaceList{
					Items: []corev1.Namespace{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "webhook",
							},
						},
					},
				},
				Services: emptyServiceList,
				MutatingWebhookConfigurationsBeta: &ar.MutatingWebhookConfigurationList{
					Items: []ar.MutatingWebhookConfiguration{
						func() ar.MutatingWebhookConfiguration {
							mwc := baseMWC
							mw := baseMW
							mw.ClientConfig = ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Namespace: "webhook",
									Name:      "service",
								},
							}
							mwc.Webhooks = append(mwc.Webhooks, mw)
							return mwc
						}(),
					},
				},
				ValidatingWebhookConfigurationsBeta: &ar.ValidatingWebhookConfigurationList{
					Items: []ar.ValidatingWebhookConfiguration{
						func() ar.ValidatingWebhookConfiguration {
							vwc := baseVWC
							vw := baseVW
							vw.ClientConfig = ar.WebhookClientConfig{
								Service: &ar.ServiceReference{
									Namespace: "webhook",
									Name:      "service",
								},
							}
							vwc.Webhooks = append(vwc.Webhooks, vw)
							return vwc
						}(),
					},
				},
				SystemNamespace: &corev1.Namespace{},
			},
			expected: []checks.Diagnostic{
				{
					Severity: checks.Error,
					Message:  "Validating webhook vw_foo is configured against a service that does not exist.",
					Kind:     checks.ValidatingWebhookConfiguration,
				},
				{
					Severity: checks.Error,
					Message:  "Mutating webhook mw_foo is configured against a service that does not exist.",
					Kind:     checks.MutatingWebhookConfiguration,
				},
			},
		},
	}

	betaWebhookCheck := betaWebhookCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics, err := betaWebhookCheck.Run(test.objs)
			assert.NoError(t, err)

			// skip checking object and owner since for this checker it just uses the object being checked.
			var strippedDiagnostics []checks.Diagnostic
			for _, d := range diagnostics {
				d.Object = nil
				d.Owners = nil
				strippedDiagnostics = append(strippedDiagnostics, d)
			}

			assert.ElementsMatch(t, test.expected, strippedDiagnostics)
		})
	}
}
