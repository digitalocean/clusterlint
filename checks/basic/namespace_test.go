package basic

import (
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
	assert.Equal(t, []string{"basic"}, defaultNamespaceCheck.Groups())
	assert.NotEmpty(t, defaultNamespaceCheck.Description())
}

func TestNamespaceCheckRegistration(t *testing.T) {
	defaultNamespaceCheck := &defaultNamespaceCheck{}
	check, err := checks.Get("default-namespace")
	assert.NoError(t, err)
	assert.Equal(t, check, defaultNamespaceCheck)
}

func TestNamespaceWarning(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{"no objects in cluster", empty(), nil},
		{"user created objects in default namespace", userCreatedObjects(), errors()},
	}

	namespace := defaultNamespaceCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := namespace.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
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

func errors() []checks.Diagnostic {
	objs := userCreatedObjects()
	pod := objs.Pods.Items[0]
	template := objs.PodTemplates.Items[0]
	pvc := objs.PersistentVolumeClaims.Items[0]
	cm := objs.ConfigMaps.Items[0]
	service := objs.Services.Items[0]
	secret := objs.Secrets.Items[0]
	sa := objs.ServiceAccounts.Items[0]
	d := []checks.Diagnostic{

		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.PodTemplate,
			Object:   &template.ObjectMeta,
			Owners:   template.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.PersistentVolumeClaim,
			Object:   &pvc.ObjectMeta,
			Owners:   pvc.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.ConfigMap,
			Object:   &cm.ObjectMeta,
			Owners:   cm.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Service,
			Object:   &service.ObjectMeta,
			Owners:   service.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.Secret,
			Object:   &secret.ObjectMeta,
			Owners:   secret.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace",
			Kind:     checks.ServiceAccount,
			Object:   &sa.ObjectMeta,
			Owners:   sa.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
