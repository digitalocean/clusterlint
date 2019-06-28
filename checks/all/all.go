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
)
