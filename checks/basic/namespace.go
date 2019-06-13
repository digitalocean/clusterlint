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
	checks.Register(&check{})
}

type check struct{}

// Name returns a unique name for this check.
func (nc *check) Name() string {
	return "namespace"
}

// Groups returns a list of group names this check should be part of.
func (nc *check) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (nc *check) Description() string {
	return "Checks if there are any user created k8s objects in the default namespace."
}

// warn adds warnings for k8s objects that should not be in the default namespace
func warn(k8stype string, item metav1.ObjectMeta, w *[]error, mu *sync.Mutex) {
	if namespace == item.GetNamespace() {
		mu.Lock()
		*w = append(*w, fmt.Errorf("%s '%s' is in the default namespace", k8stype, item.GetName()))
		mu.Unlock()
	}
}

// collect retrieves all objects of a specific type that are in
// default namespace
func collect(item interface{}, names *[]string, guard *sync.Mutex) {
	obj, ok := item.(metav1.ObjectMeta)
	if ok {
		if namespace == obj.GetNamespace() {
			guard.Lock()
			*names = append(*names, obj.GetName())
			guard.Unlock()
		}
	}
}

// warning adds a special warning for secrets, services and SAs which have
// one default k8s object in the default namespace.
func warning(k8stype string, obj []string, w *[]error, mu *sync.Mutex) {
	if len(obj) > 1 {
		mu.Lock()
		*w = append(*w, fmt.Errorf("There are user created %s defined in the default namespace: %s.", k8stype, strings.Join(obj, ",")))
		mu.Unlock()
	}
}

// checkPods checks if there are pods in the default namespace
func checkPods(items *corev1.PodList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Pod", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkPodTemplates checks if there are pod templates in the default namespace
func checkPodTemplates(items *corev1.PodTemplateList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Pod template", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkPVCs checks if there are pvcs in the default namespace
func checkPVCs(items *corev1.PersistentVolumeClaimList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Persistent Volume Claim", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkConfigMaps checks if there are config maps in the default namespace
func checkConfigMaps(items *corev1.ConfigMapList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Config Map", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkQuotas checks if there are quotas in the default namespace
func checkQuotas(items *corev1.ResourceQuotaList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Resource Quota", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkLimits checks if there are limits in the default namespace
func checkLimits(items *corev1.LimitRangeList, w *[]error, mu *sync.Mutex) error {
	var g errgroup.Group
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			warn("Limit Range", item.ObjectMeta, w, mu)
			return nil
		})
	}
	err := g.Wait()
	return err
}

// checkServices checks if there are user created services in the default namespace
func checkServices(items *corev1.ServiceList, w *[]error, mu *sync.Mutex) error {
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
	err1 := g.Wait()
	if err1 != nil {
		return err1
	}
	warning("services", names, w, mu)
	return nil
}

// checkSecrets checks if there are user created secrets in the default namespace
func checkSecrets(items *corev1.SecretList, w *[]error, mu *sync.Mutex) error {
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
	err1 := g.Wait()
	if err1 != nil {
		return err1
	}
	warning("secrets", names, w, mu)
	return nil
}

// checkSA checks if there are user created SAs in the default namespace
func checkSA(items *corev1.ServiceAccountList, w *[]error, mu *sync.Mutex) error {
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
	err1 := g.Wait()
	if err1 != nil {
		return err1
	}
	warning("service account", names, w, mu)
	return nil
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *check) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var w []error
	var g errgroup.Group
	var mu sync.Mutex
	g.Go(func() error {
		return checkPods(objects.Pods, &w, &mu)
	})

	g.Go(func() error {
		return checkPodTemplates(objects.PodTemplates, &w, &mu)
	})

	g.Go(func() error {
		return checkPVCs(objects.PersistentVolumeClaims, &w, &mu)
	})

	g.Go(func() error {
		return checkConfigMaps(objects.ConfigMaps, &w, &mu)
	})

	g.Go(func() error {
		return checkQuotas(objects.ResourceQuotas, &w, &mu)
	})

	g.Go(func() error {
		return checkLimits(objects.LimitRanges, &w, &mu)
	})

	g.Go(func() error {
		return checkServices(objects.Services, &w, &mu)
	})

	g.Go(func() error {
		return checkSecrets(objects.Secrets, &w, &mu)
	})

	g.Go(func() error {
		return checkSA(objects.ServiceAccounts, &w, &mu)
	})

	return w, nil, g.Wait()
}
