/*
Copyright 2020 DigitalOcean

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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const cmNamespace = "k8s"

func TestUnusedConfigMapCheckMeta(t *testing.T) {
	unusedCMCheck := unusedCMCheck{}
	assert.Equal(t, "unused-config-map", unusedCMCheck.Name())
	assert.Equal(t, []string{"basic"}, unusedCMCheck.Groups())
	assert.NotEmpty(t, unusedCMCheck.Description())
}

func TestUnusedConfigMapCheckRegistration(t *testing.T) {
	unusedCMCheck := &unusedCMCheck{}
	check, err := checks.Get("unused-config-map")
	assert.NoError(t, err)
	assert.Equal(t, check, unusedCMCheck)
}

func TestUnusedConfigMapWarning(t *testing.T) {
	unusedCMCheck := unusedCMCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no config maps",
			objs:     &kube.Objects{Nodes: &corev1.NodeList{}, Pods: &corev1.PodList{}, ConfigMaps: &corev1.ConfigMapList{}},
			expected: nil,
		},
		{
			name:     "volume mounted config map",
			objs:     configMapVolume(),
			expected: nil,
		},
		{
			name:     "environment variable references config map",
			objs:     configMapEnvSource(),
			expected: nil,
		},
		{
			name:     "environment variable value from references config map",
			objs:     configMapEnvVarValueFromSource(),
			expected: nil,
		},
		{
			name:     "projected volume references config map",
			objs:     projectedVolume(),
			expected: nil,
		},
		{
			name:     "node config source references config map",
			objs:     nodeConfigSource(),
			expected: nil,
		},
		{
			name: "unused config map",
			objs: initConfigMap(),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Unused config map",
					Kind:     checks.ConfigMap,
					Object:   &metav1.ObjectMeta{Name: "cm_foo", Namespace: cmNamespace},
					Owners:   GetOwners(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := unusedCMCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initConfigMap() *kube.Objects {
	objs := &kube.Objects{
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Node", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "node_foo"},
				},
			},
		},
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: cmNamespace},
				},
			},
		},
		ConfigMaps: &corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "cm_foo", Namespace: cmNamespace},
				},
			},
		},
	}
	return objs
}

func nodeConfigSource() *kube.Objects {
	objs := initConfigMap()
	objs.Nodes.Items[0].Spec = corev1.NodeSpec{
		ConfigSource: &corev1.NodeConfigSource{
			ConfigMap: &corev1.ConfigMapNodeConfigSource{
				Name:      "cm_foo",
				Namespace: cmNamespace,
			},
		},
	}
	return objs
}

func configMapVolume() *kube.Objects {
	objs := initConfigMap()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "bar",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "cm_foo"},
					},
				},
			}},
	}
	return objs
}

func projectedVolume() *kube.Objects {
	objs := initConfigMap()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "bar",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{Name: "cm_foo"},
								},
							},
						},
					},
				},
			}},
	}
	return objs
}

func configMapEnvSource() *kube.Objects {
	objs := initConfigMap()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "test-container",
				Image: "docker.io/nginx",
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "cm_foo"},
						},
					},
				},
			}},
	}
	return objs
}

func configMapEnvVarValueFromSource() *kube.Objects {
	objs := initConfigMap()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "test-container",
				Image: "docker.io/nginx",
				Env: []corev1.EnvVar{
					{
						Name: "special_env_var",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "cm_foo"},
							},
						},
					},
				},
			},
		},
	}
	return objs
}
