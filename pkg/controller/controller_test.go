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
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"

	"github.com/fairwindsops/controller-utils/pkg/log"
)

func TestGetTopController(t *testing.T) {
	// TODO move to a centralized place
	log.SetLogger(testLog.TestLogger{T: t})
	dynamic := dynamicFake.NewSimpleDynamicClient(k8sruntime.NewScheme())
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
			"kind": "Pod",
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
			"kind": "Pod",
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "core/v1",
						"kind":       "ReplicaNotASet",
						"name":       "rs",
					},
				},
				"name":      "poddy-bad",
				"namespace": "test",
			},
			"spec": map[string]interface{}{},
		},
	}
	rs := unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "ReplicaSet",
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
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name":      "dep",
				"namespace": "test",
			},
		},
	}
	mapping, err := restMapper.RESTMapping(gvpod.WithKind("Pod").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &pod, metav1.CreateOptions{})
	assert.NoError(t, err)
	mapping, err = restMapper.RESTMapping(gv.WithKind("ReplicaSet").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &rs, metav1.CreateOptions{})
	assert.NoError(t, err)
	mapping, err = restMapper.RESTMapping(gv.WithKind("Deployment").GroupKind())
	assert.NoError(t, err)
	_, err = dynamic.Resource(mapping.Resource).Namespace("test").Create(context.TODO(), &dep, metav1.CreateOptions{})
	assert.NoError(t, err)
	podObj, err := meta.Accessor(&pod)
	assert.NoError(t, err)
	controller, err := GetTopController(context.TODO(), dynamic, restMapper, podObj)
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	rsObj, err := meta.Accessor(&rs)
	assert.NoError(t, err)
	controller, err = GetTopController(context.TODO(), dynamic, restMapper, rsObj)
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	depObj, err := meta.Accessor(&dep)
	assert.NoError(t, err)
	controller, err = GetTopController(context.TODO(), dynamic, restMapper, depObj)
	assert.NoError(t, err)
	assert.Equal(t, "dep", controller.GetName())
	pod2Obj, err := meta.Accessor(&pod2)
	assert.NoError(t, err)
	controller, err = GetTopController(context.TODO(), dynamic, restMapper, pod2Obj)
	assert.Error(t, err)
}
