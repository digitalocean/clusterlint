package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/digitalocean/clusterlint"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const all string = ""

type KubernetesAPI struct {
	Client kubernetes.Interface
}

func main() {
	api := &KubernetesAPI{Client: buildClient()}
	api.fetch()
}

func (k KubernetesAPI) fetch() *clusterlint.KubeObjects {
	client := k.Client.CoreV1()
	opts := metav1.ListOptions{}
	objects := &clusterlint.KubeObjects{}
	var err error

	objects.Nodes, err = client.Nodes().List(opts)
	handleError(err)

	objects.PersistentVolumes, err = client.PersistentVolumes().List(opts)
	handleError(err)

	objects.ComponentStatuses, err = client.ComponentStatuses().List(opts)
	handleError(err)

	objects.Pods, err = client.Pods(all).List(opts)
	handleError(err)

	objects.PodTemplates, err = client.PodTemplates(all).List(opts)
	handleError(err)

	objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(all).List(opts)
	handleError(err)

	objects.ConfigMaps, err = client.ConfigMaps(all).List(opts)
	handleError(err)

	objects.Secrets, err = client.Secrets(all).List(opts)
	handleError(err)

	objects.Services, err = client.Services(all).List(opts)
	handleError(err)

	objects.ServiceAccounts, err = client.ServiceAccounts(all).List(opts)
	handleError(err)

	objects.ResourceQuotas, err = client.ResourceQuotas(all).List(opts)
	handleError(err)

	objects.LimitRanges, err = client.LimitRanges(all).List(opts)
	handleError(err)

	return objects
}

func buildClient() kubernetes.Interface {
	k8sconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "absolute path to the kubeconfig file")
	context := flag.String("context", "", "context for the kubernetes client. default: current context")
	flag.Parse()

	var config *rest.Config
	if "" != *context {
		config, _ = buildConfigFromFlags(context, k8sconfig)
	} else {
		config, _ = clientcmd.BuildConfigFromFlags("", *k8sconfig)
	}

	client := kubernetes.NewForConfigOrDie(config)
	return client
}

func buildConfigFromFlags(context, kubeconfigPath *string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: *context,
		}).ClientConfig()
}

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
