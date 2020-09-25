package podspec

var podSpecFields = []string{"jobTemplate", "spec", "template"}

// GetPodSpec looks inside arbitrary YAML for a PodSpec
func GetPodSpec(yaml map[string]interface{}) interface{} {
	for _, child := range podSpecFields {
		if childYaml, ok := yaml[child]; ok {
			return GetPodSpec(childYaml.(map[string]interface{}))
		}
	}
	if _, ok := yaml["containers"]; ok {
		return yaml
	}
	return nil
}
