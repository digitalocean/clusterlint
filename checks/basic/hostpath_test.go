package basic

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestHostpathCheckMeta(t *testing.T) {
	hostPathCheck := hostPathCheck{}
	assert.Equal(t, "hostpath-volume", hostPathCheck.Name())
	assert.Equal(t, "Check if there are pods using hostpath volumes", hostPathCheck.Description())
	assert.Equal(t, []string{"basic"}, hostPathCheck.Groups())
}

func TestHostpathCheckRegistration(t *testing.T) {
	hostPathCheck := &hostPathCheck{}
	check, err := checks.Get("hostpath-volume")
	assert.Equal(t, check, hostPathCheck)
	assert.Nil(t, err)
}

func TestHostpathVolumeError(t *testing.T) {
	scenarios := []struct {
		name     string
		arg      *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "no pods",
			arg:      initPod(),
			expected: nil,
		},
		{
			name:     "pod with no volumes",
			arg:      container("docker.io/nginx:foo"),
			expected: nil,
		},
		{
			name: "pod with other volume",
			arg: volume(corev1.VolumeSource{
				GitRepo: &corev1.GitRepoVolumeSource{Repository: "boo"},
			}),
			expected: nil,
		},
		{
			name: "pod with hostpath volume",
			arg: volume(corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/tmp"},
			}),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Error,
					Message:  "Avoid using hostpath for volume 'bar'.",
					Kind:     checks.Pod,
					Object:   GetObjectMeta(),
					Owners:   GetOwners(),
				},
			},
		},
	}

	hostPathCheck := hostPathCheck{}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d, err := hostPathCheck.Run(scenario.arg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, scenario.expected, d)
		})
	}
}

func volume(volumeSrc corev1.VolumeSource) *kube.Objects {
	objs := initPod()
	objs.Pods.Items[0].Spec = corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name:         "bar",
				VolumeSource: volumeSrc,
			}},
	}
	return objs
}
