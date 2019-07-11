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

package checks

import (
	"errors"
	"sync"

	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
)

// Run applies the filters and runs the resultant check list in parallel
func Run(client *kube.Client, checkFilter CheckFilter, diagnosticFilter DiagnosticFilter) ([]Diagnostic, error) {
	objects, err := client.FetchObjects()
	if err != nil {
		return nil, err
	}

	all, err := checkFilter.FilterChecks()
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, errors.New("No checks to run. Are you sure that you provided the right names for groups and checks?")
	}
	var diagnostics []Diagnostic
	var mu sync.Mutex
	var g errgroup.Group

	for _, check := range all {
		check := check
		g.Go(func() error {
			d, err := check.Run(objects)
			if err != nil {
				return err
			}
			mu.Lock()
			diagnostics = append(diagnostics, d...)
			mu.Unlock()
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return nil, err
	}
	diagnostics = filterEnabled(diagnostics)
	diagnostics = filterSeverity(diagnosticFilter.Severity, diagnostics)
	return diagnostics, err
}

func filterEnabled(diagnostics []Diagnostic) []Diagnostic {
	var ret []Diagnostic
	for _, d := range diagnostics {
		if IsEnabled(d.Check, d.Object) {
			ret = append(ret, d)
		}
	}
	return ret
}

func filterSeverity(level Severity, diagnostics []Diagnostic) []Diagnostic {
	if level == "" {
		return diagnostics
	}
	var ret []Diagnostic
	for _, d := range diagnostics {
		if d.Severity == level {
			ret = append(ret, d)
		}
	}
	return ret
}
