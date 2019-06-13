package basic

import (
	"fmt"
	"strings"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func checkNamespace(k8stype string, item metav1.ObjectMeta, w *[]error, mu *sync.Mutex) {
	if "default" == item.GetNamespace() {
		mu.Lock()
		*w = append(*w, fmt.Errorf("%s '%s' is in the default namespace", k8stype, item.GetName()))
		mu.Unlock()
	}
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *check) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var w, e []error
	var g errgroup.Group
	var mu sync.Mutex
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, pod := range objects.Pods.Items {
			pod := pod
			r.Go(func() error {
				checkNamespace("Pod", pod.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, podTemplate := range objects.PodTemplates.Items {
			podTemplate := podTemplate
			r.Go(func() error {
				checkNamespace("Pod Template", podTemplate.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, claim := range objects.PersistentVolumeClaims.Items {
			claim := claim
			r.Go(func() error {
				checkNamespace("Persistent Volume Claim", claim.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, configMap := range objects.ConfigMaps.Items {
			configMap := configMap
			r.Go(func() error {
				checkNamespace("Config Map", configMap.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})

	g.Go(func() (err error) {
		var r errgroup.Group
		for _, quota := range objects.ResourceQuotas.Items {
			quota := quota
			r.Go(func() error {
				checkNamespace("Resource quota", quota.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, limit := range objects.LimitRanges.Items {
			limit := limit
			r.Go(func() error {
				checkNamespace("Limit Range", limit.ObjectMeta, &w, &mu)
				return nil
			})
		}
		err1 := r.Wait()
		return err1
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		var services []string
		var serviceMu sync.Mutex
		for _, service := range objects.Services.Items {
			service := service
			r.Go(func() error {
				if "default" == service.GetNamespace() {
					serviceMu.Lock()
					services = append(services, service.GetName())
					serviceMu.Unlock()
				}
				return nil
			})
		}
		err1 := r.Wait()
		if err1 != nil {
			return err1
		}
		if len(services) > 1 {
			mu.Lock()
			w = append(w, fmt.Errorf("There are non-default secrets defined in the default namespace: %s.", strings.Join(services, ",")))
			mu.Unlock()
		}
		return nil
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		var secrets []string
		var secretMu sync.Mutex
		for _, secret := range objects.Secrets.Items {
			secret := secret
			r.Go(func() error {
				if "default" == secret.GetNamespace() {
					secretMu.Lock()
					secrets = append(secrets, secret.GetName())
					secretMu.Unlock()
				}
				return nil
			})
		}
		err1 := r.Wait()
		if err1 != nil {
			return err1
		}
		if len(secrets) > 1 {
			mu.Lock()
			w = append(w, fmt.Errorf("There are non-default secrets defined in the default namespace: %s.", strings.Join(secrets, ",")))
			mu.Unlock()
		}

		return nil
	})

	g.Go(func() (err error) {
		var r errgroup.Group
		var sas []string
		var saMu sync.Mutex
		for _, sa := range objects.ServiceAccounts.Items {
			sa := sa
			r.Go(func() error {
				if "default" == sa.GetNamespace() {
					saMu.Lock()
					sas = append(sas, sa.GetName())
					saMu.Unlock()
				}
				return nil
			})
		}
		err1 := r.Wait()
		if err1 != nil {
			return err1
		}
		if len(sas) > 1 {
			mu.Lock()
			w = append(w, fmt.Errorf("There are non-default service accounts defined in the default namespace: %s.", strings.Join(sas, ",")))
			mu.Unlock()
		}

		return nil
	})

	err2 := g.Wait()
	return w, e, err2
}
