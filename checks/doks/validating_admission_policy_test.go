/*
Copyright 2026 DigitalOcean

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
	ar "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidatingAdmissionPolicyCheckMeta(t *testing.T) {
	c := validatingAdmissionPolicyCheck{}
	assert.Equal(t, "validating-admission-policy", c.Name())
	assert.Equal(t, []string{"doks"}, c.Groups())
	assert.NotEmpty(t, c.Description())
}

func TestValidatingAdmissionPolicyCheckRegistration(t *testing.T) {
	c := &validatingAdmissionPolicyCheck{}
	check, err := checks.Get("validating-admission-policy")
	assert.NoError(t, err)
	assert.Equal(t, check, c)
}

func TestValidatingAdmissionPolicyError(t *testing.T) {
	fail := ar.Fail
	ignore := ar.Ignore

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name: "no policies",
			objs: &kube.Objects{
				SystemNamespace:                   &corev1.Namespace{},
				ValidatingAdmissionPolicies:       &ar.ValidatingAdmissionPolicyList{},
				ValidatingAdmissionPolicyBindings: &ar.ValidatingAdmissionPolicyBindingList{},
			},
			expected: nil,
		},
		{
			name: "policy with deny binding applies to core/v1 in kube-system",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{},
				[]string{""},
				[]string{"v1"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: vapErrors(),
		},
		{
			name: "policy applies to apps/v1",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{},
				[]string{"apps"},
				[]string{"v1"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: vapErrors(),
		},
		{
			name: "failurePolicy unset defaults to Fail",
			objs: vapTestObjects(
				nil,
				&metav1.LabelSelector{},
				[]string{"*"},
				[]string{"*"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: vapErrors(),
		},
		{
			name: "failurePolicy is Ignore still flagged",
			objs: vapTestObjects(
				&ignore,
				&metav1.LabelSelector{},
				[]string{"*"},
				[]string{"*"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: vapErrors(),
		},
		{
			name: "binding does not enforce deny",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{},
				[]string{"*"},
				[]string{"*"},
				[]ar.ValidationAction{ar.Warn, ar.Audit},
				nil,
			),
			expected: nil,
		},
		{
			name: "policy has no binding",
			objs: func() *kube.Objects {
				objs := vapTestObjects(&fail, &metav1.LabelSelector{}, []string{"*"}, []string{"*"}, []ar.ValidationAction{ar.Deny}, nil)
				objs.ValidatingAdmissionPolicyBindings = &ar.ValidatingAdmissionPolicyBindingList{}
				return objs
			}(),
			expected: nil,
		},
		{
			name: "policy does not apply to core/apps resources",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{},
				[]string{"batch"},
				[]string{"v1"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: nil,
		},
		{
			name: "policy namespace selector does not match kube-system",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"non-existent-label": "bar"},
				},
				[]string{"*"},
				[]string{"*"},
				[]ar.ValidationAction{ar.Deny},
				nil,
			),
			expected: nil,
		},
		{
			name: "binding namespace selector excludes kube-system",
			objs: vapTestObjects(
				&fail,
				&metav1.LabelSelector{},
				[]string{"*"},
				[]string{"*"},
				[]ar.ValidationAction{ar.Deny},
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"non-existent-label": "bar"},
				},
			),
			expected: nil,
		},
	}

	check := validatingAdmissionPolicyCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := check.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func vapTestObjects(
	failurePolicy *ar.FailurePolicyType,
	nsSelector *metav1.LabelSelector,
	groups []string,
	versions []string,
	actions []ar.ValidationAction,
	bindingNSSelector *metav1.LabelSelector,
) *kube.Objects {
	var matchResources *ar.MatchResources
	if bindingNSSelector != nil {
		matchResources = &ar.MatchResources{NamespaceSelector: bindingNSSelector}
	}

	return &kube.Objects{
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
			},
		},
		ValidatingAdmissionPolicies: &ar.ValidatingAdmissionPolicyList{
			Items: []ar.ValidatingAdmissionPolicy{
				{
					TypeMeta: metav1.TypeMeta{Kind: "ValidatingAdmissionPolicy", APIVersion: "admissionregistration.k8s.io/v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "vap_foo",
					},
					Spec: ar.ValidatingAdmissionPolicySpec{
						FailurePolicy: failurePolicy,
						MatchConstraints: &ar.MatchResources{
							NamespaceSelector: nsSelector,
							ResourceRules: []ar.NamedRuleWithOperations{
								{
									RuleWithOperations: ar.RuleWithOperations{
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
		},
		ValidatingAdmissionPolicyBindings: &ar.ValidatingAdmissionPolicyBindingList{
			Items: []ar.ValidatingAdmissionPolicyBinding{
				{
					TypeMeta: metav1.TypeMeta{Kind: "ValidatingAdmissionPolicyBinding", APIVersion: "admissionregistration.k8s.io/v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "vap_foo_binding",
					},
					Spec: ar.ValidatingAdmissionPolicyBindingSpec{
						PolicyName:        "vap_foo",
						ValidationActions: actions,
						MatchResources:    matchResources,
					},
				},
			},
		},
	}
}

func vapErrors() []checks.Diagnostic {
	objs := vapTestObjects(nil, &metav1.LabelSelector{}, []string{"*"}, []string{"*"}, []ar.ValidationAction{ar.Deny}, nil)
	policy := objs.ValidatingAdmissionPolicies.Items[0]

	return []checks.Diagnostic{
		{
			Severity: checks.Error,
			Message:  "Validating admission policy is configured in such a way that it may be problematic during upgrades.",
			Kind:     checks.ValidatingAdmissionPolicy,
			Object:   &policy.ObjectMeta,
			Owners:   policy.ObjectMeta.GetOwnerReferences(),
		},
	}
}
