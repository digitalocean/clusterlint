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

package checks

import (
	"context"
	"strings"

	"github.com/digitalocean/clusterlint/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const checkAnnotation = "clusterlint.digitalocean.com/disabled-checks"
const separator = ","

const (
	// Repositories is a context key for valid repositories
	Repositories = "repositories"
)

// Check is a check that can run on Kubernetes objects.
type Check interface {
	// Name returns a unique name for this check.
	Name() string
	// Groups returns a list of group names this check should be part of. It is
	// valid to return nil or an empty list if a check does not belong in any
	// groups.
	Groups() []string
	// Description returns a detailed human-readable description of what this
	// check does.
	Description() string
	// Run runs this check on a set of Kubernetes objects. It can return
	// warnings (low-priority problems) and errors (high-priority problems) as
	// well as an error value indicating that the check failed to run.
	Run(context.Context, *kube.Objects) ([]Diagnostic, error)
}

// IsEnabled inspects the object annotations to see if a check is disabled
func IsEnabled(name string, item *metav1.ObjectMeta) bool {
	annotations := item.GetAnnotations()
	if value, ok := annotations[checkAnnotation]; ok {
		disabledChecks := strings.Split(value, separator)
		if contains(disabledChecks, name) {
			return false
		}
	}
	return true
}
