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

import (
	"errors"
	"net/http"
	"time"
)

const delimiter = ":"

type options struct {
	paths            []string
	kubeContext      string
	yaml             []byte
	transportWrapper TransportWrapper
	timeout          time.Duration
	inCluster        bool
}

// Option function that allows injecting options while building kube.Client.
type Option func(*options) error

// WithConfigFile returns an Option injected with a config file path.
func WithConfigFile(path string) Option {
	return func(o *options) error {
		o.paths = []string{path}
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

// WithMergedConfigFiles returns an Option injected with value of $KUBECONFIG
func WithMergedConfigFiles(paths []string) Option {
	return func(o *options) error {
		o.paths = paths
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

// TransportWrapper wraps an http.RoundTripper
type TransportWrapper = func(http.RoundTripper) http.RoundTripper

// WithTransportWrapper allows wrapping the underlying http.RoundTripper
func WithTransportWrapper(f TransportWrapper) Option {
	return func(o *options) error {
		o.transportWrapper = f
		return nil
	}
}

// InCluster indicates that we are accessing the Kubernetes API from a Pod
func InCluster() Option {
	return func(o *options) error {
		o.inCluster = true
		return nil
	}
}

func (o *options) validate() error {
	if o.yaml != nil && len(o.paths) != 0 {
		return errors.New("cannot specify yaml and kubeconfig file paths")
	}
	return nil
}
