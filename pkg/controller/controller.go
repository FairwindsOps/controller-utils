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

// Workload represents a workload in the cluster. It contains the top level object and all of the pods.
type Workload struct {
	TopController unstructured.Unstructured
	Pods          []unstructured.Unstructured
}

func getAllPods(ctx context.Context, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, namespace string) ([]unstructured.Unstructured, error) {
	fqKind := schema.FromAPIVersionAndKind("v1", "Pod")
	mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving mapping", "v1", "Pod")
		return nil, err
	}
	pods, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetAllTopControllers returns the highest level owning object of all pods. If a namespace is provided than this is limited to that namespace.
func GetAllTopControllers(ctx context.Context, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, namespace string) ([]Workload, error) {
	pods, err := getAllPods(ctx, dynamicClient, restMapper, namespace)
	if err != nil {
		return nil, err
	}
	workloadMap := map[string]Workload{}
	objectCache := map[string]unstructured.Unstructured{}
	// TODO avoid cycling over multiple pods with the same parent
	for _, pod := range pods {
		controller, err := GetTopController(ctx, dynamicClient, restMapper, pod, objectCache)
		if err != nil {
			// Do not return the error so that we can retrieve as many top level controllers as possible.
			log.GetLogger().Error(err, "An error occured retrieving the top level controller for this pod", pod.GetName(), pod.GetNamespace())
		}
		key := getControllerKey(controller)
		existingWorkload, ok := workloadMap[key]
		if !ok {
			existingWorkload.TopController = controller
		}
		existingWorkload.Pods = append(existingWorkload.Pods, pod)
		workloadMap[key] = existingWorkload
	}
	workloads := make([]Workload, 0)
	for _, workload := range workloadMap {
		workloads = append(workloads, workload)
	}
	return workloads, nil
}

// GetAllTopControllersSummary returns the highest level owning object of all pods and all of the pods. If a namespace is provided than this is limited to that namespace.
func GetAllTopControllersSummary(ctx context.Context, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, namespace string) ([]unstructured.Unstructured, error) {
	pods, err := getAllPods(ctx, dynamicClient, restMapper, namespace)
	if err != nil {
		return nil, err
	}
	workloadMap := map[string]unstructured.Unstructured{}
	objectCache := map[string]unstructured.Unstructured{}
	dedupedPods := dedupePods(pods)
	for _, pod := range dedupedPods {
		controller, err := GetTopController(ctx, dynamicClient, restMapper, pod, objectCache)
		if err != nil {
			// Do not return the error so that we can retrieve as many top level controllers as possible.
			log.GetLogger().Error(err, "An error occured retrieving the top level controller for this pod", pod.GetName(), pod.GetNamespace())
		}
		workloadMap[getControllerKey(controller)] = controller
	}
	workloads := make([]unstructured.Unstructured, 0)
	for _, workload := range workloadMap {
		workloads = append(workloads, workload)
	}
	return workloads, nil
}

func getControllerKey(controller unstructured.Unstructured) string {
	return fmt.Sprintf("%s/%s/%s", controller.GetKind(), controller.GetNamespace(), controller.GetName())
}

func dedupePods(pods []unstructured.Unstructured) []unstructured.Unstructured {
	var dedupedPods []unstructured.Unstructured
	dedupeMap := map[string]unstructured.Unstructured{}
	for _, pod := range pods {
		owners := pod.GetOwnerReferences()
		if len(owners) == 0 {
			dedupedPods = append(dedupedPods, pod)
			continue
		}
		dedupeMap[fmt.Sprintf("%s/%s/%s", pod.GetNamespace(), owners[0].Kind, owners[0].Name)] = pod
	}
	for _, pod := range dedupeMap {
		dedupedPods = append(dedupedPods, pod)
	}
	return dedupedPods
}

// GetTopController finds the highest level owner of whatever object is passed in.
func GetTopController(ctx context.Context, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, unstructuredObject unstructured.Unstructured, objectCache map[string]unstructured.Unstructured) (unstructured.Unstructured, error) {
	owners := unstructuredObject.GetOwnerReferences()
	if len(owners) > 0 {
		if objectCache == nil {
			objectCache = map[string]unstructured.Unstructured{}
		}
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
			err := cacheAllObjectsOfKind(ctx, firstOwner.APIVersion, firstOwner.Kind, unstructuredObject.GetNamespace(), dynamicClient, restMapper, objectCache)
			if err != nil {
				return unstructuredObject, err
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				return unstructuredObject, errors.New("this object could not be found for this object " + key)
			}
		}
		return GetTopController(ctx, dynamicClient, restMapper, abstractObject, objectCache)
	}
	return unstructuredObject, nil
}

func cacheAllObjectsOfKind(ctx context.Context, apiVersion, kind, namespace string, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving mapping", apiVersion, kind)
		return err
	}

	objects, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving parent object", mapping.Resource.Version, mapping.Resource.Resource)
		return err
	}
	for idx, object := range objects.Items {
		key := getControllerKey(object)

		objectCache[key] = objects.Items[idx]
	}
	return nil
}
