package preprocessor

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jbcom/openapi-31-to-30-converter/pkg/converter"
)

// OpenAPIVersion constants
const (
	// OpenAPIVersion30 is the target OpenAPI version (3.0.3)
	OpenAPIVersion30 = "3.0.3"

	// OpenAPIVersion31Prefix is the prefix for OpenAPI 3.1.x versions
	OpenAPIVersion31Prefix = "3.1"
)

// EnsureOpenAPICompatibility ensures the OpenAPI spec is compatible with ogen.
// It converts OpenAPI 3.1 specs to 3.0.3 compatible specs if needed.
// Returns the path to the compatible spec (either the original or a new temporary file).
func EnsureOpenAPICompatibility(specPath string) (string, error) {
	// Create a temporary file for the potentially modified spec
	tempFile, err := os.CreateTemp("", "openapi-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFile.Close() // Close immediately as the converter will reopen it
	tempFilePath := tempFile.Name()

	// Set up cleanup in case of errors
	var cleanupNeeded bool
	defer func() {
		if cleanupNeeded {
			os.Remove(tempFilePath)
		}
	}()

	// Try to convert the spec using the jbcom/openapi-31-to-30-converter library
	err = converter.Convert(specPath, tempFilePath)
	if err != nil {
		cleanupNeeded = true
		return "", fmt.Errorf("failed to convert OpenAPI spec: %w", err)
	}

	// Check if the file was actually modified (conversion was needed)
	convertedStat, err := os.Stat(tempFilePath)
	if err != nil {
		cleanupNeeded = true
		return "", fmt.Errorf("failed to stat converted file: %w", err)
	}

	// If the converted file is empty or very small, it likely failed silently
	if convertedStat.Size() < 10 {
		cleanupNeeded = true
		return "", fmt.Errorf("conversion resulted in an invalid file")
	}

	return tempFilePath, nil
}

// ConvertOpenAPI31To30 converts an OpenAPI 3.1 spec to a 3.0.3 compatible format.
// It handles all the major differences between the two versions.
func ConvertOpenAPI31To30(spec *map[string]interface{}) error {
	if spec == nil {
		return fmt.Errorf("nil spec provided")
	}

	// Downgrade version to 3.0.3
	(*spec)["openapi"] = OpenAPIVersion30

	// Process the entire spec recursively
	processSchema(*spec)

	// Handle specific OpenAPI 3.1 features that need special treatment
	handleJSONSchemaDialect(*spec)
	handleWebhooks(*spec)

	// Fix circular references
	if err := fixCircularReferences(*spec); err != nil {
		return fmt.Errorf("failed to fix circular references: %w", err)
	}

	return nil
}

// handleJSONSchemaDialect removes the jsonSchemaDialect field which is specific to OpenAPI 3.1
func handleJSONSchemaDialect(spec map[string]interface{}) {
	// Remove jsonSchemaDialect field if present
	delete(spec, "jsonSchemaDialect")
}

// handleWebhooks converts webhooks to paths with x-webhook extension
func handleWebhooks(spec map[string]interface{}) {
	// Check if webhooks are present
	webhooks, ok := spec["webhooks"].(map[string]interface{})
	if !ok {
		return
	}

	// Get or create paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		paths = make(map[string]interface{})
		spec["paths"] = paths
	}

	// Convert webhooks to paths with x-webhook extension
	for webhookPath, webhook := range webhooks {
		// Add x-webhook extension to indicate this is a webhook
		if webhookObj, ok := webhook.(map[string]interface{}); ok {
			webhookObj["x-webhook"] = true
			// Add to paths with a prefix to avoid conflicts
			paths["/webhooks/"+webhookPath] = webhookObj
		}
	}

	// Remove webhooks field
	delete(spec, "webhooks")
}

// processSchema recursively processes a schema to make it compatible with OpenAPI 3.0.3.
// It handles all the major differences between OpenAPI 3.1 and 3.0.3 schema formats.
func processSchema(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		// Process all OpenAPI 3.1 specific features
		processSchemaObject(v)

		// Process all properties recursively
		for key, value := range v {
			// Skip $ref fields to avoid processing referenced schemas
			if key != "$ref" {
				processSchema(value)
			}
		}
	case []interface{}:
		// Process all array elements recursively
		for _, item := range v {
			processSchema(item)
		}
	}
}

// processSchemaObject processes a single schema object to make it compatible with OpenAPI 3.0.3
func processSchemaObject(schema map[string]interface{}) {
	// Process type field (handle array types in OpenAPI 3.1)
	processTypeField(schema)

	// Handle exclusiveMinimum/exclusiveMaximum (different format in 3.1)
	processExclusiveMinMax(schema)

	// Handle content encoding and media type
	delete(schema, "contentEncoding")
	delete(schema, "contentMediaType")

	// Handle path parameters (must be required in 3.0)
	if inValue, ok := schema["in"]; ok {
		if inStr, ok := inValue.(string); ok && inStr == "path" {
			schema["required"] = true
		}
	}
}

