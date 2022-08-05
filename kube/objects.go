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
	"fmt"
	"net/http"

	csitypes "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	csitypesbeta "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1beta1"
	csi "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	"golang.org/x/sync/errgroup"
	arv1 "k8s.io/api/admissionregistration/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	st "k8s.io/api/storage/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Load client-go authentication plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
	VolumeSnapshotsV1               *csitypes.VolumeSnapshotList
	VolumeSnapshotsV1Content        *csitypes.VolumeSnapshotContentList
	VolumeSnapshotsBeta             *csitypesbeta.VolumeSnapshotList
	VolumeSnapshotsBetaContent      *csitypesbeta.VolumeSnapshotContentList
	StorageClasses                  *st.StorageClassList
	DefaultStorageClass             *st.StorageClass
	MutatingWebhookConfigurations   *arv1.MutatingWebhookConfigurationList
	ValidatingWebhookConfigurations *arv1.ValidatingWebhookConfigurationList
	Namespaces                      *corev1.NamespaceList
	CronJobs                        *batchv1beta1.CronJobList
}

// Client encapsulates a client for a Kubernetes cluster.
type Client struct {
	KubeClient kubernetes.Interface
	CSIClient  csi.Interface
	httpClient *http.Client
}

func (c *Client) Close() {
	if c.httpClient != nil {
		utilnet.CloseIdleConnectionsFor(c.httpClient.Transport)
	}
}

