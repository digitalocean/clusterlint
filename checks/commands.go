package checks

import (
	"fmt"

	"github.com/urfave/cli"
)

// ListChecks lists the names and desc of all checks in the group if found
// lists all checks in the registry if group is not specified
func ListChecks(c *cli.Context) error {
	allChecks, err := filter(c)
	if err != nil {
		return err
	}
	for _, check := range allChecks {
		fmt.Printf("%s : %s\n", check.Name(), check.Description())
	}

	return nil
}

func filter(c *cli.Context) ([]Check, error) {
	whitelist := c.StringSlice("g")
	blacklist := c.StringSlice("G")
	if len(whitelist) > 0 && len(blacklist) > 0 {
		return nil, fmt.Errorf("cannot specify both g and G flags")
	}

	if len(whitelist) > 0 {
		return GetGroups(whitelist), nil
	} else if len(blacklist) > 0 {
		return getChecksNotInGroups(blacklist), nil
	} else {
		return List(), nil
	}
}

// getChecksInGroups retrieves all checks within specified set of groups
// returns all checks in the registry if `groups` is unspecified
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

func contains(groups []string, group string) bool {
	for _, g := range groups {
		if g == group {
			return true
		}
	}
	return false
}
