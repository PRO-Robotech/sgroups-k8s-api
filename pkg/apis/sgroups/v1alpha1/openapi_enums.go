package v1alpha1

import (
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// GetEnumOpenAPIDefinitions returns OpenAPI definitions with enum support for custom types.
func GetEnumOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		OpenAPIPrefix + "Action": {
			Schema: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Description: "AddressGroup default action (ALLOW or DENY)",
					Type:        []string{"string"},
					Enum: []interface{}{
						string(ActionUnknown),
						string(ActionAllow),
						string(ActionDeny),
					},
				},
			},
		},
	}
}

// GetOpenAPIDefinitionsWithEnums returns OpenAPI definitions including enum metadata.
func GetOpenAPIDefinitionsWithEnums(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	defs := GetOpenAPIDefinitions(ref)
	for key, value := range GetEnumOpenAPIDefinitions(ref) {
		defs[key] = value
	}

	modifyStructFieldsWithEnums(defs)

	return defs
}

func modifyStructFieldsWithEnums(defs map[string]common.OpenAPIDefinition) {
	agSpecKey := AddressGroupSpec{}.OpenAPIModelName()
	if agSpec, exists := defs[agSpecKey]; exists {
		if agSpec.Schema.Properties != nil {
			if actionProp, ok := agSpec.Schema.Properties["defaultAction"]; ok {
				actionProp.Enum = []interface{}{
					string(ActionUnknown),
					string(ActionAllow),
					string(ActionDeny),
				}
				actionProp.Description = "Default action for the address group (ALLOW or DENY)"
				agSpec.Schema.Properties["defaultAction"] = actionProp
			}
			defs[agSpecKey] = agSpec
		}
	}
}
