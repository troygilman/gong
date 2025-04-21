package bind

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

func Bind(source url.Values, dest any) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be a pointer")
	}
	if val.IsNil() {
		return fmt.Errorf("destination is nil")
	}
	node := NewParser(NodeMapPool).Parse(source)
	defer node.Cleanup(NodeMapPool)
	return bind(node, val)
}

func bind(node Node, dest reflect.Value) error {
	t := dest.Type()
	switch dest.Kind() {
	case reflect.Pointer:
		if dest.IsNil() {
			dest.Set(reflect.New(t.Elem()))
		}
		return bind(node, dest.Elem())
	case reflect.Interface:
		if dest.IsNil() {
			// For interface{}, create a map[string]any
			m := make(map[string]any, len(node.Children))
			dest.Set(reflect.ValueOf(m))
		}
		return bind(node, dest.Elem())
	case reflect.Map:
		if dest.IsNil() {
			dest.Set(reflect.MakeMapWithSize(t, len(node.Children)))
		}
		keyType := t.Key()
		valueType := t.Elem()

		for key, child := range node.Children {
			// Create new key value
			keyValue := reflect.New(keyType).Elem()
			if err := setValueFromString(keyValue, key); err != nil {
				return err
			}

			// Create new value
			value := reflect.New(valueType).Elem()

			// If this is a leaf node (no children), set the value directly
			if len(child.Children) == 0 && child.Val != "" {
				if valueType.Kind() == reflect.Interface {
					value.Set(reflect.ValueOf(child.Val))
				} else {
					if err := setValueFromString(value, child.Val); err != nil {
						return err
					}
				}
			} else {
				// Otherwise, recursively bind the child node
				if err := bind(child, value); err != nil {
					return err
				}
			}

			// Set in map
			dest.SetMapIndex(keyValue, value)
		}
	case reflect.Struct:
		if t.ConvertibleTo(timeType) {
			if node.Val != "" {
				tm, err := time.Parse(time.RFC3339, node.Val)
				if err != nil {
					return err
				}
				dest.Set(reflect.ValueOf(tm))
			}
			return nil
		}
		for index := range dest.NumField() {
			field := dest.Field(index)
			if !field.CanInterface() {
				continue
			}
			tag := t.Field(index).Tag
			if sourceName, ok := tag.Lookup("form"); ok {
				child, ok := node.Children[sourceName]
				if !ok {
					continue
				}
				if err := bind(child, field); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		if dest.IsNil() {
			dest.Set(reflect.MakeSlice(t, 0, len(node.Children)))
		}
		if len(node.Children) > dest.Len() {
			dest.Grow(len(node.Children) - dest.Len())
		}
		for key, child := range node.Children {
			index, err := strconv.Atoi(key)
			if err != nil {
				return err
			}
			if index >= dest.Len() {
				dest.Grow(index - dest.Len() + 1)
				dest.SetLen(index + 1)
			}
			if err := bind(child, dest.Index(index)); err != nil {
				return err
			}
		}
	default:
		if dest.CanSet() {
			if err := setValueFromString(dest, node.Val); err != nil {
				return err
			}
		}
	}
	return nil
}

// setValueFromString sets a reflect.Value from a string based on its type
func setValueFromString(dest reflect.Value, str string) error {
	switch dest.Kind() {
	case reflect.String:
		dest.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		dest.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		dest.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		dest.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		dest.SetBool(val)
	default:
		if dest.CanAddr() {
			if tu, ok := dest.Addr().Interface().(encoding.TextUnmarshaler); ok {
				return tu.UnmarshalText([]byte(str))
			}
		}
	}
	return nil
}
