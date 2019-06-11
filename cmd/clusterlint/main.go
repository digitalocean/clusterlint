package main

import (
	"flag"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const all string = ""

type KubeObjects struct {
	Nodes                  *corev1.NodeList
	PersistentVolumes      *corev1.PersistentVolumeList
	ComponentStatuses      *corev1.ComponentStatusList
	Pods                   *corev1.PodList
	PodTemplates           *corev1.PodTemplateList
	PersistentVolumeClaims *corev1.PersistentVolumeClaimList
	ConfigMaps             *corev1.ConfigMapList
	Services               *corev1.ServiceList
	Secrets                *corev1.SecretList
	ServiceAccounts        *corev1.ServiceAccountList
	ResourceQuotas         *corev1.ResourceQuotaList
	LimitRanges            *corev1.LimitRangeList
}

func main() {
	client := buildClient()
	opts := metav1.ListOptions{}
	Objects := KubeObjects{}
	var err error

	Objects.Nodes, err = client.Nodes().List(opts)
	handleError(err)

	Objects.PersistentVolumes, err = client.PersistentVolumes().List(opts)
	handleError(err)

	Objects.ComponentStatuses, err = client.ComponentStatuses().List(opts)
	handleError(err)

	Objects.Pods, err = client.Pods(all).List(opts)
	handleError(err)

	Objects.PodTemplates, err = client.PodTemplates(all).List(opts)
	handleError(err)

	Objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(all).List(opts)
	handleError(err)

	Objects.ConfigMaps, err = client.ConfigMaps(all).List(opts)
	handleError(err)

	Objects.Secrets, err = client.Secrets(all).List(opts)
	handleError(err)

	Objects.Services, err = client.Services(all).List(opts)
	handleError(err)

	Objects.ServiceAccounts, err = client.ServiceAccounts(all).List(opts)
	handleError(err)

	Objects.ResourceQuotas, err = client.ResourceQuotas(all).List(opts)
	handleError(err)

	Objects.LimitRanges, err = client.LimitRanges(all).List(opts)
	handleError(err)
}

func buildClient() v1.CoreV1Interface {
	k8sconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "absolute path to the kubeconfig file")
	flag.Parse()

	config, _ := clientcmd.BuildConfigFromFlags("", *k8sconfig)
	client := kubernetes.NewForConfigOrDie(config).CoreV1()
	return client
}

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
