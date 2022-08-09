/*
Copyright 2021 DigitalOcean

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

package doks

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	st "k8s.io/api/storage/v1"
)

func TestDobsPodOwnerCheckMeta(t *testing.T) {
	dobsPodOwner := dobsPodOwner{}
	assert.Equal(t, "dobs-pod-owner", dobsPodOwner.Name())
	assert.Equal(t, []string{"doks"}, dobsPodOwner.Groups())
	assert.NotEmpty(t, dobsPodOwner.Description())
}

func TestDobsPodOwnerCheckRegistration(t *testing.T) {
	dobsPodOwner := &dobsPodOwner{}
	check, err := checks.Get("dobs-pod-owner")
	assert.NoError(t, err)
	assert.Equal(t, check, dobsPodOwner)
}

func TestDobsPodOwnerWarning(t *testing.T) {
	dobsPodOwner := dobsPodOwner{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pods",
			objs:     &kube.Objects{Pods: &corev1.PodList{}},
			expected: nil,
		},
		{
			name:     "no pods referencing dobs volumes",
			objs:     noDobs(),
			expected: nil,
		},
		{
			name: "bare dobs pod referenced by pvc",
			objs: pvcDobs(DOBlockStorageName, DOCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by deprecated pvc annotation",
			objs: provideStorageClassInAnnotation(pvcDobs(DOBlockStorageName, DOCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by pvc -- with legacy driver",
			objs: pvcDobs(DOBlockStorageName, LegacyCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by deprecated pvc annotation -- with legacy driver",
			objs: provideStorageClassInAnnotation(pvcDobs(DOBlockStorageName, LegacyCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by pvc with default storage class",
			objs: pvcDobs("", DOCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by deprecated pvc with default storage class",
			objs: provideStorageClassInAnnotation(pvcDobs("", DOCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by pvc with default storage class -- with legacy driver",
			objs: pvcDobs("", LegacyCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by deprecated pvc annotation with default storage class -- with legacy driver",
			objs:   provideStorageClassInAnnotation(pvcDobs("", LegacyCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by csi",
			objs: csiDobs(DOCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "bare dobs pod referenced by legacy csi driver",
			objs: csiDobs(LegacyCSIDriver),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object:   &metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Owners:   nil,
				},
			},
		},
		{
			name: "dobs pod owned by deployment",
			objs: deployment(pvcDobs("", DOCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "foo",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app",
						},
					},
				},
			},
		},
		{
			name: "dobs pod owned by deployment, with deprecated pvc annotation",
			objs: deployment(provideStorageClassInAnnotation(pvcDobs("", DOCSIDriver))),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "foo",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app",
						},
					},
				},
			},
		},
		{
			name: "dobs pods owned by multiple deployments",
			objs: multiDeployment(),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "foo",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app-1",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app-1",
						},
					},
				},
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "bar",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app-2",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app-2",
						},
					},
				},
			},
		},
		{
			name: "dobs pod owned by deployment -- with legacy driver",
			objs: deployment(pvcDobs("", LegacyCSIDriver)),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "foo",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app",
						},
					},
				},
			},
		},
		{
			name: "dobs pod owned by deployment, with deprecated pvc annotation -- with legacy driver",
			objs: deployment(provideStorageClassInAnnotation(pvcDobs("", LegacyCSIDriver))),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Pod referencing DOBS volumes must be owned by StatefulSet",
					Kind:     checks.Pod,
					Object: &metav1.ObjectMeta{
						Name:      "foo",
						Namespace: metav1.NamespaceDefault,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "web-app",
							},
						},
					},
					Owners: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "web-app",
						},
					},
				},
			},
		},
		{
			name:     "dobs pod owned by statefulset",
			objs:     statefulSet(pvcDobs("", DOCSIDriver)),
			expected: nil,
		},
		{
			name:     "dobs pod owned by statefulset, with deprecated pvc annotation",
			objs:     statefulSet(provideStorageClassInAnnotation(pvcDobs("", DOCSIDriver))),
			expected: nil,
		},
		{
			name:     "dobs pod owned by statefulset -- with legacy driver",
			objs:     statefulSet(pvcDobs("", LegacyCSIDriver)),
			expected: nil,
		},
		{
			name:     "dobs pod owned by statefulset, with deprecated pvc annotation -- with legacy driver",
			objs:     statefulSet(provideStorageClassInAnnotation(pvcDobs("", LegacyCSIDriver))),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := dobsPodOwner.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func csiDobs(driver string) *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "csi-do",
								VolumeSource: corev1.VolumeSource{
									CSI: &corev1.CSIVolumeSource{
										Driver: driver,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return objs
}

func statefulSet(objs *kube.Objects) *kube.Objects {
	objs.Pods.Items[0].OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Name:       "woot",
		},
	}
	return objs
}

func deployment(objs *kube.Objects) *kube.Objects {
	objs.Pods.Items[0].OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "web-app",
		},
	}
	return objs
}

// Users request dynamically provisioned storage by including a storage class in their PersistentVolumeClaim.
// Before Kubernetes v1.6, this was done via the volume.beta.kubernetes.io/storage-class annotation. However, this annotation is deprecated since v1.9. 
// https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/#:~:text=Users%20request%20dynamically,the%20PersistentVolumeClaim%20object.
func provideStorageClassInAnnotation(objs *kube.Objects) *kube.Objects {
	if objs.PersistentVolumeClaims.Items[0].Spec.StorageClassName != nil {
		if objs.PersistentVolumeClaims.Items[0].Annotations == nil {
			objs.PersistentVolumeClaims.Items[0].Annotations = make(map[string]string)
		}
		objs.PersistentVolumeClaims.Items[0].Annotations["volume.beta.kubernetes.io/storage-class"] = *objs.PersistentVolumeClaims.Items[0].Spec.StorageClassName
		objs.PersistentVolumeClaims.Items[0].Spec.StorageClassName = nil
	}
	return objs
}

func multiDeployment() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "def-pvc-source-1",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "def-pvc-1",
									},
								},
							},
						},
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "def-pvc-source-2",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "def-pvc-2",
									},
								},
							},
						},
					},
				},
			},
		},
		PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{
			Items: []corev1.PersistentVolumeClaim{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "def-pvc-1", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "dobs-v1",
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "def-pvc-2", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "dobs-v2",
					},
				},
			},
		},
		StorageClasses: &st.StorageClassList{
			Items: []st.StorageClass{
				{
					TypeMeta:    metav1.TypeMeta{Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
					ObjectMeta:  metav1.ObjectMeta{Name: DOBlockStorageName, Namespace: metav1.NamespaceDefault},
					Provisioner: DOCSIDriver,
				},
			},
		},
		DefaultStorageClass: &st.StorageClass{
			TypeMeta:    metav1.TypeMeta{Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
			ObjectMeta:  metav1.ObjectMeta{Name: DOBlockStorageName, Namespace: metav1.NamespaceDefault},
			Provisioner: DOCSIDriver,
		},
	}

	objs.Pods.Items[0].OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "web-app-1",
		},
	}
	objs.Pods.Items[1].OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "web-app-2",
		},
	}

	return objs
}

func pvcDobs(storageClass, driver string) *kube.Objects {
	var sc *string
	if storageClass != "" {
		sc = &storageClass
	}

	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "def-pvc-source",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "def-pvc",
									},
								},
							},
						},
					},
				},
			},
		},
		PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{
			Items: []corev1.PersistentVolumeClaim{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "def-pvc", Namespace: metav1.NamespaceDefault},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "dobs-v1",
						StorageClassName: sc,
					},
				},
			},
		},
		StorageClasses: &st.StorageClassList{
			Items: []st.StorageClass{
				{
					TypeMeta:    metav1.TypeMeta{Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
					ObjectMeta:  metav1.ObjectMeta{Name: DOBlockStorageName, Namespace: metav1.NamespaceDefault},
					Provisioner: driver,
				},
			},
		},
		DefaultStorageClass: &st.StorageClass{
			TypeMeta:    metav1.TypeMeta{Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
			ObjectMeta:  metav1.ObjectMeta{Name: DOBlockStorageName, Namespace: metav1.NamespaceDefault},
			Provisioner: driver,
		},
	}
	return objs
}

func noDobs() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "bar",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: "secret_foo",
									},
								},
							},
						},
					},
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
