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
