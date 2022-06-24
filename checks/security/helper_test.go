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

package security

import (
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

func containerPrivileged(privileged bool) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
			}},
	}
	return objs
}

func containerNonRoot(pod, container *bool) *kube.Objects {
	objs := initPod()
	podSecurityContext := &corev1.PodSecurityContext{}
	if pod != nil {
		podSecurityContext = &corev1.PodSecurityContext{RunAsNonRoot: pod}
	}
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		SecurityContext: podSecurityContext,
		Containers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{RunAsNonRoot: container},
			}},
	}
	return objs
}

func initContainerNonRoot(pod, container *bool) *kube.Objects {
	objs := initPod()
	podSecurityContext := &corev1.PodSecurityContext{}
	if pod != nil {
		podSecurityContext = &corev1.PodSecurityContext{RunAsNonRoot: pod}
	}
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		SecurityContext: podSecurityContext,
		InitContainers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{RunAsNonRoot: container},
			}},
	}
	return objs
}

func containerSecurityContextNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "bar",
			}},
	}
	return objs
}

func containerPrivilegedNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{},
			}},
	}
	return objs
}

func initContainerPrivileged(privileged bool) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
			}},
	}
	return objs
}

func initContainerSecurityContextNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "bar",
			}},
	}
	return objs
}

func initContainerPrivilegedNil() *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:            "bar",
				SecurityContext: &corev1.SecurityContext{},
			}},
	}
	return objs
}
