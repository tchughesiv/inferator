// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRule":       schema_pkg_apis_rule_v1alpha1_OperationRule(ref),
		"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleSpec":   schema_pkg_apis_rule_v1alpha1_OperationRuleSpec(ref),
		"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleStatus": schema_pkg_apis_rule_v1alpha1_OperationRuleStatus(ref),
	}
}

func schema_pkg_apis_rule_v1alpha1_OperationRule(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "OperationRule is the Schema for the operationrules API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleSpec", "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1.OperationRuleStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_rule_v1alpha1_OperationRuleSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "OperationRuleSpec defines the desired state of OperationRule",
				Type:        []string{"object"},
			},
		},
	}
}

func schema_pkg_apis_rule_v1alpha1_OperationRuleStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "OperationRuleStatus defines the observed state of OperationRule",
				Type:        []string{"object"},
			},
		},
	}
}
