package basic

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnusedPVCheckMeta(t *testing.T) {
	unusedPVCheck := unusedPVCheck{}
	assert.Equal(t, "unused-pv", unusedPVCheck.Name())
	assert.Equal(t, []string{"basic"}, unusedPVCheck.Groups())
	assert.NotEmpty(t, unusedPVCheck.Description())
}

func TestUnusedPVCheckRegistration(t *testing.T) {
	unusedPVCheck := &unusedPVCheck{}
	check, err := checks.Get("unused-pv")
	assert.NoError(t, err)
	assert.Equal(t, check, unusedPVCheck)
}

func TestUnusedPVWarning(t *testing.T) {
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pvs",
			objs:     &kube.Objects{PersistentVolumes: &corev1.PersistentVolumeList{}},
			expected: nil,
		},
		{
			name:     "bound pv",
			objs:     bound(),
			expected: nil,
		},
		{
			name: "unused pv",
			objs: unused(),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "Unused Persistent Volume 'pv_foo'.",
					Kind:     checks.PersistentVolume,
					Object:   &metav1.ObjectMeta{Name: "pv_foo"},
					Owners:   GetOwners(),
				},
			},
		},
	}

	unusedPVCheck := unusedPVCheck{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := unusedPVCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func initPV() *kube.Objects {
	objs := &kube.Objects{
		PersistentVolumes: &corev1.PersistentVolumeList{
			Items: []corev1.PersistentVolume{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: "pv_foo"},
				},
			},
		},
	}
	return objs
}

func bound() *kube.Objects {
	objs := initPV()
	objs.PersistentVolumes.Items[0].Spec = corev1.PersistentVolumeSpec{
		ClaimRef: &corev1.ObjectReference{
			Kind:      "PersistentVolumeClaim",
			Name:      "foo",
			Namespace: "k8s",
		},
	}
	return objs
}

func unused() *kube.Objects {
	objs := initPV()
	objs.PersistentVolumes.Items[0].Spec = corev1.PersistentVolumeSpec{}
	return objs
}