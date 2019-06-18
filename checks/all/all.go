// Package all imports all other check-containing packages, for the side-effect
// of having them registered in the check registry.
package all

import (
	_ "github.com/digitalocean/clusterlint/checks/basic"
	_ "github.com/digitalocean/clusterlint/checks/doks"
	_ "github.com/digitalocean/clusterlint/checks/noop"
	_ "github.com/digitalocean/clusterlint/checks/security"
)
