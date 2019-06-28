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
	allChecks, err := filter(c)
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
	name := c.String("name")

	client, err := kube.NewClient(c.GlobalString("kubeconfig"), c.GlobalString("context"))
	if err != nil {
		return err
	}

	objects, err := client.FetchObjects()
	if err != nil {
		return err
	}

	if name == "" {
		return runChecksForGroups(objects, c)
	}
	return runCheck(name, objects, c)
}

// runChecksForGroup runs all checks in the specified group if found
// runs all checks in the registry if group is not specified
func runChecksForGroups(objects *kube.Objects, c *cli.Context) error {
	allChecks, err := filter(c)
	if err != nil {
		return err
	}
	var diagnostics []Diagnostic
	var mu sync.Mutex
	var g errgroup.Group

	for _, check := range allChecks {
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
	showDiagnostics(diagnostics, c)

	return err
}

// runCheck runs a specific check identified by check.Name()
// errors out if the check is not found in the registry
func runCheck(name string, objects *kube.Objects, c *cli.Context) error {
	check, err := Get(name)
	if err != nil {
		return err
	}

	fmt.Println("Running check: ", name)
	diagnostics, err := check.Run(objects)
	if err != nil {
		return err
	}
	return showDiagnostics(diagnostics, c)
}

// showErrorsAndWarnings displays all the errors and warnings returned by checks
func showDiagnostics(diagnostics []Diagnostic, c *cli.Context) error {
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

// filterSeverity uses level to filter diagnostics to show to user. If level is blank, returns all diagnostics
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
