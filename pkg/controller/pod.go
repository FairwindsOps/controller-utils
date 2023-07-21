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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podSpecFields = []string{"jobTemplate", "spec", "template"}

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
