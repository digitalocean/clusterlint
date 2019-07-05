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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Diagnostic encapsulates the information each check returns.
type Diagnostic struct {
	Check    string
	Severity Severity
	Message  string
	Kind     Kind
	Object   *metav1.ObjectMeta
	Owners   []metav1.OwnerReference
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("[%s] %s/%s/%s: %s", d.Severity, d.Object.Namespace,
		d.Kind, d.Object.Name, d.Message)
}

// Severity identifies the level of priority for each diagnostic.
type Severity string

// Kind represents the kind of k8s object the diagnoatic is about
type Kind string

const (
	// Error means that the diagnostic message should be fixed immediately
	Error Severity = "error"
	// Warning indicates that the diagnostic should be fixed but okay to ignore in some special cases
	Warning Severity = "warning"
	// Suggestion means that a user need not implement it, but is in line with the recommended best practices
	Suggestion Severity = "suggestion"
	// Pod identifies Kubernetes objects of kind `pod`
	Pod Kind = "pod"
	// PodTemplate identifies Kubernetes objects of kind `pod template`
	PodTemplate Kind = "pod_template"
	// PersistentVolumeClaim identifies Kubernetes objects of kind `persistent volume claim`
	PersistentVolumeClaim Kind = "persistent_volume_claim"
	// ConfigMap identifies Kubernetes objects of kind `config map`
	ConfigMap Kind = "config_map"
	// Service identifies Kubernetes objects of kind `service`
	Service Kind = "service"
	// Secret identifies Kubernetes objects of kind `secret`
	Secret Kind = "secret"
	// ServiceAccount identifies Kubernetes objects of kind `service account`
	ServiceAccount Kind = "service_account"
	// PersistentVolume identifies Kubernetes objects of kind `persistent volume`
	PersistentVolume Kind = "persistent_volume"
	// ValidatingWebhookConfiguration identifies Kubernetes objects of kind `validating webhook configuration`
	ValidatingWebhookConfiguration Kind = "validating_webhook_configuration"
	// MutatingWebhookConfiguration identifies Kubernetes objects of kind `mutating webhook configuration`
	MutatingWebhookConfiguration Kind = "mutating_webhook_configuration"
	// Deployment identifies Kubernetes objects of kind `deployment`
	Deployment Kind = "deployment"
	// DaemonSet identifies Kubernetes objects of kind `daemon_set`
	DaemonSet Kind = "daemon_set"
	// StatefulSet identifies Kubernetes objects of kind `stateful_set`
	StatefulSet Kind = "stateful_set"
	// Ingress  identifies Kubernetes objects of kind `ingress`
	Ingress Kind = "ingress"
)
