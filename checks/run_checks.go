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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"runtime/debug"

	"github.com/digitalocean/clusterlint/kube"
	"golang.org/x/sync/errgroup"
)

// Run applies the filters and runs the resultant check list in parallel
func Run(ctx context.Context, client *kube.Client, checkFilter CheckFilter, diagnosticFilter DiagnosticFilter, objectFilter kube.ObjectFilter) (*CheckResult, error) {
	objects, err := client.FetchObjects(ctx, objectFilter)
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
	checkDuration := make(map[string]time.Duration)
	for _, check := range all {
		check := check
		g.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("Recovered from panic in check '%s': %v", check.Name(), string(debug.Stack()))
				}
			}()
			start := time.Now()
			d, err := check.Run(objects)
			elapsed := time.Since(start)
			if err != nil {
				return err
			}
			mu.Lock()
			// Fill in the check names for the diagnostics. Doing this here
			// absolves checks of needing to do it and also ensures they're
			// consistent.
			for i := 0; i < len(d); i++ {
				d[i].Check = check.Name()
			}
			diagnostics = append(diagnostics, d...)
			checkDuration[check.Name()] = elapsed
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
	CheckResult := &CheckResult{Diagnostics: diagnostics, Durations: checkDuration}
	return CheckResult, err
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

// CheckResult is the output returned by the Run function
type CheckResult struct {
	Diagnostics []Diagnostic
	Durations   map[string]time.Duration
}
