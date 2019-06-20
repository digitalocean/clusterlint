package basic

import (
	"fmt"

	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func initPod() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func container(image string) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "bar",
				Image: image,
			}},
	}
	return objs
}

func initContainer(image string) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:  "bar",
				Image: image,
			}},
	}
	return objs
}

func issues(s string) []error {
	issue := []error{
		fmt.Errorf(s),
	}
	return issue
}
