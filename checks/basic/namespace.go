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

const namespace = "default"

func init() {
	checks.Register(&defaultNamespaceCheck{})
}

type defaultNamespaceCheck struct{}

type alert struct {
	Warnings []error
	Errors   []error
	mu       sync.Mutex
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

// warn adds warnings for k8s objects that should not be in the default namespace
func warn(k8stype string, item metav1.ObjectMeta, alert *alert) {
	if namespace == item.GetNamespace() {
		alert.mu.Lock()
		alert.Warnings = append(alert.Warnings, fmt.Errorf("%s '%s' is in the default namespace", k8stype, item.GetName()))
		alert.mu.Unlock()
	}
}

// collect retrieves all objects of a specific type that are in
// default namespace
func collect(item metav1.ObjectMeta, names *[]string, guard *sync.Mutex) {
	if namespace == item.GetNamespace() {
		guard.Lock()
		*names = append(*names, item.GetName())
		guard.Unlock()
	}
}

// warning adds a special warning for secrets, services and SAs which have
// one default k8s object in the default namespace.
func warning(k8stype string, obj []string, alert *alert) {
	if len(obj) > 1 {
		alert.mu.Lock()
		alert.Warnings = append(alert.Warnings, fmt.Errorf("There are user created %s defined in the default namespace: %s.", k8stype, strings.Join(obj, ",")))
		alert.mu.Unlock()
	}
}

// checkPods checks if there are pods in the default namespace
func checkPods(items *corev1.PodList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Pod", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkPodTemplates checks if there are pod templates in the default namespace
func checkPodTemplates(items *corev1.PodTemplateList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Pod template", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkPVCs checks if there are pvcs in the default namespace
func checkPVCs(items *corev1.PersistentVolumeClaimList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Persistent Volume Claim", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkConfigMaps checks if there are config maps in the default namespace
func checkConfigMaps(items *corev1.ConfigMapList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Config Map", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkQuotas checks if there are quotas in the default namespace
func checkQuotas(items *corev1.ResourceQuotaList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Resource Quota", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkLimits checks if there are limits in the default namespace
func checkLimits(items *corev1.LimitRangeList, alert *alert) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Limit Range", item.ObjectMeta, alert)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkServices checks if there are user created services in the default namespace
func checkServices(items *corev1.ServiceList, alert *alert) error {
	var g errgroup.Group
	var names []string
	var guard sync.Mutex
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			collect(item.ObjectMeta, &names, &guard)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	warning("services", names, alert)
	return nil
}

// checkSecrets checks if there are user created secrets in the default namespace
func checkSecrets(items *corev1.SecretList, alert *alert) error {
	var g errgroup.Group
	var names []string
	var guard sync.Mutex
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			collect(item.ObjectMeta, &names, &guard)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	warning("secrets", names, alert)
	return nil
}

// checkSA checks if there are user created SAs in the default namespace
func checkSA(items *corev1.ServiceAccountList, alert *alert) error {
	var g errgroup.Group
	var names []string
	var guard sync.Mutex
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			collect(item.ObjectMeta, &names, &guard)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	warning("service account", names, alert)
	return nil
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *defaultNamespaceCheck) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	alert := &alert{}
	var g errgroup.Group
	g.Go(func() error {
		return checkPods(objects.Pods, alert)
	})

	g.Go(func() error {
		return checkPodTemplates(objects.PodTemplates, alert)
	})

	g.Go(func() error {
		return checkPVCs(objects.PersistentVolumeClaims, alert)
	})

	g.Go(func() error {
		return checkConfigMaps(objects.ConfigMaps, alert)
	})

	g.Go(func() error {
		return checkQuotas(objects.ResourceQuotas, alert)
	})

	g.Go(func() error {
		return checkLimits(objects.LimitRanges, alert)
	})

	g.Go(func() error {
		return checkServices(objects.Services, alert)
	})

	g.Go(func() error {
		return checkSecrets(objects.Secrets, alert)
	})

	g.Go(func() error {
		return checkSA(objects.ServiceAccounts, alert)
	})

	return alert.Warnings, alert.Errors, g.Wait()
}
