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
)

var podSpecFields = []string{"jobTemplate", "spec", "template"}

// GetPodSpec looks inside arbitrary YAML for a PodSpec
func GetPodSpec(obj map[string]interface{}) (*corev1.PodSpec, error) {
	// TODO examine this for ways to make it more efficient.
	for _, child := range podSpecFields {
		if childYaml, ok := obj[child]; ok {
			return GetPodSpec(childYaml.(map[string]interface{}))
		}
	}
	if _, ok := obj["containers"]; !ok {
		return nil, nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	podSpec := corev1.PodSpec{}
	err = json.Unmarshal(b, &podSpec)
	return &podSpec, err
}
