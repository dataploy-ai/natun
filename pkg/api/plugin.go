/*
Copyright 2022 Natun.

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

package api

import (
	"context"
	manifests "github.com/natun-ai/natun/pkg/api/v1alpha1"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Plugins interface {
	BindConfig | FeatureApply | DataConnectorReconcile | StateFactory |
		CollectNotifierFactory | WriteNotifierFactory |
		HistoricalWriterFactory
}

// BindConfig adds config flags for the plugin.
type BindConfig func(set *pflag.FlagSet) error

// FeatureApply applies changes on the feature abstraction.
type FeatureApply func(metadata Metadata, builderSpec []byte, api FeatureAbstractAPI, engine Engine) error

// DataConnectorReconcile is the interface to be implemented by plugins that want to be reconciled in the operator.
// This is useful for plugins that need to spawn external Feature Ingestion.
type DataConnectorReconcile func(ctx context.Context, client client.Client, scheme *runtime.Scheme, coreAddr string, conn *manifests.DataConnector) error

// StateFactory is the interface to be implemented by plugins that implements storage providers.
type StateFactory func(viper *viper.Viper) (State, error)

// NotifierFactory is the interface to be implemented by plugins that implements Notifier.
type NotifierFactory[T Notification] func(viper *viper.Viper) (Notifier[T], error)
type CollectNotifierFactory NotifierFactory[CollectNotification]
type WriteNotifierFactory NotifierFactory[WriteNotification]

type HistoricalWriterFactory func(viper *viper.Viper) (HistoricalWriter, error)
