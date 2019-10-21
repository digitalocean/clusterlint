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
	"context"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	checks.Register(&defaultNamespaceCheck{})
}

type defaultNamespaceCheck struct{}

type alert struct {
	diagnostics []checks.Diagnostic
	mu          sync.Mutex
}

// GetWarnings returns alert.warnings
func (alert *alert) GetDiagnostics() []checks.Diagnostic {
	return alert.diagnostics
}

// SetWarnings sets alert.warnings
func (alert *alert) SetDiagnostics(d []checks.Diagnostic) {
	alert.diagnostics = d
}

// warn adds warnings for k8s objects that should not be in the default namespace
func (alert *alert) warn(k8stype checks.Kind, itemMeta metav1.ObjectMeta) {
	d := checks.Diagnostic{
		Check:    "default-namespace",
		Severity: checks.Warning,
		Message:  "Avoid using the default namespace",
		Kind:     k8stype,
		Object:   &itemMeta,
		Owners:   itemMeta.GetOwnerReferences(),
	}
	alert.mu.Lock()
	alert.diagnostics = append(alert.diagnostics, d)
	alert.mu.Unlock()
}

// Name returns a unique name for this check.
func (nc *defaultNamespaceCheck) Name() string {
	return "default-namespace"
}

// Groups returns a list of group names this check should be part of.
func (nc *defaultNamespaceCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *defaultNamespaceCheck) Description() string {
	return "Checks if there are any user created k8s objects in the default namespace."
}

// checkPods checks if there are pods in the default namespace
func (nc *defaultNamespaceCheck) checkPods(items *corev1.PodList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn(checks.Pod, item.ObjectMeta)
		}
	}
}

// checkPodTemplates checks if there are pod templates in the default namespace
func (nc *defaultNamespaceCheck) checkPodTemplates(items *corev1.PodTemplateList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn(checks.PodTemplate, item.ObjectMeta)
		}
	}
}

// checkPVCs checks if there are pvcs in the default namespace
func (nc *defaultNamespaceCheck) checkPVCs(items *corev1.PersistentVolumeClaimList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn(checks.PersistentVolumeClaim, item.ObjectMeta)
		}
	}
}

// checkConfigMaps checks if there are config maps in the default namespace
func (nc *defaultNamespaceCheck) checkConfigMaps(items *corev1.ConfigMapList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn(checks.ConfigMap, item.ObjectMeta)
		}
	}
}

// checkServices checks if there are user created services in the default namespace
func (nc *defaultNamespaceCheck) checkServices(items *corev1.ServiceList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && item.GetName() != "kubernetes" {
			alert.warn(checks.Service, item.ObjectMeta)
		}
	}
}

// checkSecrets checks if there are user created secrets in the default namespace
func (nc *defaultNamespaceCheck) checkSecrets(items *corev1.SecretList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && item.Type != corev1.SecretTypeServiceAccountToken {
			alert.warn(checks.Secret, item.ObjectMeta)
		}
	}
}

// checkSA checks if there are user created SAs in the default namespace
func (nc *defaultNamespaceCheck) checkSA(items *corev1.ServiceAccountList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && item.GetName() != "default" {
			alert.warn(checks.ServiceAccount, item.ObjectMeta)
		}
	}
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *defaultNamespaceCheck) Run(_ context.Context, objects *kube.Objects) ([]checks.Diagnostic, error) {
	alert := &alert{}
	var g errgroup.Group
	g.Go(func() error {
		nc.checkPods(objects.Pods, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkPodTemplates(objects.PodTemplates, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkPVCs(objects.PersistentVolumeClaims, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkConfigMaps(objects.ConfigMaps, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkServices(objects.Services, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkSecrets(objects.Secrets, alert)
		return nil
	})

	g.Go(func() error {
		nc.checkSA(objects.ServiceAccounts, alert)
		return nil
	})

	err := g.Wait()
	return alert.GetDiagnostics(), err
}
