package resolvers

import (
	"github.com/perfectmak/k4indie/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
)

func EndpointsWithDomains(endpoints *v1alpha1.ApplicationEndpoints) []v1alpha1.ApplicationEndpoint {
	result := make([]v1alpha1.ApplicationEndpoint, 0, len(*endpoints))

	for _, endpoint := range *endpoints {
		if endpoint.Domain != "" {
			result = append(result, endpoint)
		}
	}

	return result
}

func BuildIngressRules(appName string, endpoints []v1alpha1.ApplicationEndpoint) []networkingv1.IngressRule {
	rules := make([]networkingv1.IngressRule, 0, len(endpoints))

	// group by domains
	groups := map[string][]v1alpha1.ApplicationEndpoint{}
	for _, endpoint := range endpoints {
		if _, exists := groups[endpoint.Domain]; !exists {
			groups[endpoint.Domain] = make([]v1alpha1.ApplicationEndpoint, 0, 5)
		}

		groups[endpoint.Domain] = append(groups[endpoint.Domain], endpoint)
	}

	// generate rules
	for domain, groupEndpoints := range groups {
		paths := make([]networkingv1.HTTPIngressPath, 0, len(groupEndpoints))

		for _, endpoint := range groupEndpoints {
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path: endpoint.DomainPath,
				PathType: &[]networkingv1.PathType{
					networkingv1.PathTypePrefix,
				}[0],
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: appName,
						Port: networkingv1.ServiceBackendPort{
							Number: endpoint.Port,
						},
					},
				},
			})
		}

		domainRules := networkingv1.IngressRule{
			Host: domain,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		}

		rules = append(rules, domainRules)
	}

	return rules
}
