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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceCheckMeta(t *testing.T) {
	defaultNamespaceCheck := defaultNamespaceCheck{}
	assert.Equal(t, "default-namespace", defaultNamespaceCheck.Name())
	assert.Equal(t, []string{"basic"}, defaultNamespaceCheck.Groups())
	assert.NotEmpty(t, defaultNamespaceCheck.Description())
}

func TestNamespaceCheckRegistration(t *testing.T) {
	defaultNamespaceCheck := &defaultNamespaceCheck{}
	check, err := checks.Get("default-namespace")
	assert.NoError(t, err)
	assert.Equal(t, check, defaultNamespaceCheck)
}

func TestNamespaceWarning(t *testing.T) {
	namespace := defaultNamespaceCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{"no objects in cluster", empty(), nil},
		{"user created objects in default namespace", userCreatedObjects(), errors(namespace)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := namespace.Run(context.Background(), test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func empty() *kube.Objects {
	objs := &kube.Objects{
		Pods:                   &corev1.PodList{},
		PodTemplates:           &corev1.PodTemplateList{},
		PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{},
		ConfigMaps:             &corev1.ConfigMapList{},
		Services:               &corev1.ServiceList{},
		Secrets:                &corev1.SecretList{},
		ServiceAccounts:        &corev1.ServiceAccountList{},
		ResourceQuotas:         &corev1.ResourceQuotaList{},
		LimitRanges:            &corev1.LimitRangeList{},
	}
	return objs
}

func userCreatedObjects() *kube.Objects {
	objs := empty()
	objs.Pods = &corev1.PodList{Items: []corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "default"}}}}
	objs.PodTemplates = &corev1.PodTemplateList{Items: []corev1.PodTemplate{{ObjectMeta: metav1.ObjectMeta{Name: "template_foo", Namespace: "default"}}}}
	objs.PersistentVolumeClaims = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc_foo", Namespace: "default"}}}}
	objs.ConfigMaps = &corev1.ConfigMapList{Items: []corev1.ConfigMap{{ObjectMeta: metav1.ObjectMeta{Name: "cm_foo", Namespace: "default"}}}}
	objs.Services = &corev1.ServiceList{Items: []corev1.Service{{ObjectMeta: metav1.ObjectMeta{Name: "svc_foo", Namespace: "default"}}}}
	objs.Secrets = &corev1.SecretList{Items: []corev1.Secret{{ObjectMeta: metav1.ObjectMeta{Name: "secret_foo", Namespace: "default"}}}}
	objs.ServiceAccounts = &corev1.ServiceAccountList{Items: []corev1.ServiceAccount{{ObjectMeta: metav1.ObjectMeta{Name: "sa_foo", Namespace: "default"}}}}
	return objs
}

func errors(n defaultNamespaceCheck) []checks.Diagnostic {
	objs := userCreatedObjects()
	pod := objs.Pods.Items[0]
	template := objs.PodTemplates.Items[0]
	pvc := objs.PersistentVolumeClaims.Items[0]
	cm := objs.ConfigMaps.Items[0]
	service := objs.Services.Items[0]
	secret := objs.Secrets.Items[0]
	sa := objs.ServiceAccounts.Items[0]
	d := []checks.Diagnostic{

		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.PodTemplate,
			Object:   &template.ObjectMeta,
			Owners:   template.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.PersistentVolumeClaim,
			Object:   &pvc.ObjectMeta,
			Owners:   pvc.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.ConfigMap,
			Object:   &cm.ObjectMeta,
			Owners:   cm.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Service,
			Object:   &service.ObjectMeta,
			Owners:   service.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Secret,
			Object:   &secret.ObjectMeta,
			Owners:   secret.ObjectMeta.GetOwnerReferences(),
		},
		{
			Check:    n.Name(),
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.ServiceAccount,
			Object:   &sa.ObjectMeta,
			Owners:   sa.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
