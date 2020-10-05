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
