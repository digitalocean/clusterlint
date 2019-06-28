package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/urfave/cli"

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
					Name:  "g",
					Usage: "list all checks in groups `GROUP1, GROUP2`",
				},
				cli.StringSliceFlag{
					Name:  "G",
					Usage: "list all checks not in groups `GROUP1, GROUP2`",
				},
			},
			Action: checks.ListChecks,
		},
		{
			Name:  "run",
			Usage: "run all checks in the registry",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "g",
					Usage: "run all checks in groups `GROUP1, GROUP2`",
				},
				cli.StringSliceFlag{
					Name:  "G",
					Usage: "run all checks not in groups `GROUP1, GROUP2`",
				},
				cli.StringFlag{
					Name:  "name, n",
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
			Action: checks.RunChecks,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("failed: %v", err)
		os.Exit(1)
	}
}
