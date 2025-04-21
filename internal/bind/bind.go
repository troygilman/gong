package bind

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var arrayExpr = regexp.MustCompile(`^(.*?)\[([^\]]*)\](.*)$`)

func Bind(formData url.Values, dest any) error {
	parser := Parser{
		formData:  formData,
		arrayExpr: arrayExpr,
	}
	return parser.Parse(dest)
}

// Parser is a custom form parser for parsing nested form data
type Parser struct {
	formData  url.Values
	arrayExpr *regexp.Regexp
}

// Parse parses form data into the provided struct
func (p *Parser) Parse(dst interface{}) error {
	// Get a pointer to the destination value
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	// Get the actual value the pointer points to
	dstVal = dstVal.Elem()
	if dstVal.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to a struct")
	}

	// Process form data into a nested map structure
	nestedData := p.buildNestedMap()

	// Fill the struct with the parsed data
	return p.fillStruct(dstVal, nestedData, "")
}

// buildNestedMap converts flat form data to a nested map structure
func (p *Parser) buildNestedMap() map[string]interface{} {
	result := make(map[string]interface{})

	for key, values := range p.formData {
		// Handle single value fields
		if !p.arrayExpr.MatchString(key) {
			if len(values) == 1 {
				result[key] = values[0]
			} else {
				result[key] = values
			}
			continue
		}

		// Handle array/nested fields
		p.parseNestedKey(key, values, result)
	}

	return result
}

// parseNestedKey recursively parses a nested form key like "people[1][first_name]"
func (p *Parser) parseNestedKey(key string, values []string, data map[string]interface{}) {
	matches := p.arrayExpr.FindStringSubmatch(key)
	if len(matches) != 4 {
		// Not a nested key
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
		return
	}

	// Extract base key, index/key, and remaining parts
	baseKey := matches[1]
	indexOrKey := matches[2]
	remaining := matches[3]

	// Create base map/slice if it doesn't exist
	if _, exists := data[baseKey]; !exists {
		// If the index is numeric, create a map to collect array elements
		if indexOrKey != "" {
			data[baseKey] = make(map[string]interface{})
		} else {
			// Handle array with empty index like people[]
			data[baseKey] = make([]interface{}, 0)
		}
	}

	if remaining == "" {
		// This is a leaf node (no more nesting)
		if baseVal, ok := data[baseKey].(map[string]interface{}); ok {
			if len(values) == 1 {
				baseVal[indexOrKey] = values[0]
			} else {
				baseVal[indexOrKey] = values
			}
		} else if _, ok := data[baseKey].([]interface{}); ok && indexOrKey == "" {
			// Append to slice for empty index
			if len(values) == 1 {
				data[baseKey] = append(data[baseKey].([]interface{}), values[0])
			} else {
				data[baseKey] = append(data[baseKey].([]interface{}), values)
			}
		}
		return
	}

	// Handle further nesting
	if baseVal, ok := data[baseKey].(map[string]interface{}); ok {
		if _, exists := baseVal[indexOrKey]; !exists {
			if p.arrayExpr.MatchString(remaining) {
				baseVal[indexOrKey] = make(map[string]interface{})
			} else {
				// Remove the leading brackets for the field name
				fieldName := strings.TrimPrefix(remaining, "[")
				fieldName = strings.TrimSuffix(fieldName, "]")
				baseVal[indexOrKey] = make(map[string]interface{})
				mapVal := baseVal[indexOrKey].(map[string]interface{})
				if len(values) == 1 {
					mapVal[fieldName] = values[0]
				} else {
					mapVal[fieldName] = values
				}
				return
			}
		}

		// Continue parsing for the next level
		nestedKey := fmt.Sprintf("%s%s", indexOrKey, remaining)
		p.parseNestedKey(nestedKey, values, baseVal)
	}
}

// fillStruct fills a struct with values from the nested data map
func (p *Parser) fillStruct(dstVal reflect.Value, data map[string]interface{}, prefix string) error {
	dstType := dstVal.Type()

	// Iterate through struct fields
	for i := 0; i < dstVal.NumField(); i++ {
		field := dstVal.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := dstType.Field(i)
		formTag := structField.Tag.Get("form")
		if formTag == "-" {
			continue
		}

		if formTag == "" {
			formTag = structField.Name
		}

		formPath := formTag
		if prefix != "" {
			formPath = prefix + "." + formTag
		}

		// Get the value from the nested data
		value, err := p.getValueByPath(data, formPath)
		if err != nil || value == nil {
			continue
		}

		if err := p.setFieldValue(field, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", formPath, err)
		}
	}

	return nil
}

