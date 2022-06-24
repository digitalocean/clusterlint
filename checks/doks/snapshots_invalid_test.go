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

package doks

import (
	"testing"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/stretchr/testify/assert"
)

func TestInvalidSnapshotCheckMeta(t *testing.T) {
	invalidSnapshotCheck := invalidSnapshotCheck{}
	assert.Equal(t, "invalid-volume-snapshot", invalidSnapshotCheck.Name())
	assert.Equal(t, []string{"doks"}, invalidSnapshotCheck.Groups())
	assert.NotEmpty(t, invalidSnapshotCheck.Description())
}

func TestInvalidSnapshotCheckRegistration(t *testing.T) {
	invalidSnapshotCheck := &invalidSnapshotCheck{}
	check, err := checks.Get("invalid-volume-snapshot")
	assert.NoError(t, err)
	assert.Equal(t, check, invalidSnapshotCheck)
}

func TestInvalidSnapshots(t *testing.T) {
	invalidSnapshotCheck := invalidSnapshotCheck{}

	tests := []struct {
		name     string
		objs     *kube.Objects
		expected []checks.Diagnostic
	}{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := invalidSnapshotCheck.Run(test.objs)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, d)
		})
	}
}
