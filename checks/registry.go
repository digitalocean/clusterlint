package checks

import (
	"errors"
	"fmt"
	"sync"
)

type checkRegistry struct {
	mu     sync.RWMutex
	checks map[string]Check
	groups map[string][]Check
}

var (
	registry *checkRegistry
	initOnce sync.Once
)

// Register registers a check. This should be called from each check
// implementation's init().
func Register(check Check) error {
	initOnce.Do(func() {
		registry = &checkRegistry{
			checks: make(map[string]Check),
			groups: make(map[string][]Check),
		}
	})

	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := check.Name()
	if name == "" {
		return errors.New("checks must have non-empty names")
	}
	if _, ok := registry.checks[name]; ok {
		return fmt.Errorf("check named %q already exists", name)
	}

	registry.checks[name] = check
	for _, group := range check.Groups() {
		registry.groups[group] = append(registry.groups[group], check)
	}

	return nil
}

// List returns all the registered checks.
func List() []Check {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	ret := make([]Check, 0, len(registry.checks))
	for _, check := range registry.checks {
		ret = append(ret, check)
	}

	return ret
}

// ListGroups returns the names of all registered check groups.
func ListGroups() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	ret := make([]string, 0, len(registry.groups))
	for group := range registry.groups {
		ret = append(ret, group)
	}

	return ret
}

// GetGroup returns the checks in a particular group.
func GetGroup(name string) []Check {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	ret := make([]Check, 0, len(registry.groups[name]))
	for _, check := range registry.groups[name] {
		ret = append(ret, check)
	}

	return ret
}

// GetGroups returns checks that belong to any of the specified group.
func GetGroups(groups []string) []Check {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	var ret []Check
	for _, group := range groups {
		if checks, ok := registry.groups[group]; ok {
			ret = append(ret, checks...)
		}
		// else do we log an error?
	}

	return ret
}

// Get retrieves the specified check from the registry,
// throws an error if not found
func Get(name string) (Check, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	if registry.checks[name] != nil {
		return registry.checks[name], nil
	}
	return nil, fmt.Errorf("Check not found: %s", name)
}
