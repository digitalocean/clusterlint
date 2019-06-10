package main

import (
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	var kubeconfig string = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	client := kubernetes.NewForConfigOrDie(config).CoreV1()
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

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
