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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchObjects(t *testing.T) {
	api := &Client{
		kubeClient: fake.NewSimpleClientset(),
	}

	actual, err := api.FetchObjects()
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
}
