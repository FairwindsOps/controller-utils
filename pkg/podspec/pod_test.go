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

package podspec

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodSpec(t *testing.T) {
	podSpec := readPodSpecFile(t, "./tests/deployment.json")
	assert.NotNil(t, podSpec)
	assert.Equal(t, 1, len(podSpec.(map[string]interface{})["containers"].([]interface{})))
	podSpec = readPodSpecFile(t, "./tests/secret.json")
	assert.Nil(t, podSpec)
}

func readPodSpecFile(t *testing.T, file string) interface{} {
	contents, err := ioutil.ReadFile(file)
	assert.NoError(t, err)
	var object map[string]interface{}
	err = json.Unmarshal(contents, &object)
	assert.NoError(t, err)
	return GetPodSpec(object)

}
