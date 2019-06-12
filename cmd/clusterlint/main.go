package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/digitalocean/clusterlint/checks"
	_ "github.com/digitalocean/clusterlint/checks/noop"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const all string = ""

type KubernetesAPI struct {
	Client kubernetes.Interface
}

func main() {
	app := cli.NewApp()
	app.Name = "clusterlint"
	app.Usage = "Linter for k8sobjects from a live cluster"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Print help docs")
		return nil
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
			Action: func(c *cli.Context) error {
				group := c.String("group")
				listChecks(group)
				return nil
			},
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
			Action: func(c *cli.Context) error {
				group := c.String("group")
				name := c.String("name")
				runChecks(group, name)
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic("boo")
	}
}

// listChecks lists the names and desc of all checks in the group if found
// lists all checks in the registry if group is not specified
func listChecks(group string) {
	allChecks := getChecks(group)
	for _, check := range allChecks {
		fmt.Printf("%s : %s\n", check.Name(), check.Description())
	}
}

func runChecks(group, name string) {
	api := &KubernetesAPI{Client: buildClient()}
	objects := api.fetch()
	if "" == name {
		runChecksForGroup(group, objects)
	} else {
		runCheck(name, objects)
	}
}

// runChecksForGroup runs all checks in the specified group if found
// runs all checks in the registry if group is not specified
func runChecksForGroup(group string, objects *kube.Objects) {
	allChecks := getChecks(group)
	var warnings, errors []error
	var mu sync.Mutex
	var g errgroup.Group

	for _, check := range allChecks {
		check := check
		g.Go(func() error {
			log.Println("Running check: ", check.Name())
			w, e, err := check.Run(objects)
			if err != nil {
				return err
			}
			mu.Lock()
			warnings = append(warnings, w...)
			errors = append(errors, e...)
			mu.Unlock()
			return nil
		})
	}
	err := g.Wait()
	showErrorsAndWarnings(warnings, errors)
	if err != nil {
		handleError(err)
	}
}

// runCheck runs a specific check identified by check.Name()
// errors out if the check is not found in the registry
func runCheck(name string, objects *kube.Objects) {
	check, err := checks.Get(name)
	if err != nil {
		handleError(err)
	}

	log.Println("Running check: ", name)
	warnings, errors, err := check.Run(objects)
	showErrorsAndWarnings(warnings, errors)
	handleError(err)
}

//showErrorsAndWarnings displays all the errors and warnings returned by checks
func showErrorsAndWarnings(warnings, errors []error) {
	for _, warning := range warnings {
		log.Println("Warning: ", warning.Error())
	}
	for _, err := range errors {
		log.Println("Error: ", err.Error())
	}
}

// getChecks retrieves all checks within given group
// returns all checks in the registry if group in unspecified
func getChecks(group string) []checks.Check {
	if group == "" {
		return checks.List()
	}
	return checks.GetGroup(group)
}

// fetch initializes a kube.Objects instance with live cluster objects
// Currently limited to core k8s API objects
func (k KubernetesAPI) fetch() *kube.Objects {
	client := k.Client.CoreV1()
	opts := metav1.ListOptions{}
	objects := &kube.Objects{}
	var err error

	objects.Nodes, err = client.Nodes().List(opts)
	handleError(err)

	objects.PersistentVolumes, err = client.PersistentVolumes().List(opts)
	handleError(err)

	objects.ComponentStatuses, err = client.ComponentStatuses().List(opts)
	handleError(err)

	objects.Pods, err = client.Pods(all).List(opts)
	handleError(err)

	objects.PodTemplates, err = client.PodTemplates(all).List(opts)
	handleError(err)

	objects.PersistentVolumeClaims, err = client.PersistentVolumeClaims(all).List(opts)
	handleError(err)

	objects.ConfigMaps, err = client.ConfigMaps(all).List(opts)
	handleError(err)

	objects.Secrets, err = client.Secrets(all).List(opts)
	handleError(err)

	objects.Services, err = client.Services(all).List(opts)
	handleError(err)

	objects.ServiceAccounts, err = client.ServiceAccounts(all).List(opts)
	handleError(err)

	objects.ResourceQuotas, err = client.ResourceQuotas(all).List(opts)
	handleError(err)

	objects.LimitRanges, err = client.LimitRanges(all).List(opts)
	handleError(err)

	return objects
}

// buildClient parses command line args and initializes the k8s client
// to invoke APIs
func buildClient() kubernetes.Interface {
	k8sconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "absolute path to the kubeconfig file")
	context := flag.String("context", "", "context for the kubernetes client. default: current context")
	flag.Parse()

	var config *rest.Config
	if "" != *context {
		config, _ = buildConfigFromFlags(context, k8sconfig)
	} else {
		config, _ = clientcmd.BuildConfigFromFlags("", *k8sconfig)
	}

	client := kubernetes.NewForConfigOrDie(config)
	return client
}

// buildConfigFromFlags initializes client config with given context
func buildConfigFromFlags(context, kubeconfigPath *string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: *context,
		}).ClientConfig()
}

// handleError logs error to stdout and exits
func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
