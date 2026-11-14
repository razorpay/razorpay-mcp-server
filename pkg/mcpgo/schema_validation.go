package mcpgo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// toNumber reports the numeric value of v. Arguments that arrive over the
// transport are JSON-decoded into float64, while in-process callers pass Go
// native numbers, so both forms are accepted.
func toNumber(value interface{}) (float64, bool) {
	if number, ok := value.(json.Number); ok {
		parsed, err := number.Float64()
		return parsed, err == nil
	}

	switch reflect.ValueOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return float64(reflect.ValueOf(value).Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return float64(reflect.ValueOf(value).Uint()), true
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(value).Float(), true
	default:
		return 0, false
	}
}

func isKind(value interface{}, kinds ...reflect.Kind) bool {
	actual := reflect.ValueOf(value).Kind()
	for _, kind := range kinds {
		if actual == kind {
			return true
		}
	}
	return false
}

// lengthOf returns the element count of a slice, array or map.
func lengthOf(value interface{}) (int, bool) {
	if !isKind(value, reflect.Slice, reflect.Array, reflect.Map) {
		return 0, false
	}
	return reflect.ValueOf(value).Len(), true
}

// patternCache stores compiled parameter patterns. Tool schemas are fixed at
// build time, so the cache stays small.
var patternCache sync.Map

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if cached, ok := patternCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil //nolint:errcheck // stored type
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	patternCache.Store(pattern, compiled)
	return compiled, nil
}

// ValidateArguments checks call arguments against the declared parameter
// schemas and returns one message per violation. The MCP transport does not
// validate input, so a tool that declares a constraint must enforce it here.
// An absent optional parameter is not a violation. Messages reuse the wording
// of the per-tool validator so that a caller sees one consistent format.
func ValidateArguments(
	parameters []ToolParameter,
	arguments interface{},
) []string {
	args, ok := arguments.(map[string]interface{})
	if !ok {
		return nil
	}

	var problems []string
	for _, param := range parameters {
		value, present := args[param.Name]
		if !present || value == nil {
			if required, _ := param.Schema["required"].(bool); required {
				problems = append(
					problems,
					"missing required parameter: "+param.Name,
				)
			}
			continue
		}
		problems = append(
			problems,
			validateValue(param.Name, param.Schema, value)...,
		)
	}

	// Problems stay in parameter declaration order so that messages read the
	// same way as the per-tool validator's.
	return problems
}

func validateValue(
	name string,
	schema map[string]interface{},
	value interface{},
) []string {
	var problems []string

	declaredType, _ := schema["type"].(string)
	if problem := checkType(name, declaredType, value); problem != "" {
		// A wrong type makes every other check meaningless.
		return []string{problem}
	}

	if problem := checkEnum(name, schema, value); problem != "" {
		problems = append(problems, problem)
	}

	switch declaredType {
	case "number", "integer":
		problems = append(problems, checkNumber(name, schema, value)...)
	case "string":
		problems = append(problems, checkString(name, schema, value)...)
	case "array":
		problems = append(problems, checkArray(name, schema, value)...)
	case "object":
		problems = append(problems, checkObject(name, schema, value)...)
	}

	return problems
}

// wrongType reuses the wording the per-tool validator already produces, so a
// caller sees one consistent message for a bad type.
func wrongType(name string) string {
	return "invalid parameter type: " + name
}

// checkIntegerType reports why value is not a usable integer, or "".
func checkIntegerType(name string, value interface{}) string {
	number, ok := toNumber(value)
	if !ok {
		return wrongType(name)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return fmt.Sprintf("%s must be a finite integer", name)
	}
	if number != math.Trunc(number) {
		return fmt.Sprintf(
			"%s must be a whole number, got %v. Amounts are in the "+
				"smallest currency unit, so they cannot have a fraction",
			name, number,
		)
	}
	if math.Abs(number) > float64(1<<53) {
		return fmt.Sprintf("%s is too large to represent exactly", name)
	}
	return ""
}

func checkType(name, declaredType string, value interface{}) string {
	switch declaredType {
	case "number":
		if _, ok := toNumber(value); !ok {
			return wrongType(name)
		}
	case "integer":
		return checkIntegerType(name, value)
	case "string":
		if !isKind(value, reflect.String) {
			return wrongType(name)
		}
	case "boolean":
		if !isKind(value, reflect.Bool) {
			return wrongType(name)
		}
	case "array":
		if !isKind(value, reflect.Slice, reflect.Array) {
			return wrongType(name)
		}
	case "object":
		if !isKind(value, reflect.Map) {
			return wrongType(name)
		}
	}
	return ""
}

