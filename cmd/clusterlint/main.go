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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	// Side-effect import to get all the checks registered.
	_ "github.com/digitalocean/clusterlint/checks/all"
)

func main() {
	app := cli.NewApp()
	app.Name = "clusterlint"
	app.Usage = "Linter for k8sobjects from a live cluster"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Value: filepath.Join(os.Getenv("HOME"), ".kube", "config"),
			Usage: "absolute path to the kubeconfig file",
		},
		cli.StringFlag{
			Name:  "context",
			Usage: "context for the kubernetes client. default: current context",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "list all checks in the registry",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "g, groups",
					Usage: "list all checks in groups `GROUP1, GROUP2`",
				},
				cli.StringSliceFlag{
					Name:  "G, ignore-groups",
					Usage: "list all checks not in groups `GROUP1, GROUP2`",
				},
			},
			Action: listChecks,
		},
		{
			Name:  "run",
			Usage: "run all checks in the registry",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "g, groups",
					Usage: "run all checks in groups `GROUP1, GROUP2`",
				},
				cli.StringSliceFlag{
					Name:  "G, ignore-groups",
					Usage: "run all checks not in groups `GROUP1, GROUP2`",
				},
				cli.StringSliceFlag{
					Name:  "c, checks",
					Usage: "run a specific check",
				},
				cli.StringSliceFlag{
					Name:  "C, ignore-checks",
					Usage: "run a specific check",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "output format [text|json]. Default: text",
				},
				cli.StringFlag{
					Name:  "level, l",
					Usage: "Filter output messages based on severity [error|warning|suggestion]. Default: all",
				},
			},
			Action: runChecks,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("failed: %v", err)
		os.Exit(1)
	}
}

// listChecks lists the names and desc of all checks in the group if found
// lists all checks in the registry if group is not specified
func listChecks(c *cli.Context) error {
	filter, err := checks.NewCheckFilter(c.StringSlice("g"), c.StringSlice("G"), nil, nil)
	if err != nil {
		return err
	}
	allChecks, err := filter.FilterChecks()
	if err != nil {
		return err
	}

	for _, check := range allChecks {
		fmt.Printf("%s : %s\n", check.Name(), check.Description())
	}

	return nil
}

// runChecks runs all the checks based on the flags passed.
func runChecks(c *cli.Context) error {
	client, err := kube.NewClient(c.GlobalString("kubeconfig"), c.GlobalString("context"))
	if err != nil {
		return err
	}

	objects, err := client.FetchObjects()
	if err != nil {
		return err
	}

	return run(objects, c)
}

func run(objects *kube.Objects, c *cli.Context) error {
	filter, err := checks.NewCheckFilter(c.StringSlice("g"), c.StringSlice("G"), c.StringSlice("c"), c.StringSlice("C"))
	if err != nil {
		return err
	}

	all, err := filter.FilterChecks()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("No checks to run. Are you sure that you provided the right names for groups and checks?")
	}
	var diagnostics []checks.Diagnostic
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

func write(diagnostics []checks.Diagnostic, c *cli.Context) error {
	output := c.String("output")
	level := checks.Severity(c.String("level"))
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

func filterSeverity(level checks.Severity, diagnostics []checks.Diagnostic) []checks.Diagnostic {
	if level == "" {
		return diagnostics
	}
	var ret []checks.Diagnostic
	for _, d := range diagnostics {
		if d.Severity == level {
			ret = append(ret, d)
		}
	}
	return ret
}
