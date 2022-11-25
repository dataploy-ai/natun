/*
Copyright (c) 2022 RaptorML authors.

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

package engine

import (
	"fmt"
	"github.com/raptor-ml/raptor/api"
	manifests "github.com/raptor-ml/raptor/api/v1alpha1"
	"github.com/raptor-ml/raptor/internal/stats"
	"github.com/raptor-ml/raptor/pkg/plugins"
)

// FeatureWithEngine converts the k8s Feature CRD to the internal engine implementation.
// This is useful as a standalone function for validating features.
func FeatureWithEngine(e api.ExtendedManager, in *manifests.Feature) (*Feature, error) {
	fd, err := api.FeatureDescriptorFromManifest(in)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FeatureDescriptor from CR: %w", err)
	}

	ft := Feature{
		FeatureDescriptor: *fd,
	}

	if p := plugins.FeatureAppliers.Get(ft.Builder); p != nil {
		err := p(ft.FeatureDescriptor, in.Spec.Builder, &ft, e)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no feature builder found for %s", ft.Builder)
	}
	return &ft, nil
}

// BindFeature converts the k8s Feature CRD to the internal implementation, and adds it to the engine.
func (e *engine) BindFeature(in *manifests.Feature) error {
	ft, err := FeatureWithEngine(e, in)
	if err != nil {
		return fmt.Errorf("failed to parse FeatureDescriptor from CR: %w", err)
	}
	return e.bindFeature(ft)
}

func (e *engine) UnbindFeature(fqn string) error {
	defer stats.DecNumberOfFeatures()
	e.features.Delete(fqn)
	e.logger.Info("feature unbound", "feature", fqn)
	return nil
}

func (e *engine) bindFeature(f *Feature) error {
	defer stats.IncNumberOfFeatures()
	if e.HasFeature(f.FQN) {
		return fmt.Errorf("%w: %s", api.ErrFeatureAlreadyExists, f.FQN)
	}
	e.features.Store(f.FQN, f)
	e.logger.Info("feature bound", "FQN", f.FQN)
	return nil
}

func (e *engine) HasFeature(fqn string) bool {
	_, ok := e.features.Load(fqn)
	return ok
}

func (e *engine) BindDataSource(fd api.DataSource) error {
	e.dataSources.Store(fd.FQN, fd)
	return nil
}
func (e *engine) UnbindDataSource(FQN string) error {
	e.dataSources.Delete(FQN)
	return nil
}
func (e *engine) HasDataSource(FQN string) bool {
	_, ok := e.dataSources.Load(FQN)
	return ok
}

func (e *engine) GetDataSource(fqn string) (api.DataSource, error) {
	fd, ok := e.dataSources.Load(fqn)
	if !ok {
		return api.DataSource{}, fmt.Errorf("DataSource %s not found", fqn)
	}
	return fd.(api.DataSource), nil
}
