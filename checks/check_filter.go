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
			if contains(c.IncludeChecks, check.Name()) {
				ret = append(ret, check)
			}
		}
		return ret, nil
	}

	if len(c.ExcludeChecks) > 0 {
		for _, check := range all {
			if !contains(c.ExcludeChecks, check.Name()) {
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
		if !contains(groups, group) {
			ret = append(ret, GetGroup(group)...)
		}
	}
	return ret
}

func contains(list []string, name string) bool {
	for _, l := range list {
		if l == name {
			return true
		}
	}
	return false
}
