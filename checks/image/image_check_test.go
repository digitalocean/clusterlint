/*
Copyright 2019 DigitalOcean

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

package image

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckImageError(t *testing.T) {
	validRepoNames := []string{
		"docker/docker",
		"library/debian",
		"debian",
		"docker.io/docker/docker",
		"docker.io/library/debian",
		"docker.io/debian",
		"index.docker.io/docker/docker",
		"index.docker.io/library/debian",
		"index.docker.io/debian",
		"127.0.0.1:5000/docker/docker",
		"127.0.0.1:5000/library/debian",
		"127.0.0.1:5000/debian",
		"thisisthesongthatneverendsitgoesonandonandonthisisthesongthatnev",
		"busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",

		// This test case was moved from invalid to valid since it is valid input
		// when specified with a hostname, it removes the ambiguity from about
		// whether the value is an identifier or repository name
		"docker.io/1a3f5e7d9c1b3a5f7e9d1c3b5a7f9e1d3c5b7a9f1e3d5d7c9b1a3f5e7d9c1b3a",
		// Allow embedded hyphens.
		"docker-rules/docker",
		// Allow multiple hyphens as well.
		"docker---rules/docker",
		//Username doc and image name docker being tested.
		"doc/docker",
		// single character names are now allowed.
		"d/docker",
		"jess/t",
		// Consecutive underscores.
		"dock__er/docker",
	}
	invalidRepoNames := []string{
		"https://github.com/docker/docker",
		"docker/Docker",
		"-docker",
		"-docker/docker",
		"-docker.io/docker/docker",
		"docker///docker",
		"docker.io/docker/Docker",
		"docker.io/docker///docker",
		"1a3f5e7d9c1b3a5f7e9d1c3b5a7f9e1d3c5b7a9f1e3d5d7c9b1a3f5e7d9c1b3a",
		// Don't allow underscores everywhere (as opposed to hyphens).
		"____/____",

		"_docker/_docker",

		// Disallow consecutive periods.
		"dock..er/docker",
		"dock_.er/docker",
		"dock-.er/docker",

		// No repository.
		"docker/",

		//namespace too long
		"this_is_not_a_valid_namespace_because_its_lenth_is_greater_than_255_this_is_not_a_valid_namespace_because_its_lenth_is_greater_than_255_this_is_not_a_valid_namespace_because_its_lenth_is_greater_than_255_this_is_not_a_valid_namespace_because_its_lenth_is_greater_than_255/docker",
	}
	for _, v := range validRepoNames {
		t.Run(fmt.Sprintf("good-repo-%s", v), func(t *testing.T) {
			_, err := checkImageError("foo", v)
			assert.NoError(t, err)
		})

	}
	for _, iv := range invalidRepoNames {
		t.Run(fmt.Sprintf("bad-repo-%s", iv), func(t *testing.T) {
			_, err := checkImageError("foo", iv)
			assert.Error(t, err)
		})
	}
}

func TestLintImage(t *testing.T) {
	tests := []struct {
		name         string
		image        string
		validDomains map[string]struct{}
		warnings     []error
	}{
		{
			name:  "no domain",
			image: "digitalocean/foo:latest",
			warnings: []error{
				errEmptyDomain,
				errNoSha,
			},
		},
		{
			name:  "invalid domain",
			image: "docker.io/digitalocean/foo:latest",
			validDomains: map[string]struct{}{
				"docr.space": struct{}{},
			},
			warnings: []error{
				errInvalidDomain,
				errNoSha,
			},
		},
		{
			name:  "docker.io fine sometimes",
			image: "docker.io/digitalocean/foo:latest",
			validDomains: map[string]struct{}{
				"docker.io": struct{}{},
			},
			warnings: []error{
				errNoSha,
			},
		},
		{
			name:  "any domain fine with none valid",
			image: "docker.io/digalocean/foo:latest",
			warnings: []error{
				errNoSha,
			},
		},
		{
			name:  "ok",
			image: "docr.space/doks/foo@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			validDomains: map[string]struct{}{
				"docr.space": struct{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := checkImageError("foo", tt.image)
			assert.NoError(t, err)

			warnings := lintImage(ref, tt.validDomains, "foo", tt.image)
			wL := len(tt.warnings)
			assert.Len(t, warnings, wL)
			for _, w := range warnings {
				for _, e := range tt.warnings {
					is := errors.Is(w, e)
					if is {
						wL--
						break
					}
				}
			}
			assert.Equal(t, wL, 0, "one or more of the warnings did not match expectations")
		})
	}

}
