package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var simplifiedRe = regexp.MustCompile(`(\S+)=(".*"|<% .* %>|\S+)`)

// ParseSimplified will attempt to parse a "simplified" line such as:
// action.name key=val key=val
func ParseSimplified(line string) (map[string]interface{}, error) {
	params := map[string]interface{}{}

	v := simplifiedRe.FindAllStringSubmatch(line, -1)
	if v != nil {
		for _, i := range v {
			if len(i) == 3 {
				key := i[1]
				value := strings.Trim(i[2], `"`)

				params[key] = value
			}
		}
	}

	return params, nil
}

// ValidateTags ensures a struct field is valid by the custom tags it has.
func ValidateTags(s interface{}) error {
	vValue := reflect.ValueOf(s)
	if vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}

	tValue := reflect.TypeOf(s)
	if tValue.Kind() == reflect.Ptr {
		tValue = tValue.Elem()
	}

	for i := 0; i < vValue.NumField(); i++ {
		vField := vValue.Field(i)
		tField := tValue.Field(i)

		if vField.Kind() == reflect.Struct {
			if tField.Name != strings.Title(tField.Name) {
				continue
			}

			s := vValue.Field(i).Addr().Interface()
			if err := ValidateTags(s); err != nil {
				return err
			}
		}

		z := reflect.Zero(vField.Type())

		if requiredTag := tField.Tag.Get("required"); requiredTag == "true" {
			switch vField.Kind() {
			case reflect.String:
				if vField.Interface() == z.Interface() {
					return fmt.Errorf("missing input: %s", tField.Name)
				}
			case reflect.Slice:
				if vField.Interface() == nil {
					return fmt.Errorf("missing input: %s", tField.Name)
				}
			}
		}

		if !vField.CanSet() {
			continue
		}

		if defaultTag := tField.Tag.Get("default"); defaultTag != "" {
			if vField.Interface() == z.Interface() {
				switch tField.Type.Name() {
				case "bool":
					switch defaultTag {
					case "true":
						vValue.Field(i).SetBool(true)
					case "false":
						vValue.Field(i).SetBool(false)
					}

				case "int":
					v, err := strconv.ParseInt(defaultTag, 10, 64)
					if err != nil {
						return err
					}

					vValue.Field(i).SetInt(v)

				case "string":
					vValue.Field(i).SetString(defaultTag)
				}

			}
		}
	}

	return nil
}