func checkEnum(
	name string,
	schema map[string]interface{},
	value interface{},
) string {
	allowed, ok := schema["enum"].([]interface{})
	if !ok || len(allowed) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(allowed))
	for _, candidate := range allowed {
		if candidate == value {
			return ""
		}
		rendered = append(rendered, fmt.Sprintf("%v", candidate))
	}

	return fmt.Sprintf(
		"%s must be one of: %s", name, strings.Join(rendered, ", "),
	)
}

func checkNumber(
	name string,
	schema map[string]interface{},
	value interface{},
) []string {
	number, ok := toNumber(value)
	if !ok {
		return nil
	}

	var problems []string
	if minimum, ok := schema["minimum"].(float64); ok && number < minimum {
		problems = append(problems, fmt.Sprintf(
			"%s must be at least %v, got %v", name, minimum, number,
		))
	}
	if maximum, ok := schema["maximum"].(float64); ok && number > maximum {
		problems = append(problems, fmt.Sprintf(
			"%s must be at most %v, got %v", name, maximum, number,
		))
	}
	return problems
}

func checkString(
	name string,
	schema map[string]interface{},
	value interface{},
) []string {
	if !isKind(value, reflect.String) {
		return nil
	}
	text := reflect.ValueOf(value).String()

	var problems []string
	if minLength, ok := schema["minLength"].(int); ok && len(text) < minLength {
		problems = append(problems, fmt.Sprintf(
			"%s must be at least %d characters", name, minLength,
		))
	}
	if maxLength, ok := schema["maxLength"].(int); ok && len(text) > maxLength {
		problems = append(problems, fmt.Sprintf(
			"%s must be at most %d characters", name, maxLength,
		))
	}
	// An empty string means "not supplied". Required and minLength already
	// cover that case, and they give a clearer message than a regex would.
	if pattern, ok := schema["pattern"].(string); ok && pattern != "" &&
		text != "" {
		compiled, err := compilePattern(pattern)
		if err == nil && !compiled.MatchString(text) {
			problems = append(problems, fmt.Sprintf(
				"%s has an invalid format, expected %s", name, pattern,
			))
		}
	}
	return problems
}

func checkArray(
	name string,
	schema map[string]interface{},
	value interface{},
) []string {
	count, ok := lengthOf(value)
	if !ok {
		return nil
	}

	var problems []string
	if minItems, ok := schema["minItems"].(int); ok && count < minItems {
		problems = append(problems, fmt.Sprintf(
			"%s must have at least %d items", name, minItems,
		))
	}
	if maxItems, ok := schema["maxItems"].(int); ok && count > maxItems {
		problems = append(problems, fmt.Sprintf(
			"%s must have at most %d items", name, maxItems,
		))
	}
	return problems
}

func checkObject(
	name string,
	schema map[string]interface{},
	value interface{},
) []string {
	count, ok := lengthOf(value)
	if !ok {
		return nil
	}

	var problems []string
	minProps, ok := schema["minProperties"].(int)
	if ok && count < minProps {
		problems = append(problems, fmt.Sprintf(
			"%s must have at least %d entries", name, minProps,
		))
	}
	maxProps, ok := schema["maxProperties"].(int)
	if ok && count > maxProps {
		problems = append(problems, fmt.Sprintf(
			"%s must have at most %d entries", name, maxProps,
		))
	}
	return problems
}

// withSchemaValidation wraps a handler so that declared constraints are
// enforced before the handler runs. Tools move money, so an out-of-range or
// wrongly typed value must be rejected here rather than forwarded upstream.
func withSchemaValidation(
	parameters []ToolParameter,
	next ToolHandler,
) ToolHandler {
	return func(ctx context.Context, r CallToolRequest) (*ToolResult, error) {
		if problems := ValidateArguments(parameters, r.Arguments); len(
			problems,
		) > 0 {
			return NewToolResultError(
				"Validation errors:\n- " + strings.Join(problems, "\n- "),
			), nil
		}
		return next(ctx, r)
	}
}