// getValueByPath retrieves a value from nested map using a dot-notation path
func (p *Parser) getValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			return current[part], nil
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("path not found: %s", path)
		}
	}

	return nil, fmt.Errorf("invalid path: %s", path)
}

// setFieldValue sets the appropriate value to the struct field based on its type
func (p *Parser) setFieldValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		if strVal, ok := value.(string); ok {
			field.SetString(strVal)
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intVal int64
		switch v := value.(type) {
		case string:
			var err error
			intVal, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
		case float64:
			intVal = int64(v)
		case int:
			intVal = int64(v)
		case int64:
			intVal = v
		default:
			return fmt.Errorf("cannot convert %T to int", value)
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var uintVal uint64
		switch v := value.(type) {
		case string:
			var err error
			uintVal, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				return err
			}
		case float64:
			uintVal = uint64(v)
		case uint:
			uintVal = uint64(v)
		case uint64:
			uintVal = v
		default:
			return fmt.Errorf("cannot convert %T to uint", value)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		var floatVal float64
		switch v := value.(type) {
		case string:
			var err error
			floatVal, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		default:
			return fmt.Errorf("cannot convert %T to float", value)
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		var boolVal bool
		switch v := value.(type) {
		case string:
			var err error
			boolVal, err = strconv.ParseBool(v)
			if err != nil {
				return err
			}
		case bool:
			boolVal = v
		default:
			return fmt.Errorf("cannot convert %T to bool", value)
		}
		field.SetBool(boolVal)

	case reflect.Slice:
		return p.setSliceValue(field, value)

	case reflect.Struct:
		return p.setStructValue(field, value)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Type().String())
	}

	return nil
}

// setSliceValue handles setting slice values from form data
func (p *Parser) setSliceValue(field reflect.Value, value interface{}) error {
	// Handle map of indexed values (typical for form arrays)
	if indexedMap, ok := value.(map[string]interface{}); ok {
		// Determine max index
		maxIndex := -1
		for k := range indexedMap {
			if idx, err := strconv.Atoi(k); err == nil && idx > maxIndex {
				maxIndex = idx
			}
		}

		if maxIndex == -1 {
			return nil // No valid indices
		}

		// Create a slice of the appropriate size
		sliceType := field.Type().Elem()
		newSlice := reflect.MakeSlice(field.Type(), maxIndex+1, maxIndex+1)

		// Fill the slice with values
		for idxStr, val := range indexedMap {
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				continue
			}

			elemValue := newSlice.Index(idx)

			if sliceType.Kind() == reflect.Struct {
				// For struct elements, need to create a map
				if nestedMap, ok := val.(map[string]interface{}); ok {
					if err := p.fillStruct(elemValue, nestedMap, ""); err != nil {
						return err
					}
				}
			} else {
				// For simple types
				if err := p.setFieldValue(elemValue, val); err != nil {
					return err
				}
			}
		}

		field.Set(newSlice)
		return nil
	}

	// Handle direct array values
	if arrayVal, ok := value.([]interface{}); ok {
		newSlice := reflect.MakeSlice(field.Type(), len(arrayVal), len(arrayVal))

		for i, val := range arrayVal {
			elemValue := newSlice.Index(i)
			if err := p.setFieldValue(elemValue, val); err != nil {
				return err
			}
		}

		field.Set(newSlice)
		return nil
	}

	// Handle single value that should be in a slice
	newSlice := reflect.MakeSlice(field.Type(), 1, 1)
	elemValue := newSlice.Index(0)

	if err := p.setFieldValue(elemValue, value); err != nil {
		return err
	}

	field.Set(newSlice)
	return nil
}

// setStructValue handles setting struct values from form data
func (p *Parser) setStructValue(field reflect.Value, value interface{}) error {
	if mapVal, ok := value.(map[string]interface{}); ok {
		return p.fillStruct(field, mapVal, "")
	}
	return fmt.Errorf("cannot set struct field from non-map value: %T", value)
}
