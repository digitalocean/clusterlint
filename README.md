# Clusterlint

Linter to check k8s API objects from a live cluster.

### Background

The idea for this tool was conceived to address some of the issues users face during upgrade of the cluster to a new kubernetes version.
This also documents some of the recommended practices to follow while writing the object configs.

### Install

```bash
go install ./cmd/clusterlint
```

The above command creates the `clusterlint` binary in `$GOPATH/bin`

### Usage

```bash
clusterlint list [options]  // list all checks available to the user
clusterlint run [options]  // run all or specific checks
```
