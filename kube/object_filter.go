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

package kube

import (
	// Load client-go authentication plugins
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// ObjectFilter stores k8s object's fields that needs to be included or excluded while running checks
type ObjectFilter struct {
	IncludeNamespace string
	ExcludeNamespace string
}

// NewObjectFilter is a constructor to initialize an instance of ObjectFilter
func NewObjectFilter(includeNamespace, excludeNamespace string) (ObjectFilter, error) {
	if len(includeNamespace) > 0 && len(excludeNamespace) > 0 {
		return ObjectFilter{}, fmt.Errorf("cannot specify both include and exclude namespace conditions")
	}
	return ObjectFilter{
		IncludeNamespace: includeNamespace,
		ExcludeNamespace: excludeNamespace,
	}, nil
}

// NamespaceOptions returns ListOptions for filtering by namespace
func (f ObjectFilter) NamespaceOptions(opts metav1.ListOptions) metav1.ListOptions {
	if len(f.IncludeNamespace) > 0 {
		opts.FieldSelector = fields.OneTermEqualSelector("metadata.namespace", f.IncludeNamespace).String()
	}
	if len(f.ExcludeNamespace) > 0 {
		opts.FieldSelector = fields.OneTermNotEqualSelector("metadata.namespace", f.ExcludeNamespace).String()
	}
	return opts
}
