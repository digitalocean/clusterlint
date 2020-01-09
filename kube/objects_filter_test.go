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

package kube

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceError(t *testing.T) {
	_, err := NewObjectsFilter([]string{"kube-system"}, []string{"kube-system"})

	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("cannot specify both include and exclude namespace conditions"), err)
}

func TestFilter(t *testing.T) {
	filter, err := NewObjectsFilter([]string{"namespace_1"},nil)
	assert.NoError(t, err)
	objects := namespaceObjects()
	filter.Filter(objects)
	assert.Equal(t, namespace1Objects(), objects)

	filter, err = NewObjectsFilter(nil,[]string{"namespace_2"})
	assert.NoError(t, err)
	objects = namespaceObjects()
	filter.Filter(objects)
	assert.Equal(t, namespace1Objects(), objects)
}

func objects() *Objects {
	return &Objects{
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
}

func namespaceObjects() *Objects {
	objs := objects()
	objs.Pods = &corev1.PodList{Items: []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "pod_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "pod_2", Namespace: "namespace_2"}},
	}}
	objs.PodTemplates = &corev1.PodTemplateList{Items: []corev1.PodTemplate{
		{ObjectMeta: metav1.ObjectMeta{Name: "template_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "template_2", Namespace: "namespace_2"}},
	}}
	objs.PersistentVolumeClaims = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{
		{ObjectMeta: metav1.ObjectMeta{Name: "pvc_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "pvc_2", Namespace: "namespace_2"}},
	}}
	objs.ConfigMaps = &corev1.ConfigMapList{Items: []corev1.ConfigMap{
		{ObjectMeta: metav1.ObjectMeta{Name: "cm_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "cm_2", Namespace: "namespace_2"}},
	}}
	objs.Services = &corev1.ServiceList{Items: []corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "svc_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "svc_2", Namespace: "namespace_2"}},
	}}
	objs.Secrets = &corev1.SecretList{Items: []corev1.Secret{
		{ObjectMeta: metav1.ObjectMeta{Name: "secret_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "secret_2", Namespace: "namespace_2"}},
	}}
	objs.ServiceAccounts = &corev1.ServiceAccountList{Items: []corev1.ServiceAccount{
		{ObjectMeta: metav1.ObjectMeta{Name: "sa_1", Namespace: "namespace_1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "sa_2", Namespace: "namespace_2"}},
	}}
	return objs
}

func namespace1Objects() *Objects {
	objs := objects()
	objs.Pods = &corev1.PodList{Items: []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "pod_1", Namespace: "namespace_1"}},
	}}
	objs.PodTemplates = &corev1.PodTemplateList{Items: []corev1.PodTemplate{
		{ObjectMeta: metav1.ObjectMeta{Name: "template_1", Namespace: "namespace_1"}},
	}}
	objs.PersistentVolumeClaims = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{
		{ObjectMeta: metav1.ObjectMeta{Name: "pvc_1", Namespace: "namespace_1"}},
	}}
	objs.ConfigMaps = &corev1.ConfigMapList{Items: []corev1.ConfigMap{
		{ObjectMeta: metav1.ObjectMeta{Name: "cm_1", Namespace: "namespace_1"}},
	}}
	objs.Services = &corev1.ServiceList{Items: []corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "svc_1", Namespace: "namespace_1"}},
	}}
	objs.Secrets = &corev1.SecretList{Items: []corev1.Secret{
		{ObjectMeta: metav1.ObjectMeta{Name: "secret_1", Namespace: "namespace_1"}},
	}}
	objs.ServiceAccounts = &corev1.ServiceAccountList{Items: []corev1.ServiceAccount{
		{ObjectMeta: metav1.ObjectMeta{Name: "sa_1", Namespace: "namespace_1"}},
	}}
	return objs
}