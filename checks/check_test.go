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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckIsDisabled(t *testing.T) {
	const name string = "pod_foo"
	pod := initPod(name)
	assert.False(t, IsEnabled(name, &pod.ObjectMeta))
}

func TestCheckIsEnabled(t *testing.T) {
	const name string = "pod_foo"
	pod := initPod("")
	assert.True(t, IsEnabled(name, &pod.ObjectMeta))
}

func TestNoClusterlintAnnotation(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod_foo", Annotations: map[string]string{"porg": "star wars"},
		},
	}
	assert.True(t, IsEnabled("pod_foo", &pod.ObjectMeta))
}

func initPod(name string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod_foo", Annotations: map[string]string{checkAnnotation: fmt.Sprintf("%s, bar, baz", name)},
		},
	}
	return pod
}
