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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"plugin"
	"strings"
	"time"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/fatih/color"
	"github.com/urfave/cli"

	// Side-effect import to get all the checks registered.
	_ "github.com/digitalocean/clusterlint/checks/all"
)

const delimiter = ":"

func main() {
	app := cli.NewApp()
	app.Name = "clusterlint"
	app.Usage = "Linter for k8s objects from a live cluster"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "absolute path to the kubeconfig file",
		},
		cli.StringFlag{
			Name:  "context",
			Usage: "context for the kubernetes client. default: current context",
		},
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "configure timeout for the kubernetes client. default: 30s",
			Value: time.Second * 30,
		},
		cli.StringSliceFlag{
			Name:  "plugins",
			Usage: "paths of Go plugins to load containing local checks",
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
			Before: loadPlugins,
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
					Name:  "n, namespace",
					Usage: "run checks in specific namespace",
				},
				cli.StringFlag{
					Name:  "N, ignore-namespace",
					Usage: "run checks not in specific namespace",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "output format [text|json]. Default: text",
				},
				cli.StringFlag{
					Name:  "level, l",
					Usage: "Filter output messages based on severity [error|warning|suggestion]. Default: all",
				},
				cli.BoolFlag{
					Name:  "no-color",
					Usage: "Disable color output",
				},
			},
			Before: loadPlugins,
			Action: runChecks,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("failed: %v", err)
		os.Exit(1)
	}
}

func loadPlugins(c *cli.Context) error {
	plugins := c.GlobalStringSlice("plugins")
	for _, p := range plugins {
		_, err := plugin.Open(p)
		if err != nil {
			return err
		}
	}

	return nil
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
	var kubeconfigFilePaths []string

	if kubeconfig := c.GlobalString("kubeconfig"); kubeconfig != "" {
		kubeconfigFilePaths = []string{kubeconfig}
	} else if value := os.Getenv("KUBECONFIG"); value != "" {
		kubeconfigFilePaths = strings.Split(value, delimiter)
	}

	client, err := kube.NewClient(kube.WithMergedConfigFiles(kubeconfigFilePaths), kube.WithKubeContext(c.GlobalString("context")), kube.WithTimeout(c.GlobalDuration("timeout")))
	if err != nil {
		return err
	}

	filter, err := checks.NewCheckFilter(c.StringSlice("g"), c.StringSlice("G"), c.StringSlice("c"), c.StringSlice("C"))
	if err != nil {
		return err
	}

	diagnosticFilter := checks.DiagnosticFilter{Severity: checks.Severity(c.String("level"))}

	objectFilter, err := kube.NewObjectFilter(c.String("n"), c.String("N"))
	if err != nil {
		return err
	}

	output, err := checks.Run(context.Background(), client, filter, diagnosticFilter, objectFilter)
	if err != nil {
		return err
	}
	err = write(output, c)
	return err
}

func write(checkResult *checks.CheckResult, c *cli.Context) error {
	output := c.String("output")

	switch output {
	case "json":
		err := json.NewEncoder(os.Stdout).Encode(checkResult)
		if err != nil {
			return err
		}
	default:
		if c.Bool("no-color") {
			color.NoColor = true
		}
		e := color.New(color.FgRed)
		w := color.New(color.FgYellow)
		s := color.New(color.FgBlue)
		for _, d := range checkResult.Diagnostics {
			switch d.Severity {
			case checks.Error:
				e.Println(d)
			case checks.Warning:
				w.Println(d)
			case checks.Suggestion:
				s.Println(d)
			default:
				fmt.Println(d)
			}
		}
	}

	return nil
}