// processTypeField handles array types in OpenAPI 3.1
func processTypeField(schema map[string]interface{}) {
	if typeValue, ok := schema["type"]; ok {
		if typeArray, ok := typeValue.([]interface{}); ok {
			// Check if null is in the type array (nullable in 3.1)
			hasNull := false
			var nonNullTypes []interface{}

			for _, t := range typeArray {
				if typeStr, ok := t.(string); ok {
					if typeStr == "null" {
						hasNull = true
					} else {
						nonNullTypes = append(nonNullTypes, t)
					}
				}
			}

			if hasNull {
				// In OpenAPI 3.0, nullable is a boolean property
				schema["nullable"] = true

				// If there's only one non-null type, use it directly
				if len(nonNullTypes) == 1 {
					schema["type"] = nonNullTypes[0]
				} else if len(nonNullTypes) > 1 {
					// Convert multiple non-null types to oneOf
					convertTypesToOneOf(schema, nonNullTypes)
				}
			} else if len(typeArray) > 1 {
				// Convert multiple types to oneOf
				convertTypesToOneOf(schema, typeArray)
			}
		}
	}
}

// convertTypesToOneOf converts multiple types to a oneOf schema
func convertTypesToOneOf(schema map[string]interface{}, types []interface{}) {
	oneOf := make([]interface{}, 0, len(types))
	for _, t := range types {
		if typeStr, ok := t.(string); ok {
			oneOf = append(oneOf, map[string]interface{}{
				"type": typeStr,
			})
		}
	}
	delete(schema, "type")
	schema["oneOf"] = oneOf
}

// processExclusiveMinMax handles exclusiveMinimum/exclusiveMaximum (different format in 3.1)
func processExclusiveMinMax(schema map[string]interface{}) {
	// Handle exclusiveMinimum
	if exclusiveMin, ok := schema["exclusiveMinimum"]; ok {
		if exclusiveMinNum, ok := exclusiveMin.(float64); ok {
			// In OpenAPI 3.0, exclusiveMinimum is a boolean and minimum is the value
			schema["minimum"] = exclusiveMinNum
			schema["exclusiveMinimum"] = true
		}
	}

	// Handle exclusiveMaximum
	if exclusiveMax, ok := schema["exclusiveMaximum"]; ok {
		if exclusiveMaxNum, ok := exclusiveMax.(float64); ok {
			// In OpenAPI 3.0, exclusiveMaximum is a boolean and maximum is the value
			schema["maximum"] = exclusiveMaxNum
			schema["exclusiveMaximum"] = true
		}
	}
}

// fixCircularReferences identifies and fixes circular references in the schema.
// It breaks circular references by replacing them with stub objects.
func fixCircularReferences(spec map[string]interface{}) error {
	// Get the components schemas
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return nil
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Create a placeholder schema for circular references
	schemas["CircularReferenceStub"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Placeholder for circular reference",
			},
		},
	}

	// Build a dependency graph of schemas
	dependencyGraph := buildDependencyGraph(schemas)

	// Detect circular dependencies
	circularDeps := detectCircularDependencies(dependencyGraph)

	// Log detected circular dependencies
	if len(circularDeps) > 0 {
		log.Printf("Detected %d circular dependencies in the schema", len(circularDeps))
		for i, cycle := range circularDeps {
			log.Printf("Circular dependency %d: %s", i+1, strings.Join(cycle, " -> "))
		}
	}

	// Break circular dependencies
	breakCircularDependencies(schemas, circularDeps)

	return nil
}

// buildDependencyGraph builds a graph of schema dependencies
func buildDependencyGraph(schemas map[string]interface{}) map[string][]string {
	graph := make(map[string][]string)

	for schemaName, schema := range schemas {
		schemaMap, ok := schema.(map[string]interface{})
		if !ok {
			continue
		}

		// Find all references in this schema
		deps := findReferences(schemaMap)
		if len(deps) > 0 {
			graph[schemaName] = deps
		}
	}

	return graph
}

// findReferences finds all schema references in a schema object
func findReferences(schema map[string]interface{}) []string {
	var refs []string

	// Check for direct $ref
	if ref, ok := schema["$ref"].(string); ok {
		if strings.HasPrefix(ref, "#/components/schemas/") {
			refName := strings.TrimPrefix(ref, "#/components/schemas/")
			refs = append(refs, refName)
		}
	}

	// Check properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for _, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				propRefs := findReferences(propMap)
				refs = append(refs, propRefs...)
			}
		}
	}

	// Check array items
	if items, ok := schema["items"].(map[string]interface{}); ok {
		itemRefs := findReferences(items)
		refs = append(refs, itemRefs...)
	}

	// Check oneOf, anyOf, allOf
	for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
		if array, ok := schema[keyword].([]interface{}); ok {
			for _, item := range array {
				if itemMap, ok := item.(map[string]interface{}); ok {
					itemRefs := findReferences(itemMap)
					refs = append(refs, itemRefs...)
				}
			}
		}
	}

	return refs
}

