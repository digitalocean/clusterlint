package basic

import (
	"fmt"
	"strings"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	checks.Register(&defaultNamespaceCheck{})
}

type defaultNamespaceCheck struct{}

type alert struct {
	warnings []error
	mu       sync.Mutex
}

// GetWarnings returns alert.warnings
func (alert *alert) GetWarnings() []error {
	return alert.warnings
}

// SetWarnings sets alert.warnings
func (alert *alert) SetWarnings(w []error) {
	alert.warnings = w
}

// AppendWarning appends a warning to the warnings slice
func (alert *alert) AppendWarning(err error) {
	alert.warnings = append(alert.warnings, err)
}

// warn adds warnings for k8s objects that should not be in the default namespace
func (alert *alert) warn(k8stype string, item metav1.ObjectMeta) {
	alert.mu.Lock()
	alert.AppendWarning(fmt.Errorf("%s '%s' is in the default namespace.", k8stype, item.GetName()))
	alert.mu.Unlock()
}

// Name returns a unique name for this check.
func (nc *defaultNamespaceCheck) Name() string {
	return "default-namespace"
}

// Groups returns a list of group names this check should be part of.
func (nc *defaultNamespaceCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *defaultNamespaceCheck) Description() string {
	return "Checks if there are any user created k8s objects in the default namespace."
}

// checkPods checks if there are pods in the default namespace
func checkPods(items *corev1.PodList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn("Pod", item.ObjectMeta)
		}
	}
}

// checkPodTemplates checks if there are pod templates in the default namespace
func checkPodTemplates(items *corev1.PodTemplateList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn("Pod template", item.ObjectMeta)
		}
	}
}

// checkPVCs checks if there are pvcs in the default namespace
func checkPVCs(items *corev1.PersistentVolumeClaimList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn("Persistent Volume Claim", item.ObjectMeta)
		}
	}
}

// checkConfigMaps checks if there are config maps in the default namespace
func checkConfigMaps(items *corev1.ConfigMapList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() {
			alert.warn("Config Map", item.ObjectMeta)
		}
	}
}

// checkServices checks if there are user created services in the default namespace
func checkServices(items *corev1.ServiceList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && item.GetName() != "kubernetes" {
			alert.warn("Service", item.ObjectMeta)
		}
	}
}

// checkSecrets checks if there are user created secrets in the default namespace
func checkSecrets(items *corev1.SecretList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && !strings.Contains(item.GetName(), "default-token-") {
			alert.warn("Secret", item.ObjectMeta)
		}
	}
}

// checkSA checks if there are user created SAs in the default namespace
func checkSA(items *corev1.ServiceAccountList, alert *alert) {
	for _, item := range items.Items {
		if corev1.NamespaceDefault == item.GetNamespace() && item.GetName() != "default" {
			alert.warn("Service Account", item.ObjectMeta)
		}
	}
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *defaultNamespaceCheck) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	alert := &alert{}
	var g errgroup.Group
	g.Go(func() error {
		checkPods(objects.Pods, alert)
		return nil
	})

	g.Go(func() error {
		checkPodTemplates(objects.PodTemplates, alert)
		return nil
	})

	g.Go(func() error {
		checkPVCs(objects.PersistentVolumeClaims, alert)
		return nil
	})

	g.Go(func() error {
		checkConfigMaps(objects.ConfigMaps, alert)
		return nil
	})

	g.Go(func() error {
		checkServices(objects.Services, alert)
		return nil
	})

	g.Go(func() error {
		checkSecrets(objects.Secrets, alert)
		return nil
	})

	g.Go(func() error {
		checkSA(objects.ServiceAccounts, alert)
		return nil
	})

	return alert.warnings, nil, g.Wait()
}
