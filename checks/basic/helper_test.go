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

package basic

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func initPod() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func initMultiplePods() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_1", Namespace: "k8s"},
				},
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_2", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func GetObjectMeta() *metav1.ObjectMeta {
	objs := initPod()
	return &objs.Pods.Items[0].ObjectMeta
}

func GetOwners() []metav1.OwnerReference {
	objs := initPod()
	return objs.Pods.Items[0].ObjectMeta.GetOwnerReferences()
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

func issues(severity checks.Severity, message string, kind checks.Kind, check string) []checks.Diagnostic {
	d := []checks.Diagnostic{
		{
			Severity: severity,
			Message:  message,
			Kind:     kind,
			Object:   GetObjectMeta(),
			Owners:   GetOwners(),
		},
	}
	return d
}
