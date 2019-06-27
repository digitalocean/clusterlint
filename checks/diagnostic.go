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

type Severity string
type Kind string

const (
	Error                 Severity = "error"
	Warning               Severity = "warning"
	Suggestion            Severity = "suggestion"
	Pod                   Kind     = "pod"
	PodTemplate           Kind     = "pod template"
	PersistentVolumeClaim Kind     = "persistent volume claim"
	ConfigMap             Kind     = "config map"
	Service               Kind     = "service"
	Secret                Kind     = "secret"
	ServiceAccount        Kind     = "service account"
	PersistentVolume      Kind     = "persistent volume"
)
