package basic

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnusedPVCCheckMeta(t *testing.T) {
	unusedClaimCheck := unusedClaimCheck{}
	assert.Equal(t, "unused-pvc", unusedClaimCheck.Name())
	assert.Equal(t, []string{"basic"}, unusedClaimCheck.Groups())
	assert.NotEmpty(t, unusedClaimCheck.Description())
}

func TestUnusedPVCCheckRegistration(t *testing.T) {
	unusedClaimCheck := &unusedClaimCheck{}
	check, err := checks.Get("unused-pvc")
	assert.NoError(t, err)
	assert.Equal(t, check, unusedClaimCheck)
}

func TestUnusedPVCWarning(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pvcs",
			objs:     &kube.Objects{Pods: &corev1.PodList{}, PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{}},
			expected: nil,
		},
		{
			name: "pod with pvc",
			objs: boundPVC(corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "pvc_foo",
				},
			}),
			expected: nil,
		},
		{
			name: "unused pvc",
			objs: initPVC(),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Unused persistent volume claim",
					Kind:     checks.PersistentVolumeClaim,
					Object:   &metav1.ObjectMeta{Name: "pvc_foo", Namespace: "k8s"},
					Owners:   GetOwners(),
				},
			},
		},
	}

	unusedClaimCheck := unusedClaimCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := unusedClaimCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initPVC() *kube.Objects {
	objs := &kube.Objects{
		Pods: &corev1.PodList{
			Items: []corev1.Pod{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pod_foo", Namespace: "k8s"},
				},
			},
		},
		PersistentVolumeClaims: &corev1.PersistentVolumeClaimList{
			Items: []corev1.PersistentVolumeClaim{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pvc_foo", Namespace: "k8s"},
				},
			},
		},
	}
	return objs
}

func boundPVC(volumeSrc corev1.VolumeSource) *kube.Objects {
	objs := initPVC()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name:         "bar",
				VolumeSource: volumeSrc,
			}},
	}
	return objs
}
