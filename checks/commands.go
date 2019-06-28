package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/digitalocean/clusterlint/kube"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

// ListChecks lists the names and desc of all checks in the group if found
// lists all checks in the registry if group is not specified
func ListChecks(c *cli.Context) error {
	allChecks, err := filterGroups(c)
	if err != nil {
		return err
	}
	for _, check := range allChecks {
		fmt.Printf("%s : %s\n", check.Name(), check.Description())
	}

	return nil
}

// RunChecks runs all the checks based on the flags passed.
func RunChecks(c *cli.Context) error {
	client, err := kube.NewClient(c.GlobalString("kubeconfig"), c.GlobalString("context"))
	if err != nil {
		return err
	}

	objects, err := client.FetchObjects()
	if err != nil {
		return err
	}

	return runChecks(objects, c)
}

func runChecks(objects *kube.Objects, c *cli.Context) error {
	all, err := filterGroups(c)
	if err != nil {
		return err
	}
	all, err = filterChecks(all, c)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("No checks to run. Are you sure that you provided the right names for groups and checks?")
	}
	var diagnostics []Diagnostic
	var mu sync.Mutex
	var g errgroup.Group

	for _, check := range all {
		check := check
		g.Go(func() error {
			fmt.Println("Running check: ", check.Name())
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
	write(diagnostics, c)

	return err
}

func write(diagnostics []Diagnostic, c *cli.Context) error {
	output := c.String("output")
	level := Severity(c.String("level"))
	filtered := filterSeverity(level, diagnostics)
	switch output {
	case "json":
		err := json.NewEncoder(os.Stdout).Encode(filtered)
		if err != nil {
			return err
		}
	default:
		for _, diagnostic := range filtered {
			fmt.Printf("%s\n", diagnostic)
		}
	}

	return nil
}

func filterSeverity(level Severity, diagnostics []Diagnostic) []Diagnostic {
	if level == "" {
		return diagnostics
	}
	var filtered []Diagnostic
	for _, d := range diagnostics {
		if d.Severity == level {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func filterGroups(c *cli.Context) ([]Check, error) {
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

func filterChecks(all []Check, c *cli.Context) ([]Check, error) {
	whitelist := c.StringSlice("c")
	blacklist := c.StringSlice("C")

	if len(whitelist) > 0 && len(blacklist) > 0 {
		return nil, fmt.Errorf("cannot specify both c and C flags")
	}

	var ret []Check

	if len(whitelist) > 0 {
		for _, c := range all {
			if contains(whitelist, c.Name()) {
				ret = append(ret, c)
			}
		}
		return ret, nil
	} else if len(blacklist) > 0 {
		for _, c := range all {
			if !contains(blacklist, c.Name()) {
				ret = append(ret, c)
			}
		}
		return ret, nil
	} else {
		return all, nil
	}

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
