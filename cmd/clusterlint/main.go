package main

import (
	"encoding/json"
	"log"
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
				cli.StringFlag{
					Name:  "group, g",
					Usage: "list all checks in group `GROUP`",
				},
			},
			Action: listChecks,
		},
		{
			Name:  "run",
			Usage: "run all checks in the registry",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "group, g",
					Usage: "run all checks in group `GROUP`",
				},
				cli.StringFlag{
					Name:  "name, n",
					Usage: "run a specific check",
				},
			},
			Action: runChecks,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Printf("failed: %v", err)
		os.Exit(1)
	}
}

// listChecks lists the names and desc of all checks in the group if found
// lists all checks in the registry if group is not specified
func listChecks(c *cli.Context) error {
	group := c.String("group")
	allChecks := getChecks(group)
	for _, check := range allChecks {
		log.Printf("%s : %s\n", check.Name(), check.Description())
	}

	return nil
}

func runChecks(c *cli.Context) error {
	group := c.String("group")
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
		return runChecksForGroup(group, objects)
	}
	return runCheck(name, objects)
}

// runChecksForGroup runs all checks in the specified group if found
// runs all checks in the registry if group is not specified
func runChecksForGroup(group string, objects *kube.Objects) error {
	allChecks := getChecks(group)
	var diagnostics []checks.Diagnostic
	var mu sync.Mutex
	var g errgroup.Group

	for _, check := range allChecks {
		check := check
		g.Go(func() error {
			log.Println("Running check: ", check.Name())
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
	err := g.Wait()
	showDiagnostics(diagnostics)

	return err
}

// runCheck runs a specific check identified by check.Name()
// errors out if the check is not found in the registry
func runCheck(name string, objects *kube.Objects) error {
	check, err := checks.Get(name)
	if err != nil {
		return err
	}

	log.Println("Running check: ", name)
	diagnostics, err := check.Run(objects)
	if err != nil {
		return err
	}
	return showDiagnostics(diagnostics)
}

// showErrorsAndWarnings displays all the errors and warnings returned by checks
func showDiagnostics(diagnostics []checks.Diagnostic) error {
	resp, err := json.Marshal(diagnostics)
	if err != nil {
		return err
	}
	log.Println(string(resp))
	return nil
}

// getChecks retrieves all checks within given group
// returns all checks in the registry if group in unspecified
func getChecks(group string) []checks.Check {
	if group == "" {
		return checks.List()
	}
	return checks.GetGroup(group)
}
