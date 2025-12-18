// Copyright © 2025 Ping Identity Corporation

package schema

import (
	"reflect"
	"slices"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
)

// MustGenerateSchema generates a JSON schema for type T.
//
// uuid.UUID fields are represented as strings with "uuid" format.
//
// "AdditionalProperties" (a field automatically added in the PingOne client generation process)
// is removed from "required" for all schemas.
//
// The JSON schema "additionalProperties" field is also configured to allow additional properties,
// to handle future API changes where new fields may be added that are not yet supported in the client.
func MustGenerateSchema[T any]() *jsonschema.Schema {
	forOptions := &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[uuid.UUID](): {
				Type:   "string",
				Format: "uuid",
			},
		},
	}

	schema, err := jsonschema.For[T](forOptions)
	if err != nil {
		panic("failed to generate schema: " + err.Error())
	}

	schema = excludeFieldsFromRequired(schema, []string{"AdditionalProperties"})
	// Allow additional properties to handle future API changes
	schema = allowAdditionalProperties(schema)

	return schema
}

// excludeFieldsFromRequired removes specified fields from the "required" list of a JSON schema, returning a new schema.
// This function traverses the schema and removes any properties whose names are in the excludeFieldsList.
// It handles nested objects and arrays to ensure complete field exclusion throughout the schema hierarchy.
func excludeFieldsFromRequired(schema *jsonschema.Schema, excludeFieldsList []string) *jsonschema.Schema {
	if schema == nil || len(excludeFieldsList) == 0 {
		return nil
	}

	// Create a copy of the schema to avoid modifying the original
	result := &jsonschema.Schema{}
	*result = *schema

	// Remove excluded fields from "required"
	if result.Required != nil {
		result.Required = slices.DeleteFunc(result.Required, func(e string) bool {
			return slices.Contains(excludeFieldsList, e)
		})
	}

	// Handle properties (object fields)
	if result.Properties != nil {
		newProperties := make(map[string]*jsonschema.Schema)
		for key, propSchema := range result.Properties {
			// Recursively process nested schemas
			newProperties[key] = excludeFieldsFromRequired(propSchema, excludeFieldsList)
		}
		result.Properties = newProperties
	}

	// Handle array items, allOf, anyOf, oneOf
	if result.Items != nil {
		result.Items = excludeFieldsFromRequired(result.Items, excludeFieldsList)
	}

	if result.AllOf != nil {
		newAllOf := make([]*jsonschema.Schema, len(result.AllOf))
		for i, subschema := range result.AllOf {
			newAllOf[i] = excludeFieldsFromRequired(subschema, excludeFieldsList)
		}
		result.AllOf = newAllOf
	}

	if result.AnyOf != nil {
		newAnyOf := make([]*jsonschema.Schema, len(result.AnyOf))
		for i, subschema := range result.AnyOf {
			newAnyOf[i] = excludeFieldsFromRequired(subschema, excludeFieldsList)
		}
		result.AnyOf = newAnyOf
	}

	if result.OneOf != nil {
		newOneOf := make([]*jsonschema.Schema, len(result.OneOf))
		for i, subschema := range result.OneOf {
			newOneOf[i] = excludeFieldsFromRequired(subschema, excludeFieldsList)
		}
		result.OneOf = newOneOf
	}

	// Handle AddtionalProperties if it's a schema
	if result.AdditionalProperties != nil {
		result.AdditionalProperties = excludeFieldsFromRequired(result.AdditionalProperties, excludeFieldsList)
	}

	return result
}

// allowAdditionalProperties recursively allows additionalProperties for all object schemas.
// This allows the schema to accept new properties that may be added to the API but are not yet
// supported in the PingOne client.
func allowAdditionalProperties(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	// Create a copy of the schema to avoid modifying the original
	result := &jsonschema.Schema{}
	*result = *schema

	if result.Properties != nil {
		// By default, additionalProperties is configured to prevent any additional properties.
		// Setting it to nil will allow additional properties when validating JSON.
		result.AdditionalProperties = nil

		// Process nested object properties
		newProperties := make(map[string]*jsonschema.Schema)
		for key, propSchema := range result.Properties {
			newProperties[key] = allowAdditionalProperties(propSchema)
		}
		result.Properties = newProperties
	}

	// Handle array items, allOf, anyOf, oneOf
	if result.Items != nil {
		result.Items = allowAdditionalProperties(result.Items)
	}

	if result.AllOf != nil {
		newAllOf := make([]*jsonschema.Schema, len(result.AllOf))
		for i, subschema := range result.AllOf {
			newAllOf[i] = allowAdditionalProperties(subschema)
		}
		result.AllOf = newAllOf
	}

	if result.AnyOf != nil {
		newAnyOf := make([]*jsonschema.Schema, len(result.AnyOf))
		for i, subschema := range result.AnyOf {
			newAnyOf[i] = allowAdditionalProperties(subschema)
		}
		result.AnyOf = newAnyOf
	}

	if result.OneOf != nil {
		newOneOf := make([]*jsonschema.Schema, len(result.OneOf))
		for i, subschema := range result.OneOf {
			newOneOf[i] = allowAdditionalProperties(subschema)
		}
		result.OneOf = newOneOf
	}

	return result
}
