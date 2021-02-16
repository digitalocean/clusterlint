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

// Package all imports all other check-containing packages, for the side-effect
// of having them registered in the check registry.
package all

import (
	// Side-effect import to get all the checks in basic package registered.
	_ "github.com/digitalocean/clusterlint/checks/basic"
	// Side-effect import to get all the checks in doks package registered.
	_ "github.com/digitalocean/clusterlint/checks/doks"
	// Side-effect import to get all the checks in noop package registered.
	_ "github.com/digitalocean/clusterlint/checks/noop"
	// Side-effect import to get all the checks in security package registered.
	_ "github.com/digitalocean/clusterlint/checks/security"
	// Side-effect import to get all the checks in containerd package registered.
	_ "github.com/digitalocean/clusterlint/checks/containerd"
)
