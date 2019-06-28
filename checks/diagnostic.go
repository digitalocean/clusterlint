package checks

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Diagnostic encapsulates the information each check returns.
type Diagnostic struct {
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
	PodTemplate Kind = "pod template"
	// PersistentVolumeClaim identifies Kubernetes objects of kind `persistent volume claim`
	PersistentVolumeClaim Kind = "persistent volume claim"
	// ConfigMap identifies Kubernetes objects of kind `config map`
	ConfigMap Kind = "config map"
	// Service identifies Kubernetes objects of kind `service`
	Service Kind = "service"
	// Secret identifies Kubernetes objects of kind `secret`
	Secret Kind = "secret"
	// ServiceAccount identifies Kubernetes objects of kind `service account`
	ServiceAccount Kind = "service account"
	// PersistentVolume identifies Kubernetes objects of kind `persistent volume`
	PersistentVolume Kind = "persistent volume"
)
