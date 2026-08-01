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

package kube

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceError(t *testing.T) {
	_, err := NewObjectFilter([]string{"namespace-1"}, []string{"namespace-2"})

	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("cannot specify both include and exclude namespace conditions"), err)
}

func TestNamespaceOptions(t *testing.T) {
	filter, err := NewObjectFilter([]string{"namespace-1"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t,
		metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.namespace", "namespace-1").String()},
		filter.NamespaceOptions(metav1.ListOptions{}),
	)

	filter, err = NewObjectFilter([]string{}, []string{"namespace-2"})
	assert.NoError(t, err)
	assert.Equal(t,
		metav1.ListOptions{FieldSelector: fields.OneTermNotEqualSelector("metadata.namespace", "namespace-2").String()},
		filter.NamespaceOptions(metav1.ListOptions{}),
	)

	filter, err = NewObjectFilter([]string{"namespace-1", "namespace-2"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t,
		metav1.ListOptions{FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("metadata.namespace", "namespace-1"),
			fields.OneTermEqualSelector("metadata.namespace", "namespace-2"),
		).String()},
		filter.NamespaceOptions(metav1.ListOptions{}),
	)

	filter, err = NewObjectFilter([]string{}, []string{"namespace-3", "namespace-4"})
	assert.NoError(t, err)
	assert.Equal(t,
		metav1.ListOptions{FieldSelector: fields.AndSelectors(
			fields.OneTermNotEqualSelector("metadata.namespace", "namespace-3"),
			fields.OneTermNotEqualSelector("metadata.namespace", "namespace-4"),
		).String()},
		filter.NamespaceOptions(metav1.ListOptions{}),
	)
}
