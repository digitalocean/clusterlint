# Example Check Plugin

This directory contains an example of a check plugin for clusterlint. Plugins
can be used to implement checks that are not appropriate for addition to
clusterlint itself for whatever reason - e.g., because they encode a best
practice that is highly specific to a particular organization.

## Building

Build the plugin as a Go plugin:

```console
$ go build -buildmode=plugin github.com/digitalocean/clusterlint/example-plugin
```

You should end up with a file called `example-plugin.so` in your working
directory.

## Usage

You can then use the plugin by loading it into clusterlint at runtime:

```console
$ clusterlint --plugins=./example-plugin.so run -c example-plugin
[suggestion] kube-system/pod/kubelet-rubber-stamp-f6756bc78-6sl9r: You probably don't want to run the example plugin.
```

The example plugin produces a suggestion for each pod running in the cluster,
just to show what a plugin can do.

## Troubleshooting

The easiest problem to hit with plugins is trying to use a plugin built against
a different version of the clusterlint codebase than the clusterlint binary
you're using. In this case, you'll get a message like:

```
plugin.Open("./example-plugin"): plugin was built with a different version of package github.com/digitalocean/clusterlint/kube
```

We recommend using go module versioning to ensure you're building your plugin
against code from the clusterlint release you're using.
