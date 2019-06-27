package basic

import (
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&unusedSecretCheck{})
}

type unusedSecretCheck struct{}

type identifier struct {
	Name      string
	Namespace string
}

// Name returns a unique name for this check.
func (s *unusedSecretCheck) Name() string {
	return "unused-secret"
}

// Groups returns a list of group names this check should be part of.
func (s *unusedSecretCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (s *unusedSecretCheck) Description() string {
	return "Checks if there are unused secrets in the cluster. Ignores service account tokens"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (s *unusedSecretCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	used := make(map[identifier]bool)
	err := checkReferences(objects, &used)
	if err != nil {
		return nil, err
	}

	for _, secret := range filter(objects.Secrets.Items) {
		if _, ok := used[identifier{Name: secret.GetName(), Namespace: secret.GetNamespace()}]; !ok {
			secret := secret
			d := checks.Diagnostic{
				Severity: checks.Warning,
				Message:  "Unused secret",
				Kind:     checks.Secret,
				Object:   &secret.ObjectMeta,
				Owners:   secret.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}

//checkReferences checks each pod for config map references in volumes and environment variables
func checkReferences(objects *kube.Objects, used *map[identifier]bool) error {
	var mu sync.Mutex
	var g errgroup.Group
	for _, pod := range objects.Pods.Items {
		pod := pod
		namespace := pod.GetNamespace()
		g.Go(func() error {
			for _, volume := range pod.Spec.Volumes {
				s := volume.VolumeSource.Secret
				if s != nil {
					mu.Lock()
					(*used)[identifier{Name: s.SecretName, Namespace: namespace}] = true
					mu.Unlock()
				}
			}
			for _, imageSecret := range pod.Spec.ImagePullSecrets {
				mu.Lock()
				(*used)[identifier{Name: imageSecret.Name, Namespace: namespace}] = true
				mu.Unlock()
			}
			checkEnvVars(pod.Spec.Containers, used, namespace, &mu)
			checkEnvVars(pod.Spec.InitContainers, used, namespace, &mu)

			return nil
		})
	}

	return g.Wait()
}

// checkEnvVars checks for config map references in container environment variables
func checkEnvVars(containers []corev1.Container, used *map[identifier]bool, namespace string, mu *sync.Mutex) {
	for _, container := range containers {
		for _, env := range container.EnvFrom {
			if env.SecretRef != nil {
				mu.Lock()
				(*used)[identifier{Name: env.SecretRef.LocalObjectReference.Name, Namespace: namespace}] = true
				mu.Unlock()
			}
		}
	}
}

// filter returns Secrets that are not of type `kubernetes.io/service-account-token`
func filter(secrets []corev1.Secret) []corev1.Secret {
	var filtered []corev1.Secret
	for _, secret := range secrets {
		if secret.Type != corev1.SecretTypeServiceAccountToken {
			filtered = append(filtered, secret)
		}
	}
	return filtered
}