// FetchObjects returns the objects from a Kubernetes cluster.
// ctx is currently unused during API calls. More info: https://github.com/kubernetes/community/pull/1166
func (c *Client) FetchObjects(ctx context.Context, filter ObjectFilter) (*Objects, error) {
	client := c.KubeClient.CoreV1()
	admissionControllerClient := c.KubeClient.AdmissionregistrationV1()
	batchClient := c.KubeClient.BatchV1beta1()
	storageClient := c.KubeClient.StorageV1()
	csiClient := c.CSIClient.SnapshotV1()
	csiBetaClient := c.CSIClient.SnapshotV1beta1()
	opts := metav1.ListOptions{}
	objects := &Objects{}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		objects.Nodes, err = client.Nodes().List(gCtx, opts)
		return
	})
	g.Go(func() (err error) {
		objects.StorageClasses, err = storageClient.StorageClasses().List(gCtx, opts)
		if err != nil {
			return err
		}
		for _, s := range objects.StorageClasses.Items {
			if v, _ := s.Annotations["storageclass.kubernetes.io/is-default-class"]; v == "true" {
				defaultStorageClass := s
				objects.DefaultStorageClass = &defaultStorageClass
			}
		}
		return
	})
	g.Go(func() (err error) {
		objects.PersistentVolumes, err = client.PersistentVolumes().List(gCtx, opts)
		err = annotateFetchError("PersistentVolumes", err)
		return
	})
	g.Go(func() (err error) {
		objects.Pods, err = client.Pods(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("Pods", err)
		return
	})
	g.Go(func() (err error) {
		objects.PodTemplates, err = client.PodTemplates(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("PodTemplates", err)
		return
	})
	g.Go(func() (err error) {
		objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("PersistentVolumeClaims", err)
		return
	})
	g.Go(func() (err error) {
		objects.ConfigMaps, err = client.ConfigMaps(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("ConfigMaps", err)
		return
	})
	g.Go(func() (err error) {
		objects.Secrets, err = client.Secrets(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("Secrets", err)
		return
	})
	g.Go(func() (err error) {
		objects.Services, err = client.Services(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("Services", err)
		return
	})
	g.Go(func() (err error) {
		objects.ServiceAccounts, err = client.ServiceAccounts(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("ServiceAccounts", err)
		return
	})
	g.Go(func() (err error) {
		objects.ResourceQuotas, err = client.ResourceQuotas(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("ResourceQuotas", err)
		return
	})
	g.Go(func() (err error) {
		objects.LimitRanges, err = client.LimitRanges(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("LimitRanges", err)
		return
	})
	g.Go(func() (err error) {
		objects.SystemNamespace, err = client.Namespaces().Get(gCtx, metav1.NamespaceSystem, metav1.GetOptions{})
		if err != nil {
			err = fmt.Errorf("failed to fetch namespace %q: %s", metav1.NamespaceSystem, err)
		}
		return
	})
	g.Go(func() (err error) {
		objects.MutatingWebhookConfigurations, err = admissionControllerClient.MutatingWebhookConfigurations().List(gCtx, opts)
		err = annotateFetchError("MutatingWebhookConfigurations (v1)", err)
		return
	})
	g.Go(func() (err error) {
		objects.ValidatingWebhookConfigurations, err = admissionControllerClient.ValidatingWebhookConfigurations().List(gCtx, opts)
		err = annotateFetchError("ValidatingWebhookConfigurations (v1)", err)
		return
	})
	g.Go(func() (err error) {
		objects.Namespaces, err = client.Namespaces().List(gCtx, opts)
		err = annotateFetchError("Namespaces", err)
		return
	})
	g.Go(func() (err error) {
		objects.CronJobs, err = batchClient.CronJobs(corev1.NamespaceAll).List(gCtx, filter.NamespaceOptions(opts))
		err = annotateFetchError("CronJobs", err)
		return
	})
	g.Go(func() (err error) {
		objects.VolumeSnapshotsV1, err = csiClient.VolumeSnapshots(corev1.NamespaceAll).List(ctx, filter.NamespaceOptions(opts))
		err = annotateFetchError("VolumeSnapshotsV1", err)
		return
	})
	g.Go(func() (err error) {
		objects.VolumeSnapshotsV1Content, err = csiClient.VolumeSnapshotContents().List(ctx, filter.NamespaceOptions(opts))
		err = annotateFetchError("VolumeSnapshotsV1Contents", err)
		return
	})
	g.Go(func() (err error) {
		objects.VolumeSnapshotsBeta, err = csiBetaClient.VolumeSnapshots(corev1.NamespaceAll).List(ctx, filter.NamespaceOptions(opts))
		err = annotateFetchError("VolumeSnapshotsBeta", err)
		return
	})
	g.Go(func() (err error) {
		objects.VolumeSnapshotsBetaContent, err = csiBetaClient.VolumeSnapshotContents().List(ctx, filter.NamespaceOptions(opts))
		err = annotateFetchError("VolumeSnapshotsBetaContents", err)
		return
	})
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return objectsWithoutNils(objects), nil
}

func annotateFetchError(kind string, err error) error {
	if err == nil {
		return nil
	}
	if kerrors.IsNotFound(err) {
		// Resource doesn't exist in this cluster's version, so there aren't any
		// objects to list and check.
		return nil
	}

	return fmt.Errorf("failed to fetch %s: %s", kind, err)
}

func objectsWithoutNils(objects *Objects) *Objects {
	if objects.Nodes == nil {
		objects.Nodes = &v1.NodeList{}
	}
	if objects.PersistentVolumes == nil {
		objects.PersistentVolumes = &v1.PersistentVolumeList{}
	}
	if objects.Pods == nil {
		objects.Pods = &v1.PodList{}
	}
	if objects.PodTemplates == nil {
		objects.PodTemplates = &v1.PodTemplateList{}
	}
	if objects.PersistentVolumeClaims == nil {
		objects.PersistentVolumeClaims = &v1.PersistentVolumeClaimList{}
	}
	if objects.ConfigMaps == nil {
		objects.ConfigMaps = &v1.ConfigMapList{}
	}
	if objects.Services == nil {
		objects.Services = &v1.ServiceList{}
	}
	if objects.Secrets == nil {
		objects.Secrets = &v1.SecretList{}
	}
	if objects.ServiceAccounts == nil {
		objects.ServiceAccounts = &v1.ServiceAccountList{}
	}
	if objects.ResourceQuotas == nil {
		objects.ResourceQuotas = &v1.ResourceQuotaList{}
	}
	if objects.LimitRanges == nil {
		objects.LimitRanges = &v1.LimitRangeList{}
	}
	if objects.StorageClasses == nil {
		objects.StorageClasses = &st.StorageClassList{}
	}
	if objects.MutatingWebhookConfigurations == nil {
		objects.MutatingWebhookConfigurations = &arv1.MutatingWebhookConfigurationList{}
	}
	if objects.ValidatingWebhookConfigurations == nil {
		objects.ValidatingWebhookConfigurations = &arv1.ValidatingWebhookConfigurationList{}
	}
	if objects.Namespaces == nil {
		objects.Namespaces = &v1.NamespaceList{}
	}
	if objects.CronJobs == nil {
		objects.CronJobs = &batchv1beta1.CronJobList{}
	}
	if objects.VolumeSnapshotsV1 == nil {
		objects.VolumeSnapshotsV1 = &csitypes.VolumeSnapshotList{}
	}
	if objects.VolumeSnapshotsV1Content == nil {
		objects.VolumeSnapshotsV1Content = &csitypes.VolumeSnapshotContentList{}
	}
	if objects.VolumeSnapshotsBeta == nil {
		objects.VolumeSnapshotsBeta = &csitypesbeta.VolumeSnapshotList{}
	}
	if objects.VolumeSnapshotsBetaContent == nil {
		objects.VolumeSnapshotsBetaContent = &csitypesbeta.VolumeSnapshotContentList{}
	}
	return objects
}

func NewClientConfigFromKubeConfig(opts *options) (*rest.Config, error) {
	var err error
	config := &rest.Config{}

	if opts.yaml != nil {
		config, err = clientcmd.RESTConfigFromKubeConfig(opts.yaml)
	} else {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		if len(opts.paths) != 0 {
			loadingRules.Precedence = opts.paths
		}
		configOverrides := &clientcmd.ConfigOverrides{}
		if opts.kubeContext != "" {
			configOverrides.CurrentContext = opts.kubeContext
		}
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	}

	if err != nil {
		return nil, err
	}

	return config, nil
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

	if defOpts.inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = NewClientConfigFromKubeConfig(defOpts)
	}
	if err != nil {
		return nil, err
	}

	config.Timeout = defOpts.timeout
	if defOpts.transportWrapper != nil {
		config.Wrap(defOpts.transportWrapper)
	}

	// disable transport caching
	// see: https://github.com/kubernetes/kubernetes/issues/109289
	config.Proxy = http.ProxyFromEnvironment
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfigAndClient(config, httpClient)
	if err != nil {
		return nil, err
	}
	csiClient, err := csi.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		KubeClient: client,
		CSIClient:  csiClient,
		httpClient: httpClient,
	}, nil
}
