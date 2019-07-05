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

package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/digitalocean/clusterlint/checks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	checks.Register(&compatibilityCheck{})
}

type compatibilityCheck struct{}

// Name returns a unique name for this check.
func (v *compatibilityCheck) Name() string {
	return "version-compatibility"
}

// Groups returns a list of group names this check should be part of.
func (v *compatibilityCheck) Groups() []string {
	return []string{"versions"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (v *compatibilityCheck) Description() string {
	return "Checks if API version of the objects is supported in the target Kubernetes version"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (v *compatibilityCheck) Run(data *checks.CheckData) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	currentVersion, _ := data.Client.Get().Discovery().ServerVersion()

	var current, target interface{}
	path, _ := filepath.Abs("data/" + currentVersion.Major + "." + currentVersion.Minor + ".json")
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(bytes.NewReader(file)).Decode(&current)

	if err != nil {
		return nil, err
	}
	path, _ = filepath.Abs("data/" + data.TargetVersion + ".json")
	file, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(bytes.NewReader(file)).Decode(&target)

	if err != nil {
		return nil, err
	}
	// groupClients := data.Client.GetGroupClients()
	r, e := data.Client.Get().Discovery().ServerPreferredResources()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(r)

	extensionsResources := filterResourcesFor(r, "extensions/v1beta1")
	appsResources := filterResourcesFor(r, "apps/v1")

	fmt.Println(extensionsResources)
	fmt.Println("-----------------------------------------------------------")
	fmt.Println(appsResources)
	fmt.Println("-----------------------------------------------------------")

	client := data.Client.Get().Discovery().RESTClient()
	resp, err := client.Get().Name("deployments").Do().Get()
	// client := groupClients["apps/v1beta1"]
	// resp, err := client.Get().Name("deployments").Do().Get()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)

	// for group, c := range current.(map[string]interface{}) {
	// 	t, ok := target.(map[string]interface{})[group]
	// 	if !ok {
	// 		return nil, fmt.Errorf("Group not found in target version")
	// 	}
	// 	currentSupported := c.(map[string][]string)["supported_resources"]
	// 	targetSupported := t.(map[string][]string)["supported_resources"]
	// 	unsupportedResources := findUnsupported(currentSupported, targetSupported)
	// 	client := groupClients[group]
	//
	// 	for _, kind := range unsupportedResources {
	// 		switch kind {
	// 		case checks.Deployment:
	// 			resp, err := client.Get().Name("deployments").Do().Get()
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			for _, d := range resp.Items {
	// 				diagnostic := checks.Diagnostic{
	// 					Check:    v.Name(),
	// 					Severity: checks.Error,
	// 					Message:  fmt.Sprintf("Group version `%s` not supported in target kubernetes version `%s`", group, data.TargetVersion),
	// 					Kind:     checks.Deployment,
	// 					Object:   &d.ObjectMeta,
	// 					Owners:   d.ObjectMeta.GetOwnerReferences(),
	// 				}
	// 				diagnostics = append(diagnostics, diagnostic)
	// 			}
	// 		case checks.DaemonSet:
	// 			for _, d := range client.DaemonSets(corev1.NamespaceAll).List(opts).Items {
	// 				diagnostic := checks.Diagnostic{
	// 					Check:    v.Name(),
	// 					Severity: checks.Error,
	// 					Message:  fmt.Sprintf("Group version `%s` not supported in target kubernetes version `%s`", group, data.TargetVersion),
	// 					Kind:     checks.DaemonSet,
	// 					Object:   &d.ObjectMeta,
	// 					Owners:   d.ObjectMeta.GetOwnerReferences(),
	// 				}
	// 				diagnostics = append(diagnostics, diagnostic)
	// 			}
	// 		case checks.StatefulSet:
	// 			for _, s := range client.StatefulSets(corev1.NamespaceAll).List(opts).Items {
	// 				diagnostic := checks.Diagnostic{
	// 					Check:    v.Name(),
	// 					Severity: checks.Error,
	// 					Message:  fmt.Sprintf("Group version `%s` not supported in target kubernetes version `%s`", group, data.TargetVersion),
	// 					Kind:     checks.DaemonSet,
	// 					Object:   &s.ObjectMeta,
	// 					Owners:   s.ObjectMeta.GetOwnerReferences(),
	// 				}
	// 				diagnostics = append(diagnostics, diagnostic)
	// 			}
	// 		case checks.Ingress:
	// 			for _, i := range client.Ingress(corev1.NamespaceAll).List(opts).Items {
	// 				diagnostic := checks.Diagnostic{
	// 					Check:    v.Name(),
	// 					Severity: checks.Error,
	// 					Message:  fmt.Sprintf("Group version `%s` not supported in target kubernetes version `%s`", group, data.TargetVersion),
	// 					Kind:     checks.DaemonSet,
	// 					Object:   &i.ObjectMeta,
	// 					Owners:   i.ObjectMeta.GetOwnerReferences(),
	// 				}
	// 				diagnostics = append(diagnostics, diagnostic)
	// 			}
	//
	// 		}
	// 	}
	// }

	return diagnostics, nil
}

func filterResourcesFor(resources []*metav1.APIResourceList, group string) []metav1.APIResource {
	var groupResources []metav1.APIResource
	for _, r := range resources {
		if r.GroupVersion == group {
			groupResources = append(groupResources, r.APIResources...)
			break
		}
	}
	return groupResources
}

func findUnsupported(current, target []string) []checks.Kind {
	var unsupported []checks.Kind
	for _, resource := range current {
		if !checks.Contains(target, resource) {
			unsupported = append(unsupported, checks.Kind(resource))
		}
	}
	return unsupported
}
