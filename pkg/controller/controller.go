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
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/fairwindsops/controller-utils/pkg/log"
)

var objectCache map[string]unstructured.Unstructured = make(map[string]unstructured.Unstructured)

// GetTopController finds the highest level owner of whatever object is passed in.
func GetTopController(ctx context.Context, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, unstructuredObject metav1.Object) (metav1.Object, error) {
	owners := unstructuredObject.GetOwnerReferences()
	if len(owners) > 0 {
		if len(owners) > 1 {
			log.GetLogger().V(1).Info("Found more than one owner", unstructuredObject.GetName(), unstructuredObject.GetNamespace())
		}
		firstOwner := owners[0]
		if firstOwner.Kind == "Node" {
			// Don't treat the node as a valid controller.
			// This happens for static pods.
			return unstructuredObject, nil
		}
		key := fmt.Sprintf("%s/%s/%s", firstOwner.Kind, unstructuredObject.GetNamespace(), firstOwner.Name)
		abstractObject, ok := objectCache[key]
		if !ok {
			err := cacheAllObjectsOfKind(ctx, firstOwner.APIVersion, firstOwner.Kind, dynamicClient, restMapper, objectCache)
			if err != nil {
				return unstructuredObject, err
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				return unstructuredObject, errors.New("the owner could not be found for this object " + key)
			}
		}
		parentObject, err := meta.Accessor(&abstractObject)
		if err != nil {
			return unstructuredObject, err
		}
		return GetTopController(ctx, dynamicClient, restMapper, parentObject)
	}
	return unstructuredObject, nil
}

func cacheAllObjectsOfKind(ctx context.Context, apiVersion, kind string, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		log.GetLogger().V(0).Info("Error retrieving mapping", apiVersion, kind, err)
		return err
	}

	objects, err := dynamicClient.Resource(mapping.Resource).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.GetLogger().V(0).Info("Error retrieving parent object", mapping.Resource.Version, mapping.Resource.Resource, err)
		return err
	}
	for idx, object := range objects.Items {
		key := fmt.Sprintf("%s/%s/%s", object.GetKind(), object.GetNamespace(), object.GetName())

		objectCache[key] = objects.Items[idx]
	}
	return nil
}
