/*
Copyright 2019 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	// Load client-go authentication plugins
	"fmt"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)
// ObjectsFilter stores names of namespaces that needs to be included or excluded while running checks
type ObjectsFilter struct {
	IncludeNamespaces []string
	ExcludeNamespaces []string
}


// NewObjectsFilter is a constructor to initialize an instance of ObjectsFilter
func NewObjectsFilter(includeNamespaces, excludeNamespaces []string) (ObjectsFilter, error) {
	if len(includeNamespaces) > 0 && len(excludeNamespaces) > 0 {
		return ObjectsFilter{}, fmt.Errorf("cannot specify both include and exclude namespace conditions")
	}
	return ObjectsFilter{
		IncludeNamespaces: includeNamespaces,
		ExcludeNamespaces: excludeNamespaces,
	}, nil
}

// FilterChecks filters all to return set of checks based on the ObjectsFilter
func (f ObjectsFilter) Filter(objects *Objects) {
	if len(f.IncludeNamespaces) > 0 {
		var ps []corev1.Pod
		for _, p := range objects.Pods.Items  {
			if contains(f.IncludeNamespaces, p.Namespace) {
				ps = append(ps, p)
			}
		}
		objects.Pods.Items = ps

		var pts []corev1.PodTemplate
		for _, pt := range objects.PodTemplates.Items  {
			if contains(f.IncludeNamespaces, pt.Namespace) {
				pts = append(pts, pt)
			}
		}
		objects.PodTemplates.Items = pts

		var pvcs []corev1.PersistentVolumeClaim
		for _, pvc := range objects.PersistentVolumeClaims.Items {
			if contains(f.IncludeNamespaces, pvc.Namespace) {
				pvcs = append(pvcs, pvc)
			}
		}
		objects.PersistentVolumeClaims.Items = pvcs

		var cms []corev1.ConfigMap
		for _, cm := range objects.ConfigMaps.Items {
			if contains(f.IncludeNamespaces, cm.Namespace) {
				cms = append(cms, cm)
			}
		}
		objects.ConfigMaps.Items = cms

		var svcs []corev1.Service
		for _, svc := range objects.Services.Items {
			if contains(f.IncludeNamespaces, svc.Namespace) {
				svcs = append(svcs, svc)
			}
		}
		objects.Services.Items = svcs

		var scrts []corev1.Secret
		for _, scrt := range objects.Secrets.Items {
			if contains(f.IncludeNamespaces, scrt.Namespace) {
				scrts = append(scrts, scrt)
			}
		}
		objects.Secrets.Items = scrts

		var sas []corev1.ServiceAccount
		for _, sa := range objects.ServiceAccounts.Items {
			if contains(f.IncludeNamespaces, sa.Namespace) {
				sas = append(sas, sa)
			}
		}
		objects.ServiceAccounts.Items = sas

		var rqs []corev1.ResourceQuota
		for _, rq := range objects.ResourceQuotas.Items {
			if contains(f.IncludeNamespaces, rq.Namespace) {
				rqs = append(rqs, rq)
			}
		}
		objects.ResourceQuotas.Items = rqs

		var lrs []corev1.LimitRange
		for _, lr := range objects.LimitRanges.Items {
			if contains(f.IncludeNamespaces, lr.Namespace) {
				lrs = append(lrs, lr)
			}
		}
		objects.LimitRanges.Items = lrs

		return
	}

	if len(f.ExcludeNamespaces) > 0 {
		var ps []corev1.Pod
		for _, p := range objects.Pods.Items  {
			if !contains(f.ExcludeNamespaces, p.Namespace) {
				ps = append(ps, p)
			}
		}
		objects.Pods.Items = ps

		var pts []corev1.PodTemplate
		for _, pt := range objects.PodTemplates.Items  {
			if !contains(f.ExcludeNamespaces, pt.Namespace) {
				pts = append(pts, pt)
			}
		}
		objects.PodTemplates.Items = pts

		var pvcs []corev1.PersistentVolumeClaim
		for _, pvc := range objects.PersistentVolumeClaims.Items {
			if !contains(f.ExcludeNamespaces, pvc.Namespace) {
				pvcs = append(pvcs, pvc)
			}
		}
		objects.PersistentVolumeClaims.Items = pvcs

		var cms []corev1.ConfigMap
		for _, cm := range objects.ConfigMaps.Items {
			if !contains(f.ExcludeNamespaces, cm.Namespace) {
				cms = append(cms, cm)
			}
		}
		objects.ConfigMaps.Items = cms

		var svcs []corev1.Service
		for _, svc := range objects.Services.Items {
			if !contains(f.ExcludeNamespaces, svc.Namespace) {
				svcs = append(svcs, svc)
			}
		}
		objects.Services.Items = svcs

		var scrts []corev1.Secret
		for _, scrt := range objects.Secrets.Items {
			if !contains(f.ExcludeNamespaces, scrt.Namespace) {
				scrts = append(scrts, scrt)
			}
		}
		objects.Secrets.Items = scrts

		var sas []corev1.ServiceAccount
		for _, sa := range objects.ServiceAccounts.Items {
			if !contains(f.ExcludeNamespaces, sa.Namespace) {
				sas = append(sas, sa)
			}
		}
		objects.ServiceAccounts.Items = sas

		var rqs []corev1.ResourceQuota
		for _, rq := range objects.ResourceQuotas.Items {
			if !contains(f.ExcludeNamespaces, rq.Namespace) {
				rqs = append(rqs, rq)
			}
		}
		objects.ResourceQuotas.Items = rqs

		var lrs []corev1.LimitRange
		for _, lr := range objects.LimitRanges.Items {
			if !contains(f.ExcludeNamespaces, lr.Namespace) {
				lrs = append(lrs, lr)
			}
		}
		objects.LimitRanges.Items = lrs

		return
	}
}

