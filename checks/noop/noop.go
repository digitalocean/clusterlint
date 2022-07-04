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

package noop

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

func init() {
	checks.Register(&check{})
}

type check struct{}

// Name returns a unique name for this check.
func (nc *check) Name() string {
	return "noop"
}

// Groups returns a list of group names this check should be part of.
func (nc *check) Groups() []string {
	return nil
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *check) Description() string {
	return "Does not check anything. Returns no errors."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *check) Run(*kube.Objects) ([]checks.Diagnostic, error) {
	return nil, nil
}
