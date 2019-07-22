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

package kube

import "time"

type options struct {
	path        string
	kubeContext string
	yaml        []byte
	timeout     time.Duration
}

// Option function that allows injecting options while building kube.Client.
type Option func(*options) error

// WithConfigFile returns an Option injected with a config file path.
func WithConfigFile(path string) Option {
	return func(o *options) error {
		o.path = path
		return nil
	}
}

// WithKubeContext returns an Option injected with a kubernetes context.
func WithKubeContext(kubeContext string) Option {
	return func(o *options) error {
		o.kubeContext = kubeContext
		return nil
	}
}

// WithYaml returns an Option injected with a kubeconfig yaml.
func WithYaml(yaml []byte) Option {
	return func(o *options) error {
		o.yaml = yaml
		return nil
	}
}

// WithTimeout returns an Option injected with a timeout option while building client.
func WithTimeout(t time.Duration) Option {
	return func(o *options) error {
		o.timeout = t
		return nil
	}
}
