package basic

import (
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&unusedCMCheck{})
}

type unusedCMCheck struct{}

type identifier struct {
	Name      string
	Namespace string
}

// Name returns a unique name for this check.
func (c *unusedCMCheck) Name() string {
	return "unused-config-map"
}

// Groups returns a list of group names this check should be part of.
func (c *unusedCMCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (c *unusedCMCheck) Description() string {
	return "Checks if there are unused config maps in the cluster"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (c *unusedCMCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic

	used, err := checkReferences(objects)
	if err != nil {
		return nil, err
	}

	for _, cm := range objects.ConfigMaps.Items {
		if _, ok := used[identifier{Name: cm.GetName(), Namespace: cm.GetNamespace()}]; !ok {
			cm := cm
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused config map",
				Kind:     checks.ConfigMap,
				Object:   &cm.ObjectMeta,
				Owners:   cm.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}

//checkReferences checks each pod for config map references in volumes and environment variables
func checkReferences(objects *kube.Objects) (map[identifier]bool, error) {
	used := make(map[identifier]bool)
	var mu sync.Mutex
	var g errgroup.Group
	for _, pod := range objects.Pods.Items {
		pod := pod
		namespace := pod.GetNamespace()
		g.Go(func() error {
			for _, volume := range pod.Spec.Volumes {
				cm := volume.VolumeSource.ConfigMap
				if cm != nil {
					mu.Lock()
					used[identifier{Name: cm.LocalObjectReference.Name, Namespace: namespace}] = true
					mu.Unlock()
				}
			}
			identifiers := checkEnvVars(pod.Spec.Containers, namespace)
			identifiers = append(identifiers, checkEnvVars(pod.Spec.InitContainers, namespace)...)
			mu.Lock()
			for _, i := range identifiers {
				used[i] = true
			}
			mu.Unlock()

			return nil
		})
	}

	return used, g.Wait()
}

// checkEnvVars checks for config map references in container environment variables
func checkEnvVars(containers []corev1.Container, namespace string) []identifier {
	var refs []identifier
	for _, container := range containers {
		for _, env := range container.EnvFrom {
			if env.ConfigMapRef != nil {
				refs = append(refs, identifier{Name: env.ConfigMapRef.LocalObjectReference.Name, Namespace: namespace})
			}
		}
	}
	return refs
}
