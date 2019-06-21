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
		expected []checks.Diagnostic
	}{
		{"no objects in cluster", empty(), nil},
		{"user created objects in default namespace", userCreatedObjects(), errors()},
	}

	namespace := defaultNamespaceCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d, err := namespace.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
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
			Message:  "Avoid using the default namespace for pod 'pod_foo'",
			Object:   kube.Object{TypeInfo: &pod.TypeMeta, ObjectInfo: &pod.ObjectMeta},
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for pod template 'template_foo'",
			Object:   kube.Object{TypeInfo: &template.TypeMeta, ObjectInfo: &template.ObjectMeta},
			Owners:   template.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for persistent volume claim 'pvc_foo'",
			Object:   kube.Object{TypeInfo: &pvc.TypeMeta, ObjectInfo: &pvc.ObjectMeta},
			Owners:   pvc.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for config map 'cm_foo'",
			Object:   kube.Object{TypeInfo: &cm.TypeMeta, ObjectInfo: &cm.ObjectMeta},
			Owners:   cm.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for service 'svc_foo'",
			Object:   kube.Object{TypeInfo: &service.TypeMeta, ObjectInfo: &service.ObjectMeta},
			Owners:   service.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for secret 'secret_foo'",
			Object:   kube.Object{TypeInfo: &secret.TypeMeta, ObjectInfo: &secret.ObjectMeta},
			Owners:   secret.ObjectMeta.GetOwnerReferences(),
		},
		{
			Severity: checks.Warning,
			Message:  "Avoid using the default namespace for service account 'sa_foo'",
			Object:   kube.Object{TypeInfo: &sa.TypeMeta, ObjectInfo: &sa.ObjectMeta},
			Owners:   sa.ObjectMeta.GetOwnerReferences(),
		},
	}
	return d
}
