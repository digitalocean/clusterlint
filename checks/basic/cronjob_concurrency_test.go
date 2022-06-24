/*
Copyright 2022 DigitalOcean

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

	"github.com/stretchr/testify/assert"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func TestCronJobConcurrencyMeta(t *testing.T) {
	check := cronJobConcurrencyCheck{}
	assert.Equal(t, "cronjob-concurrency", check.Name())
	assert.Equal(t, []string{"basic"}, check.Groups())
	assert.NotEmpty(t, check.Description())
}

func TestCronJobConcurrencyCheckRegistration(t *testing.T) {
	check := &cronJobConcurrencyCheck{}
	ch, err := checks.Get("cronjob-concurrency")
	assert.NoError(t, err)
	assert.Equal(t, ch, check)
}

func TestCronJobConcurrency(t *testing.T) {
	check := cronJobConcurrencyCheck{}
	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{
		{
			name:     "cronjob with 'Forbid' policy",
			objs:     policy(batchv1beta1.ForbidConcurrent),
			expected: nil,
		},
		{
			name:     "cronjob with 'Replace' policy",
			objs:     policy(batchv1beta1.ReplaceConcurrent),
			expected: nil,
		},
		{
			name: "cronjob with 'Allow' policy",
			objs: policy(batchv1beta1.AllowConcurrent),
			expected: []checks.Diagnostic{
				{
					Severity: checks.Warning,
					Message:  "CronJob has a concurrency policy of `Allow`. Prefer to use `Forbid` or `Replace`",
					Kind:     checks.CronJob,
					Object:   &metav1.ObjectMeta{Name: "cronjob_foo"},
					Owners:   GetOwners(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := check.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}

func policy(policy batchv1beta1.ConcurrencyPolicy) *kube.Objects {
	objs := initCronJob()
	objs.CronJobs.Items[0].Spec = batchv1beta1.CronJobSpec{
		ConcurrencyPolicy: policy,
	}
	return objs
}

func initCronJob() *kube.Objects {
	objs := &kube.Objects{
		CronJobs: &batchv1beta1.CronJobList{
			Items: []batchv1beta1.CronJob{
				{
					TypeMeta:   metav1.TypeMeta{Kind: "CronJob", APIVersion: "batch/v1beta1"},
					ObjectMeta: metav1.ObjectMeta{Name: "cronjob_foo"},
				},
			},
		},
	}
	return objs
}
