package basic

import (
	"fmt"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
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

func foo(name string, obj runtime.Object, w []error, e []error, mu *sync.Mutex) error {
	var r errgroup.Group
	for _, objects := range obj {
		objects := objects
		r.Go(func() error {
			if "default" == objects.GetNamespace() {
				mu.Lock()
				w = append(w, fmt.Errorf("Pod '%s' is in the default namespace", objects.GetName()))
				mu.Unlock()
			}
			return nil
		})
	}
	err1 := r.Wait()
	return err1
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (nc *check) Run(objects *kube.Objects) (warnings []error, errors []error, err error) {
	var w, e []error
	var g errgroup.Group

	var mu sync.Mutex

	g.Go(func() (err error) {
		foo("Pod", objects.Pods.Items, w, e, &mu)
		return nil
	})
	g.Go(func() (err error) {
		var r errgroup.Group
		for _, podTemplate := range objects.PodTemplates.Items {
			podTemplate := podTemplate
			r.Go(func() error {
				if "default" == podTemplate.GetNamespace() {
					mu.Lock()
					w = append(w, fmt.Errorf("Pod Template '%s' is in the default namespace", podTemplate.GetName()))
					mu.Unlock()
				}
				return nil
			})
		}
		err1 := r.Wait()
		if err1 != nil {
			return err1
		}
		return nil
	})
	// g.Go(func() (err error) {
	// 	objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.ConfigMaps, err = client.ConfigMaps(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.Secrets, err = client.Secrets(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.Services, err = client.Services(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.ServiceAccounts, err = client.ServiceAccounts(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.ResourceQuotas, err = client.ResourceQuotas(all).List(opts)
	// 	return
	// })
	// g.Go(func() (err error) {
	// 	objects.LimitRanges, err = client.LimitRanges(all).List(opts)
	// 	return
	// })

	err2 := g.Wait()
	return w, e, err2
}
