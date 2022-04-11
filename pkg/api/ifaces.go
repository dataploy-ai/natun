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
	"github.com/go-logr/logr"
	manifests "github.com/natun-ai/natun/pkg/api/v1alpha1"
	"time"
)

// Engine is the main engine of the Core
// It is responsible for the low-level operation for the features against the feature store
type Engine interface {
	Metadata(ctx context.Context, FQN string) (Metadata, error)
	Get(ctx context.Context, FQN string, entityID string) (Value, Metadata, error)
	Set(ctx context.Context, FQN string, entityID string, val any, ts time.Time) error
	Append(ctx context.Context, FQN string, entityID string, val any, ts time.Time) error
	Incr(ctx context.Context, FQN string, entityID string, by any, ts time.Time) error
	Update(ctx context.Context, FQN string, entityID string, val any, ts time.Time) error
}
type MetadataGetter func(ctx context.Context, FQN string) (Metadata, error)

// Logger is a simple interface that returns a Logr.Logger
type Logger interface {
	Logger() logr.Logger
}

// Manager is the main manager of the Core
// It is responsible for managing features as well as operating on them
type Manager interface {
	BindFeature(in *manifests.Feature) error
	UnbindFeature(FQN string) error
	HasFeature(FQN string) bool
}
type ManagerEngine interface {
	Logger
	Manager
	Engine
}
