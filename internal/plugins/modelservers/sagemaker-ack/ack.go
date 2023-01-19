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

package sagemaker_ack

// ack-sagemaker integration
// +kubebuilder:rbac:groups=sagemaker.services.k8s.aws,resources=models,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sagemaker.services.k8s.aws,resources=endpointconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sagemaker.services.k8s.aws,resources=endpoints,verbs=get;list;watch;create;update;patch;delete

import (
	"github.com/raptor-ml/raptor/pkg/plugins"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const name = "sagemaker-ack"

func init() {
	// Register the plugin
	plugins.ModelReconciler.Register(name, reconcile)
	plugins.ModelControllerOwns.Register(name, owns)
}

var ackModelGVK = schema.GroupVersionKind{
	Group:   "sagemaker.services.k8s.aws",
	Version: "v1alpha1",
	Kind:    "Model",
}

var ackEndpointConfigGVK = schema.GroupVersionKind{
	Group:   "sagemaker.services.k8s.aws",
	Version: "v1alpha1",
	Kind:    "EndpointConfig",
}

var ackEndpointGVK = schema.GroupVersionKind{
	Group:   "sagemaker.services.k8s.aws",
	Version: "v1alpha1",
	Kind:    "Endpoint",
}

func owns() []client.Object {
	var ret []client.Object
	for _, gvk := range []schema.GroupVersionKind{ackModelGVK, ackEndpointConfigGVK, ackEndpointGVK} {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		ret = append(ret, u)
	}
	return ret
}
