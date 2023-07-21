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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodSpec(t *testing.T) {
	podMetadata, podSpec, err := GetPodMetadataAndSpec(readFile(t, "./testdata/secret.json"))
	assert.NoError(t, err)
	assert.Nil(t, podMetadata)
	assert.Nil(t, podSpec)

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/deployment.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Equal(t, 2, len(podMetadata.Labels))
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/cronjob.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Nil(t, podMetadata)
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/daemon-set.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Equal(t, 1, len(podMetadata.Labels))
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/job.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Nil(t, podMetadata)
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/replica-set.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Equal(t, 1, len(podMetadata.Labels))
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/replication-controller.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Equal(t, 1, len(podMetadata.Labels))
	assert.Equal(t, 1, len(podSpec.Containers))

	podMetadata, podSpec, err = GetPodMetadataAndSpec(readFile(t, "./testdata/stateful-set.json"))
	assert.NoError(t, err)
	assert.NotNil(t, podSpec)
	assert.Equal(t, 1, len(podMetadata.Labels))
	assert.Equal(t, 1, len(podSpec.Containers))
}

func readFile(t *testing.T, file string) map[string]any {
	contents, err := os.ReadFile(file)
	assert.NoError(t, err)
	var object map[string]any
	err = json.Unmarshal(contents, &object)
	assert.NoError(t, err)
	return object
}