// detectCircularDependencies detects circular dependencies in the dependency graph
func detectCircularDependencies(graph map[string][]string) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	path := make(map[string]bool)

	var dfs func(node string, currentPath []string)
	dfs = func(node string, currentPath []string) {
		if path[node] {
			// Found a cycle
			cycleStart := -1
			for i, n := range currentPath {
				if n == node {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(currentPath[cycleStart:], node)
				cycles = append(cycles, cycle)
			}
			return
		}

		if visited[node] {
			return
		}

		visited[node] = true
		path[node] = true
		currentPath = append(currentPath, node)

		for _, dep := range graph[node] {
			dfs(dep, currentPath)
		}

		path[node] = false
	}

	for node := range graph {
		if !visited[node] {
			dfs(node, []string{})
		}
	}

	return cycles
}

// breakCircularDependencies breaks circular dependencies in the schema
func breakCircularDependencies(schemas map[string]interface{}, cycles [][]string) {
	// Create a set of edges to break
	edgesToBreak := make(map[string]map[string]bool)

	for _, cycle := range cycles {
		if len(cycle) < 2 {
			continue
		}

		// Choose an edge to break (we'll break the last edge in the cycle)
		from := cycle[len(cycle)-2]
		to := cycle[len(cycle)-1]

		if edgesToBreak[from] == nil {
			edgesToBreak[from] = make(map[string]bool)
		}
		edgesToBreak[from][to] = true
	}

	// Break the edges
	for from, toSet := range edgesToBreak {
		schema, ok := schemas[from].(map[string]interface{})
		if !ok {
			continue
		}

		// Break references in properties
		if props, ok := schema["properties"].(map[string]interface{}); ok {
			for propName, prop := range props {
				if propMap, ok := prop.(map[string]interface{}); ok {
					breakReference(propMap, toSet, propName)
				}
			}
		}
	}

	// Add known problematic schemas to the edges to break
	// This combines automatic detection with explicit handling
	knownProblematicEdges := map[string][]string{
		"AcknowledgementCriteria": {"VerificationCriteria"},
		"VerificationCriteria":    {"AcknowledgementCriteria"},
		"DocumentCriteria":        {"VerificationCriteria"},
		"AssetCriteria":           {"VerificationCriteria"},
		"KnowledgeCriteria":       {"VerificationCriteria"},
	}

	// Process known problematic edges
	for from, toList := range knownProblematicEdges {
		schema, ok := schemas[from].(map[string]interface{})
		if !ok {
			continue
		}

		properties, ok := schema["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		// Create a set of targets to break
		toSet := make(map[string]bool)
		for _, to := range toList {
			toSet[to] = true
		}

		// Break references in properties
		for propName, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				breakReference(propMap, toSet, propName)
			}
		}

		// Remove from required list if present
		if required, ok := schema["required"].([]interface{}); ok {
			newRequired := make([]interface{}, 0, len(required))
			for _, req := range required {
				if reqStr, ok := req.(string); ok {
					// Check if this required field is one we're breaking
					isCircularField := false
					for _, to := range toList {
						if strings.EqualFold(reqStr, to) || strings.EqualFold(reqStr, strings.ToLower(to)+"s") {
							isCircularField = true
							break
						}
					}
					if !isCircularField {
						newRequired = append(newRequired, req)
					}
				}
			}
			schema["required"] = newRequired
		}
	}
}

// breakReference breaks a reference if it points to a schema in the toSet
func breakReference(obj map[string]interface{}, toSet map[string]bool, propName string) {
	// Check for direct $ref
	if ref, ok := obj["$ref"].(string); ok {
		if strings.HasPrefix(ref, "#/components/schemas/") {
			refName := strings.TrimPrefix(ref, "#/components/schemas/")
			if toSet[refName] {
				// Replace with CircularReferenceStub
				obj["$ref"] = "#/components/schemas/CircularReferenceStub"
				log.Printf("Breaking circular reference: replaced %s with CircularReferenceStub", ref)
			}
		}
	}

	// Check array items
	if items, ok := obj["items"].(map[string]interface{}); ok {
		breakReference(items, toSet, propName+".items")
	}

	// Check oneOf, anyOf, allOf
	for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
		if array, ok := obj[keyword].([]interface{}); ok {
			for i, item := range array {
				if itemMap, ok := item.(map[string]interface{}); ok {
					breakReference(itemMap, toSet, fmt.Sprintf("%s.%s[%d]", propName, keyword, i))
				}
			}
		}
	}
}
