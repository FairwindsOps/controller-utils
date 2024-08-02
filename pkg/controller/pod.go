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
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podSpecFields = []string{"jobTemplate", "spec", "template"}
var controllerValidKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "CronJob", "Job"}

// GetPodMetadataAndSpec looks inside arbitrary YAML for a PodSpec and it's metadata
func GetPodMetadataAndSpec(obj map[string]any) (*metav1.ObjectMeta, *corev1.PodSpec, error) {
	return getPodMetadataAndSpecRecursively(nil, obj)
}

func getPodMetadataAndSpecRecursively(parent map[string]any, obj map[string]any) (*metav1.ObjectMeta, *corev1.PodSpec, error) {
	// TODO examine this for ways to make it more efficient.
	for _, child := range podSpecFields {
		if childYaml, ok := obj[child]; ok {
			return getPodMetadataAndSpecRecursively(obj, childYaml.(map[string]any))
		}
	}
	if _, ok := obj["containers"]; !ok {
		return nil, nil, nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, nil, err
	}
	// pod spec found,
	var podSpec corev1.PodSpec
	err = json.Unmarshal(b, &podSpec)
	if err != nil {
		return nil, nil, err
	}
	// looks for its metadata
	metadata, err := getMetadata(parent)
	if err != nil {
		return nil, nil, err
	}
	return metadata, &podSpec, nil
}

func getMetadata(parent map[string]any) (*metav1.ObjectMeta, error) {
	if parent == nil {
		return nil, nil
	}
	if _, ok := parent["metadata"]; !ok {
		return nil, nil
	}
	b, err := json.Marshal(parent["metadata"])
	if err != nil {
		return nil, err
	}
	var metadata metav1.ObjectMeta
	err = json.Unmarshal(b, &metadata)
	return &metadata, err
}

// ValidateIfControllerMatches checks if a child object is controlled by a parent object
func ValidateIfControllerMatches(child map[string]any, controller map[string]any) error {
	if child["metadata"].(map[string]any)["ownerReferences"].([]any)[0].(map[string]any)["uid"] != controller["metadata"].(map[string]any)["uid"] {
		return fmt.Errorf("controller does not match ownerReference uid")
	}
	if child["metadata"].(map[string]any)["namespace"].(string) != controller["metadata"].(map[string]any)["namespace"].(string) {
		return fmt.Errorf("controller namespace %s does not match ownerReference namespace %s", controller["metadata"].(map[string]any)["namespace"], child["metadata"].(map[string]any)["ownerReferences"].([]any)[0].(map[string]any)["namespace"])
	}
	if !lo.Contains(controllerValidKinds, controller["kind"].(string)) {
		return fmt.Errorf("controller kind %s is not a valid controller kind", controller["kind"].(string))
	}
	childContainers := getChildContainers(child)
	controllerContainers := getControllerContainers(controller)
	if len(childContainers) != len(controllerContainers) {
		return fmt.Errorf("length of controller container does not match child containers")
	}
	childContainerNames := lo.Map(childContainers, func(container any, _ int) string {
		return container.(map[string]any)["name"].(string)
	})
	controllerContainerNames := lo.Map(controllerContainers, func(container any, _ int) string {
		return container.(map[string]any)["name"].(string)
	})
	for _, childContainerName := range childContainerNames {
		if !lo.Contains(controllerContainerNames, childContainerName) {
			return fmt.Errorf("controller does not match child containers names")
		}
	}
	return nil
}

func getChildContainers(child map[string]any) []any {
	if _, ok := child["spec"].(map[string]any)["containers"]; ok {
		return child["spec"].(map[string]any)["containers"].([]any)
	} else if _, ok := child["spec"].(map[string]any)["jobTemplate"]; ok {
		return child["spec"].(map[string]any)["jobTemplate"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
	}
	return child["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
}

func getControllerContainers(controller map[string]any) []any {
	if _, ok := controller["spec"].(map[string]any)["jobTemplate"]; ok {
		return controller["spec"].(map[string]any)["jobTemplate"].(map[string]any)["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
	}
	return controller["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
}
