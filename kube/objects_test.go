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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchObjects(t *testing.T) {
	api := &Client{
		KubeClient: fake.NewSimpleClientset(),
	}

	api.KubeClient.CoreV1().Namespaces().Create(&corev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kube-system",
			Labels: map[string]string{"doks_key": "bar"}},
	})

	actual, err := api.FetchObjects(context.Background())
	assert.NoError(t, err)

	assert.NotNil(t, actual.Nodes)
	assert.NotNil(t, actual.PersistentVolumes)
	assert.NotNil(t, actual.ComponentStatuses)
	assert.NotNil(t, actual.Pods)
	assert.NotNil(t, actual.PodTemplates)
	assert.NotNil(t, actual.PersistentVolumeClaims)
	assert.NotNil(t, actual.ConfigMaps)
	assert.NotNil(t, actual.Services)
	assert.NotNil(t, actual.Secrets)
	assert.NotNil(t, actual.ServiceAccounts)
	assert.NotNil(t, actual.ResourceQuotas)
	assert.NotNil(t, actual.LimitRanges)
	assert.NotNil(t, actual.ValidatingWebhookConfigurations)
	assert.NotNil(t, actual.MutatingWebhookConfigurations)
	assert.NotNil(t, actual.SystemNamespace)
}

func TestNewClientErrors(t *testing.T) {
	// test both yaml and filepath specified
	_, err := NewClient(WithConfigFile("some-path"), WithYaml([]byte("yaml")))
	assert.Equal(t, errors.New("cannot specify both yaml and kubeconfg file path"), err)
	// test no authentication mechanism
	_, err = NewClient()
	assert.Equal(t, errors.New("cannot authenticate Kubernetes API requests"), err)
}
