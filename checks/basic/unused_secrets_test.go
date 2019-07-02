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
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnusedSecretCheckMeta(t *testing.T) {
	unusedSecretCheck := unusedSecretCheck{}
	assert.Equal(t, "unused-secret", unusedSecretCheck.Name())
	assert.Equal(t, []string{"basic"}, unusedSecretCheck.Groups())
	assert.NotEmpty(t, unusedSecretCheck.Description())
}

func TestUnusedSecretCheckRegistration(t *testing.T) {
	unusedSecretCheck := &unusedSecretCheck{}
	check, err := checks.Get("unused-secret")
	assert.NoError(t, err)
	assert.Equal(t, check, unusedSecretCheck)
}

func TestUnusedSecretWarning(t *testing.T) {
	unusedSecretCheck := unusedSecretCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no secrets",
			objs:     &kube.Objects{Pods: &corev1.PodList{}, Secrets: &corev1.SecretList{}},
			expected: nil,
		},
		{
			name:     "secret volume",
			objs:     secretVolume(),
			expected: nil,
		},
		{
			name:     "environment variable references secret",
			objs:     secretEnvSource(),
			expected: nil,
		},
		{
			name:     "pod with image pull secrets",
			objs:     imagePullSecrets(),
			expected: nil,
		},
		{
			name:     "projected volume references secret",
			objs:     secretProjection(),
			expected: nil,
		},
		{
			name: "unused secret",
			objs: initSecret(),
			expected: []checks.Diagnostic{
				{
					Check:    unusedSecretCheck.Name(),
					Severity: checks.Warning,
					Message:  "Unused secret",
					Kind:     checks.Secret,
					Object:   &metav1.ObjectMeta{Name: "secret_foo", Namespace: "k8s"},
					Owners:   GetOwners(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := unusedSecretCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initSecret() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				},
			},
		},
		Secrets: &corev1.SecretList{
			Items: []corev1.Secret{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "secret_foo", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func secretVolume() *kube.Objects {
	objs := initSecret()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "bar",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "secret_foo",
					},
				},
			}},
	}
	return objs
}

func secretProjection() *kube.Objects {
	objs := initSecret()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "bar",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "secret_foo",
									},
								},
							},
						},
					},
				},
			}},
	}
	return objs
}

func secretEnvSource() *kube.Objects {
	objs := initSecret()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "test-container",
				Image: "docker.io/nginx",
				EnvFrom: []corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret_foo"},
						},
					},
				},
			}},
	}
	return objs
}

func imagePullSecrets() *kube.Objects {
	objs := initSecret()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "secret_foo",
			},
		},
	}
	return objs
}
