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
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no config maps",
			objs:     &kube.Objects{Pods: &corev1.PodList{}, ConfigMaps: &corev1.ConfigMapList{}},
			expected: nil,
		},
		{
			name:     "volume mounted config map",
			objs:     configMapVolume(),
			expected: nil,
		},
		{
			name:     "environment variable references to config map",
			objs:     configMapEnvSource(),
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

	unusedCMCheck := unusedCMCheck{}

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
