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

## Contributing

Contributions are welcome, in the form of either issues or pull requests. Please
see the [contribution guidelines](CONTRIBUTING.md) for details.

## License

Copyright 2019 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
