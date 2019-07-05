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
	"fmt"
)

// CheckFilter stores names of checks and groups that needs to be included or excluded while running checks
type CheckFilter struct {
	IncludeGroups []string
	ExcludeGroups []string
	IncludeChecks []string
	ExcludeChecks []string
}

// NewCheckFilter is a constructor to initialize an instance of CheckFilter
func NewCheckFilter(includeGroups, excludeGroups, includeChecks, excludeChecks []string) (CheckFilter, error) {
	if len(includeGroups) > 0 && len(excludeChecks) > 0 {
		return CheckFilter{}, fmt.Errorf("cannot specify both include and exclude group conditions")
	}
	if len(includeChecks) > 0 && len(excludeChecks) > 0 {
		return CheckFilter{}, fmt.Errorf("cannot specify both include and exclude check conditions")
	}
	return CheckFilter{
		IncludeGroups: includeGroups,
		ExcludeGroups: excludeGroups,
		IncludeChecks: includeChecks,
		ExcludeChecks: excludeChecks,
	}, nil
}

// FilterChecks filters all to return set of checks based on the CheckFilter
func (c CheckFilter) FilterChecks() ([]Check, error) {
	all, err := c.filterGroups()

	if err != nil {
		return nil, err
	}

	var ret []Check

	if len(c.IncludeChecks) > 0 {
		for _, check := range all {
			if Contains(c.IncludeChecks, check.Name()) {
				ret = append(ret, check)
			}
		}
		return ret, nil
	}

	if len(c.ExcludeChecks) > 0 {
		for _, check := range all {
			if !Contains(c.ExcludeChecks, check.Name()) {
				ret = append(ret, check)
			}
		}
		return ret, nil
	}

	return all, nil
}

func (c CheckFilter) filterGroups() ([]Check, error) {
	if len(c.IncludeGroups) > 0 {
		groups, err := GetGroups(c.IncludeGroups)
		return groups, err
	}

	if len(c.ExcludeGroups) > 0 {
		return getChecksNotInGroups(c.ExcludeGroups), nil
	}

	return List(), nil
}

func getChecksNotInGroups(groups []string) []Check {
	allGroups := ListGroups()
	var ret []Check
	for _, group := range allGroups {
		if !Contains(groups, group) {
			ret = append(ret, GetGroup(group)...)
		}
	}
	return ret
}
