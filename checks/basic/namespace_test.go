package basic

import (
	"fmt"
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceCheckMeta(t *testing.T) {
	defaultNamespaceCheck := defaultNamespaceCheck{}
	assert.Equal(t, "default-namespace", defaultNamespaceCheck.Name())
	assert.Equal(t, "Checks if there are any user created k8s objects in the default namespace.", defaultNamespaceCheck.Description())
	assert.Equal(t, []string{"basic"}, defaultNamespaceCheck.Groups())
}

func TestNamespaceCheckRegistration(t *testing.T) {
	defaultNamespaceCheck := &defaultNamespaceCheck{}
	check, err := checks.Get("default-namespace")
	assert.Equal(t, check, defaultNamespaceCheck)
	assert.Nil(t, err)
}

func TestNamespaceWarning(t *testing.T) {
	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []error
	}{
		{"no objects in cluster", empty(), nil},
		{"user created objects in default namespace", userCreatedObjects(), errors()},
	}

	namespace := defaultNamespaceCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			w, e, err := namespace.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, w)
			assert.Empty(t, e)
		})
	}
}

func empty() *kube.Objects {
	objs := &kube.Objects{
		Pods:                   &corev1.PodList{},
		PodTemplates:           &corev1.PodTemplateList{},
		PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{},
		ConfigMaps:             &corev1.ConfigMapList{},
		Services:               &corev1.ServiceList{},
		Secrets:                &corev1.SecretList{},
		ServiceAccounts:        &corev1.ServiceAccountList{},
		ResourceQuotas:         &corev1.ResourceQuotaList{},
		LimitRanges:            &corev1.LimitRangeList{},
	}
	return objs
}

func userCreatedObjects() *kube.Objects {
	objs := empty()
	objs.Pods = &corev1.PodList{Items: []corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "default"}}}}
	objs.PodTemplates = &corev1.PodTemplateList{Items: []corev1.PodTemplate{{ObjectMeta: metav1.ObjectMeta{Name: "template_foo", Namespace: "default"}}}}
	objs.PersistentVolumeClaims = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc_foo", Namespace: "default"}}}}
	objs.ConfigMaps = &corev1.ConfigMapList{Items: []corev1.ConfigMap{{ObjectMeta: metav1.ObjectMeta{Name: "cm_foo", Namespace: "default"}}}}
	objs.Services = &corev1.ServiceList{Items: []corev1.Service{{ObjectMeta: metav1.ObjectMeta{Name: "svc_foo", Namespace: "default"}}}}
	objs.Secrets = &corev1.SecretList{Items: []corev1.Secret{{ObjectMeta: metav1.ObjectMeta{Name: "secret_foo", Namespace: "default"}}}}
	objs.ServiceAccounts = &corev1.ServiceAccountList{Items: []corev1.ServiceAccount{{ObjectMeta: metav1.ObjectMeta{Name: "sa_foo", Namespace: "default"}}}}
	return objs
}

func errors() []error {
	w := []error{
		fmt.Errorf("Pod 'pod_foo' is in the default namespace."),
		fmt.Errorf("Pod template 'template_foo' is in the default namespace."),
		fmt.Errorf("Persistent Volume Claim 'pvc_foo' is in the default namespace."),
		fmt.Errorf("Config Map 'cm_foo' is in the default namespace."),
		fmt.Errorf("Service 'svc_foo' is in the default namespace."),
		fmt.Errorf("Secret 'secret_foo' is in the default namespace."),
		fmt.Errorf("Service Account 'sa_foo' is in the default namespace."),
	}
	return w
}
