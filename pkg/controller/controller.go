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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/fairwindsops/controller-utils/pkg/log"
)

const podStatusRunning = "Running"

type knownKind struct {
	kind       string
	apiVersion string
}

var knownKinds = []knownKind{{
	"Deployment", "apps/v1",
}, {
	"ReplicaSet", "apps/v1",
}, {
	"CronJob", "batch/v1",
}, {
	"Job", "batch/v1",
}, {
	"DaemonSet", "apps/v1",
}, {
	"StatefulSet", "apps/v1",
}}

// Workload represents a workload in the cluster. It contains the top level object and all of the pods.
type Workload struct {
	TopController   unstructured.Unstructured
	Pods            []unstructured.Unstructured
	PodSpec         *corev1.PodSpec
	PodCount        int
	RunningPodCount int
}

// Client is used to interact with the Kubernetes API
type Client struct {
	Context    context.Context
	Dynamic    dynamic.Interface
	RESTMapper meta.RESTMapper
}

func (client Client) getAllPods(namespace string) ([]unstructured.Unstructured, error) {
	fqKind := schema.FromAPIVersionAndKind("v1", "Pod")
	mapping, err := client.RESTMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving mapping", "v1", "Pod")
		return nil, err
	}
	pods, err := client.Dynamic.Resource(mapping.Resource).Namespace(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func getPodStatus(unst unstructured.Unstructured) string {
	obj := unst.UnstructuredContent()
	if statusI, ok := obj["status"]; ok {
		if status, ok := statusI.(map[string]interface{}); ok {
			if phaseI, ok := status["phase"]; ok {
				if phase, ok := phaseI.(string); ok {
					return phase
				}
			}
		}
	}
	return ""
}

func (client Client) prepCacheWithKnownControllers(namespace string, objectCache map[string]unstructured.Unstructured) error {
	for _, kind := range knownKinds {
		err := client.cacheAllObjectsOfKind(kind.apiVersion, kind.kind, namespace, objectCache, true)
		if err != nil {
			log.GetLogger().V(3).Info("Unable to prime cache with objects of kind " + kind.kind)
		}
	}
	return nil
}

// GetAllTopControllersSummary returns the highest level owning object of all pods
// If a namespace is provided than this is limited to that namespace.
// This can be more memory-efficient than GetAllTopControllersWithPods, since it does not include individual pods.
func (client Client) GetAllTopControllersSummary(namespace string) ([]Workload, error) {
	return client.getAllTopControllers(namespace, false)
}

// GetAllTopControllersWithPods returns the highest level owning object of all pods, as well as all pods.
// If a namespace is provided than this is limited to that namespace.
func (client Client) GetAllTopControllersWithPods(namespace string) ([]Workload, error) {
	return client.getAllTopControllers(namespace, true)
}

func (client Client) getAllTopControllers(namespace string, includePods bool) ([]Workload, error) {
	workloadMap := map[string]Workload{}
	objectCache := map[string]unstructured.Unstructured{}
	err := client.prepCacheWithKnownControllers(namespace, objectCache)
	if err != nil {
		return nil, err
	}
	for _, controller := range objectCache {
		key := getControllerKey(controller)
		podSpec, err := GetPodSpec(controller.UnstructuredContent())
		if err != nil {
			return nil, err
		}
		workloadMap[key] = Workload{
			TopController: controller,
			PodSpec:       podSpec,
		}
	}
	pods, err := client.getAllPods(namespace)
	if err != nil {
		return nil, err
	}
	// TODO avoid cycling over multiple pods with the same parent
	for _, pod := range pods {
		controller, err := client.GetTopController(pod, objectCache)
		if err != nil {
			// Do not return the error so that we can retrieve as many top level controllers as possible.
			log.GetLogger().Error(err, "An error occured retrieving the top level controller for this pod", pod.GetName(), pod.GetNamespace())
		}
		key := getControllerKey(controller)
		existingWorkload, ok := workloadMap[key]
		if !ok {
			existingWorkload.TopController = controller
			podSpec, err := GetPodSpec(controller.UnstructuredContent())
			if err != nil {
				return nil, err
			}
			if podSpec == nil {
				podSpec, err = GetPodSpec(pod.UnstructuredContent())
				if err != nil {
					return nil, err
				}
			}
			existingWorkload.PodSpec = podSpec
		}
		existingWorkload.PodCount++
		if getPodStatus(pod) == podStatusRunning {
			existingWorkload.RunningPodCount++
		}
		if includePods {
			existingWorkload.Pods = append(existingWorkload.Pods, pod)
		}
		workloadMap[key] = existingWorkload
	}
	workloads := make([]Workload, 0)
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
func (client Client) GetTopController(unstructuredObject unstructured.Unstructured, objectCache map[string]unstructured.Unstructured) (unstructured.Unstructured, error) {
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
			err := client.cacheAllObjectsOfKind(firstOwner.APIVersion, firstOwner.Kind, unstructuredObject.GetNamespace(), objectCache, false)
			if err != nil {
				return unstructuredObject, err
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				return unstructuredObject, errors.New("this object could not be found for this object " + key)
			}
		}
		return client.GetTopController(abstractObject, objectCache)
	}
	return unstructuredObject, nil
}

func (client Client) cacheAllObjectsOfKind(apiVersion, kind, namespace string, objectCache map[string]unstructured.Unstructured, mustBeTopLevel bool) error {
	log.GetLogger().V(9).Info("cache all", apiVersion, kind)
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := client.RESTMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving mapping", apiVersion, kind)
		return err
	}

	objects, err := client.Dynamic.Resource(mapping.Resource).Namespace(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		log.GetLogger().Error(err, "Error retrieving parent object", mapping.Resource.Version, mapping.Resource.Resource)
		return err
	}
	for idx, object := range objects.Items {
		if mustBeTopLevel && len(object.GetOwnerReferences()) > 0 {
			continue
		}
		key := getControllerKey(object)
		objectCache[key] = objects.Items[idx]
	}
	return nil
}
