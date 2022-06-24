/*
Copyright 2022 DigitalOcean

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
	"fmt"
	"net/http"
	"testing"

	csi "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestFetchObjects(t *testing.T) {
	tests := []struct {
		name        string
		fakeMutator func(cs *fake.Clientset)
	}{
		{
			name: "happy path",
		},
		{
			name: "resources not found",
			fakeMutator: func(cs *fake.Clientset) {
				notFoundReactionFunc := func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, &kerrors.StatusError{
						ErrStatus: metav1.Status{
							Reason:  metav1.StatusReasonNotFound,
							Message: fmt.Sprintf("%s not found", action.GetResource().Resource),
						},
					}
				}
				cs.PrependReactor("list", "mutatingwebhookconfigurations", notFoundReactionFunc)
				cs.PrependReactor("list", "validatingwebhookconfigurations", notFoundReactionFunc)
			},
		},
	}

	for _, test := range tests {
		cs := fake.NewSimpleClientset()
		csifake := csi.NewSimpleClientset()
		if test.fakeMutator != nil {
			test.fakeMutator(cs)
		}

		api := &Client{
			KubeClient: cs,
			CSIClient:  csifake,
		}

		api.KubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kube-system",
				Labels: map[string]string{"doks_key": "bar"}},
		}, metav1.CreateOptions{})

		actual, err := api.FetchObjects(context.Background(), ObjectFilter{})
		assert.NoError(t, err)

		assert.NotNil(t, actual.Nodes)
		assert.NotNil(t, actual.PersistentVolumes)
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
		assert.NotNil(t, actual.CronJobs)
		assert.NotNil(t, actual.VolumeSnapshotsV1)
		assert.NotNil(t, actual.VolumeSnapshotsBeta)
	}

}

func TestNewClientErrors(t *testing.T) {
	t.Run("both yaml and filepath specified", func(t *testing.T) {
		_, err := NewClient(WithConfigFile("some-path"), WithYaml([]byte("yaml")))
		assert.Equal(t, errors.New("cannot specify yaml and kubeconfig file paths"), err)
	})

	t.Run("both yaml and KUBECONFIG specified", func(t *testing.T) {
		_, err := NewClient(WithMergedConfigFiles([]string{"some-path"}), WithYaml([]byte("yaml")))
		assert.Equal(t, errors.New("cannot specify yaml and kubeconfig file paths"), err)
	})

	t.Run("both yaml and in-cluster specified", func(t *testing.T) {
		_, err := NewClient(WithYaml([]byte("yaml")), InCluster())
		assert.Equal(t, errors.New("cannot specify yaml or kubeconfig file paths when running in-cluster mode"), err)
	})

	t.Run("both KUBECONFIG and in-cluster specified", func(t *testing.T) {
		_, err := NewClient(WithMergedConfigFiles([]string{"some-path"}), InCluster())
		assert.Equal(t, errors.New("cannot specify yaml or kubeconfig file paths when running in-cluster mode"), err)
	})

	t.Run("in-cluster access enabled", func(t *testing.T) {
		_, err := NewClient(InCluster())
		assert.Equal(t, errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"), err)
	})
}

type failTransport struct{}

func (failTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("fail")
}

func TestNewClientRoundTripper(t *testing.T) {
	client, err := NewClient(WithTransportWrapper(func(_ http.RoundTripper) http.RoundTripper {
		return failTransport{}
	}), WithYaml([]byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: http://localhost
  name: cool
contexts:
- context:
    cluster: cool
    user: admin
  name: cool
current-context: cool
users:
- name: admin
`)))
	assert.NoError(t, err)
	_, err = client.KubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}, metav1.CreateOptions{})
	assert.Contains(t, err.Error(), "fail")
}

func TestAnnotateFetchError(t *testing.T) {
	kindName := "kind"

	tests := []struct {
		name    string
		inErr   error
		wantErr error
	}{
		{
			name:    "no error",
			inErr:   nil,
			wantErr: nil,
		},
		{
			name: "not found error",
			inErr: &kerrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonNotFound,
				},
			},
			wantErr: nil,
		},
		{
			name:    "other error",
			inErr:   errors.New("other error"),
			wantErr: fmt.Errorf("failed to fetch %s: other error", kindName),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.wantErr, annotateFetchError(kindName, test.inErr))
	}
}
