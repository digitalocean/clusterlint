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

	"golang.org/x/sync/errgroup"
	ar "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//Identifier is used to identify a specific namspace scoped object.
type Identifier struct {
	Name      string
	Namespace string
}

// Objects encapsulates all the objects from a Kubernetes cluster.
type Objects struct {
	Nodes                           *corev1.NodeList
	PersistentVolumes               *corev1.PersistentVolumeList
	ComponentStatuses               *corev1.ComponentStatusList
	SystemNamespace                 *corev1.Namespace
	Pods                            *corev1.PodList
	PodTemplates                    *corev1.PodTemplateList
	PersistentVolumeClaims          *corev1.PersistentVolumeClaimList
	ConfigMaps                      *corev1.ConfigMapList
	Services                        *corev1.ServiceList
	Secrets                         *corev1.SecretList
	ServiceAccounts                 *corev1.ServiceAccountList
	ResourceQuotas                  *corev1.ResourceQuotaList
	LimitRanges                     *corev1.LimitRangeList
	MutatingWebhookConfigurations   *ar.MutatingWebhookConfigurationList
	ValidatingWebhookConfigurations *ar.ValidatingWebhookConfigurationList
}

// Client encapsulates a client for a Kubernetes cluster.
type Client struct {
	KubeClient kubernetes.Interface
}

// FetchObjects returns the objects from a Kubernetes cluster.
// ctx is currently unused during API calls. More info: https://github.com/kubernetes/community/pull/1166
func (c *Client) FetchObjects(ctx context.Context) (*Objects, error) {
	client := c.KubeClient.CoreV1()
	admissionControllerClient := c.KubeClient.AdmissionregistrationV1beta1()
	opts := metav1.ListOptions{}
	objects := &Objects{}

	var g errgroup.Group

	g.Go(func() (err error) {
		objects.Nodes, err = client.Nodes().List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.PersistentVolumes, err = client.PersistentVolumes().List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.ComponentStatuses, err = client.ComponentStatuses().List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.Pods, err = client.Pods(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.PodTemplates, err = client.PodTemplates(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.ConfigMaps, err = client.ConfigMaps(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.Secrets, err = client.Secrets(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.Services, err = client.Services(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.ServiceAccounts, err = client.ServiceAccounts(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.ResourceQuotas, err = client.ResourceQuotas(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.LimitRanges, err = client.LimitRanges(corev1.NamespaceAll).List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.SystemNamespace, err = client.Namespaces().Get(metav1.NamespaceSystem, metav1.GetOptions{})
		return
	})
	g.Go(func() (err error) {
		objects.MutatingWebhookConfigurations, err = admissionControllerClient.MutatingWebhookConfigurations().List(opts)
		return
	})
	g.Go(func() (err error) {
		objects.ValidatingWebhookConfigurations, err = admissionControllerClient.ValidatingWebhookConfigurations().List(opts)
		return
	})

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return objects, nil
}

// NewClient builds a kubernetes client to interact with the live cluster.
// The kube config file path or the kubeconfig yaml must be specified
// If not specified, defaults are assumed - configPath: ~/.kube/config, configContext: current context
func NewClient(opts ...Option) (*Client, error) {
	defOpts := &options{}

	for _, opt := range opts {
		if err := opt(defOpts); err != nil {
			return nil, err
		}
	}

	var config *rest.Config
	var err error
	err = defOpts.validate()
	if err != nil {
		return nil, err
	}

	if defOpts.yaml != nil {
		config, err = clientcmd.RESTConfigFromKubeConfig(defOpts.yaml)
	} else if len(defOpts.paths) != 0 {
		loadingRules := &clientcmd.ClientConfigLoadingRules{Precedence: defOpts.paths}
		configOverrides := &clientcmd.ConfigOverrides{}
		if defOpts.kubeContext != "" {
			configOverrides.CurrentContext = defOpts.kubeContext
		}
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	} else {
		err = errors.New("cannot authenticate Kubernetes API requests")
	}

	if err != nil {
		return nil, err
	}
	config.Timeout = defOpts.timeout
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		KubeClient: client,
	}, nil
}
