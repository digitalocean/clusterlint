package checks

import (
	"fmt"

	"github.com/digitalocean/clusterlint/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Diagnostic encapsulates the information each check returns.
type Diagnostic struct {
	Severity Severity
	Message  string
	Object   kube.Object
	Owners   []metav1.OwnerReference
}

func (d *Diagnostic) String() string {
	return fmt.Sprintf("[%s] %s/%s/%s: %s", d.Severity, d.Object.ObjectInfo.Namespace,
		d.Object.TypeInfo.Kind, d.Object.ObjectInfo.Name, d.Message)
}

type Severity string

const (
	Error      Severity = "error"
	Warning    Severity = "warning"
	Suggestion Severity = "suggestion"
)
