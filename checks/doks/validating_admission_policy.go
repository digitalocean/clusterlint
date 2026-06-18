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
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	ar "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&validatingAdmissionPolicyCheck{})
}

type validatingAdmissionPolicyCheck struct{}

func (c *validatingAdmissionPolicyCheck) Name() string {
	return "validating-admission-policy"
}

func (c *validatingAdmissionPolicyCheck) Groups() []string {
	return []string{"doks"}
}

func (c *validatingAdmissionPolicyCheck) Description() string {
	return "Check for validating admission policies that could cause problems during upgrades or node replacement"
}

func (c *validatingAdmissionPolicyCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	for _, policy := range objects.ValidatingAdmissionPolicies.Items {
		policy := policy

		if policy.Spec.MatchConstraints == nil {
			// A policy with no match constraints does not match anything.
			continue
		}
		if !applicableNamedRules(policy.Spec.MatchConstraints.ResourceRules) {
			// Policies that do not apply to core/v1, apps/v1, apps/v1beta1,
			// apps/v1beta2 resources are fine.
			continue
		}
		if !policyDeniesInNamespace(&policy, objects.ValidatingAdmissionPolicyBindings, objects.SystemNamespace) {
			// To block reconciliation a policy needs a Deny-enforcing binding
			// whose scope reaches kube-system, where the DOKS reconciler's upgrade
			//  operations target system components.
			continue
		}

		diagnostics = append(diagnostics, checks.Diagnostic{
			Severity: checks.Error,
			Message:  "Validating admission policy is configured in such a way that it may be problematic during upgrades.",
			Kind:     checks.ValidatingAdmissionPolicy,
			Object:   &policy.ObjectMeta,
			Owners:   policy.ObjectMeta.GetOwnerReferences(),
		})
	}

	return diagnostics, nil
}

// policyDeniesInNamespace reports whether the policy could deny requests in the
// given namespace. A ValidatingAdmissionPolicy on its own never rejects
// requests. Enforcement is configured by its bindings' validationActions.
func policyDeniesInNamespace(policy *ar.ValidatingAdmissionPolicy, bindings *ar.ValidatingAdmissionPolicyBindingList, namespace *corev1.Namespace) bool {
	if namespace == nil {
		return false
	}
	if policy.Spec.MatchConstraints.NamespaceSelector != nil &&
		!selectorMatchesNamespace(policy.Spec.MatchConstraints.NamespaceSelector, namespace) {
		return false
	}
	for _, binding := range bindings.Items {
		if binding.Spec.PolicyName != policy.Name {
			continue
		}
		if !enforcesDeny(binding.Spec.ValidationActions) {
			continue
		}
		if binding.Spec.MatchResources == nil ||
			binding.Spec.MatchResources.NamespaceSelector == nil ||
			selectorMatchesNamespace(binding.Spec.MatchResources.NamespaceSelector, namespace) {
			return true
		}
	}
	return false
}

// enforcesDeny reports whether the given validationActions include Deny.
func enforcesDeny(actions []ar.ValidationAction) bool {
	for _, action := range actions {
		if action == ar.Deny {
			return true
		}
	}
	return false
}

// applicableNamedRules reports whether any of the policy's resource rules apply
// to the core/v1, apps/v1, apps/v1beta1 or apps/v1beta2 API groups, mirroring
// the logic used by the admission-controller-webhook-replacement check.
func applicableNamedRules(rules []ar.NamedRuleWithOperations) bool {
	plain := make([]ar.RuleWithOperations, 0, len(rules))
	for _, r := range rules {
		plain = append(plain, r.RuleWithOperations)
	}
	return applicable(plain)
}
