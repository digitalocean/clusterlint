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

package main

import (
	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
)

type check struct{}

// Name returns a unique name for this check.
func (nc *check) Name() string {
	return "example-plugin"
}

// Groups returns a list of group names this check should be part of.
func (nc *check) Groups() []string {
	return []string{"examples"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *check) Description() string {
	return "A sample plugin."
}

// Run runs this check on a set of Kubernetes objects.
func (nc *check) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	d := make([]checks.Diagnostic, len(objects.Pods.Items))
	for i, p := range objects.Pods.Items {
		d[i] = checks.Diagnostic{
			Message:  "You probably don't want to run the example plugin.",
			Severity: checks.Suggestion,
			Kind:     checks.Pod,
			Object:   &p.ObjectMeta,
			Owners:   p.GetOwnerReferences(),
		}
	}
	return d, nil
}

func init() {
	checks.Register(&check{})
}
