// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"testing"

	testLog "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/fairwindsops/controller-utils/pkg/log"
)

func setupFakeData(t *testing.T) (Client, unstructured.Unstructured, unstructured.Unstructured, unstructured.Unstructured, unstructured.Unstructured) {

	// TODO move to a centralized place
	log.SetLogger(testLog.NewTestLogger(t))
	gv := schema.GroupVersion{Group: "apps", Version: "v1"}
	gvpod := schema.GroupVersion{Group: "", Version: "v1"}
	gvk := gv.WithKind("Deployment")
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gv, gvpod})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	gvk = gv.WithKind("ReplicaSet")
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	gvk = gvpod.WithKind("Pod")
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	pod := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Pod",
			"namespace":  "test",
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "ReplicaSet",
						"name":       "rs",
					},
				},
				"name":      "poddy",
				"namespace": "test",
			},
			"spec": map[string]interface{}{},
		},
	}
	pod2 := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "core/v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "core/v1",
						"kind":       "ReplicaNotASet",
						"name":       "rs",
					},
				},
				"name":      "poddy-bad",
				"namespace": "test2",
			},
			"spec": map[string]interface{}{},
		},
	}
	rs := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "ReplicaSet",
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"name":       "dep",
					},
				},
				"name":      "rs",
				"namespace": "test",
			},
		},
	}
	dep := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "dep",
				"namespace": "test",
			},
		},
	}
	depNoPods := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "dep-no-pods",
				"namespace": "test",
			},
		},
	}
	dynamic := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "apps", Version: "v1", Resource: "replicasets"}: "ReplicaSetList",
			{Group: "apps", Version: "v1", Resource: "deployments"}: "DeploymentList",
			{Group: "", Version: "v1", Resource: "pods"}:            "PodsList",
		},
	)
	mapping, err := restMapper.RESTMapping(gvpod.WithKind("Pod").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &pod, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test2").Create(context.TODO(), &pod2, metav1.CreateOptions{})
	assert.NoError(t, err)
	mapping, err = restMapper.RESTMapping(gv.WithKind("ReplicaSet").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &rs, metav1.CreateOptions{})
	assert.NoError(t, err)
	mapping, err = restMapper.RESTMapping(gv.WithKind("Deployment").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &dep, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &depNoPods, metav1.CreateOptions{})
	assert.NoError(t, err)
	client := Client{
		Dynamic: dynamic,
		RESTMapper: restMapper,
		Context: context.TODO(),
	}
	return client, pod, rs, dep, pod2
}

func TestGetTopController(t *testing.T) {
	client, pod, rs, dep, pod2 := setupFakeData(t)
	controller, err := client.GetTopController(pod, map[string]unstructured.Unstructured{})
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	controller, err = client.GetTopController(rs, map[string]unstructured.Unstructured{})
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	controller, err = client.GetTopController(dep, map[string]unstructured.Unstructured{})
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	controller, err = client.GetTopController(pod2, map[string]unstructured.Unstructured{})
	assert.Error(t, err)
}

func TestGetAllTopControllers(t *testing.T) {
	client, _, _, _, _ := setupFakeData(t)
	controllers, err := client.GetAllTopControllersSummary("")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(controllers))
	controllers, err = client.GetAllTopControllersSummary("test")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(controllers))
}
