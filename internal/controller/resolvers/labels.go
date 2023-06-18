package resolvers

func MergeDefaultLabels(labelsList ...map[string]string) map[string]string {
	result := map[string]string{}
	defaultLabels := map[string]string{
		"app.kubernetes.io/part-of":    "k4indie-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
	labelsList = append(labelsList, defaultLabels)

	for _, labels := range labelsList {
		for k, v := range labels {
			result[k] = v
		}
	}

	return result
}
