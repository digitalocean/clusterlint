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

// FilterChecks filters all to return set of checks based on the CheckFilter
func (c CheckFilter) FilterChecks() ([]Check, error) {
	if len(c.IncludeChecks) > 0 && len(c.ExcludeChecks) > 0 {
		return nil, fmt.Errorf("cannot specify both c and C flags")
	}

	all, err := c.filterGroups()

	if err != nil {
		return nil, err
	}

	var ret []Check

	if len(c.IncludeChecks) > 0 {
		for _, check := range all {
			if c.contains(c.IncludeChecks, check.Name()) {
				ret = append(ret, check)
			}
		}
		return ret, nil
	} else if len(c.ExcludeChecks) > 0 {
		for _, check := range all {
			if !c.contains(c.ExcludeChecks, check.Name()) {
				ret = append(ret, check)
			}
		}
		return ret, nil
	} else {
		return all, nil
	}

}

func (c CheckFilter) filterGroups() ([]Check, error) {
	if len(c.IncludeGroups) > 0 && len(c.ExcludeGroups) > 0 {
		return nil, fmt.Errorf("cannot specify both g and G flags")
	}

	if len(c.IncludeGroups) > 0 {
		groups, err := GetGroups(c.IncludeGroups)
		return groups, err
	} else if len(c.ExcludeGroups) > 0 {
		return c.getChecksNotInGroups(c.ExcludeGroups), nil
	} else {
		return List(), nil
	}
}

func (c CheckFilter) getChecksNotInGroups(groups []string) []Check {
	allGroups := ListGroups()
	var ret []Check
	for _, group := range allGroups {
		if !c.contains(groups, group) {
			ret = append(ret, GetGroup(group)...)
		}
	}
	return ret
}

func (c CheckFilter) contains(list []string, name string) bool {
	for _, l := range list {
		if l == name {
			return true
		}
	}
	return false
}
