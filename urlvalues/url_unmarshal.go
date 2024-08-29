package urlvalues // import "go.gideaworx.io/go-encoding/urlvalues"

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// URLValuesUnmarshaler allows implementations to decode a url.Values object in a custom way
type URLValuesUnmarshaler interface {
	UnmarshalURLValues(url.Values) error
}

// UnmarshalURLValues will take a url.Values and deserialize it into the given object. The second argument
// a must be a non-nil pointer to a map[string]any, instance of URLValuesUnmarshaler, or struct. If the argument
// is a *map[string]any, each map key is the name of the parameter, and each map value s will be deserialized in the
// following way (and in the following order):
//
// * if s can be parsed as a bool, it will return a bool
//
// * if s can be parsed as a real number, it will return a float64
//
// * if s can be parsed as a complex number, it will return a complex128
//
// * if s can be parsed as an RFC3339 timestamp, it will return a time.Time
//
// * if none of the above are true, s will be return unparsed
//
// if a parameter has multiple values, the map key will contain an instance of []any with each slice element parsed
// according to the above rules. If the argument is a *struct, each parameter will be deserialized, if possible,
// to the corresponding struct field's type, using the field's "url" struct tag to map the parameter name to field
// name, if present. Unexported fields and fields with struct tag `url:"-"` are skipped. If the struct tag ends in
// ',omitempty' and the value is the type's zero value, it will not be explicitly set.
func UnmarshalURLValues(values url.Values, a any) error {
	if a == nil {
		return errors.New("second argument must not be nil")
	}

	if m, ok := a.(*map[string]any); ok {
		newMap := unmarshalMap(values)

		*m = newMap
		return nil
	}

	aType := reflect.TypeOf(a)
	if aType.Kind() == reflect.Pointer {
		if um, ok := a.(URLValuesUnmarshaler); ok {
			return um.UnmarshalURLValues(values)
		}

		newStruct, err := unmarshalStruct(values, aType.Elem())
		if err != nil {
			return err
		}

		reflect.ValueOf(a).Elem().Set(newStruct)
		return nil
	}

	return errors.New("second argument must be a non-nil pointer to a map[string]any or struct")
}

func unmarshalMap(values url.Values) map[string]any {
	m := make(map[string]any)

	for k, vslice := range values {
		if len(vslice) == 1 {
			m[k] = fromStringToAny(vslice[0])
		} else {
			aslice := make([]any, 0, len(vslice))
			for _, s := range vslice {
				aslice = append(aslice, fromStringToAny(s))
			}
			m[k] = aslice
		}
	}

	return m
}

func unmarshalStruct(values url.Values, structType reflect.Type) (reflect.Value, error) {
	if structType.Kind() != reflect.Struct {
		return reflect.Zero(structType), errors.New("structType must be struct")
	}

	retValue := reflect.New(structType).Elem()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		structFieldValue := retValue.Field(i)
		if !structField.IsExported() {
			continue
		}

		parameterName := structField.Name
		omitEmpty := false
		join := ""
		tagString, ok := structField.Tag.Lookup("url")
		if ok {
			tag, err := parseTag(tagString)
			if err != nil {
				if errors.Is(err, errSkip) {
					continue
				}

				return reflect.Zero(structType), err
			}

			parameterName = tag.name
			join = tag.joinString
			omitEmpty = tag.omitEmpty
		}

		format, ok := structField.Tag.Lookup("urlformat")
		if !ok {
			format = ""
		}

		if !values.Has(parameterName) {
			continue
		}

		parsedValue, err := fromStringsToValue(values[parameterName], structField.Type, format, join)
		if err != nil {
			return parsedValue, err
		}

		if omitEmpty && (!parsedValue.IsValid() || parsedValue.IsZero()) {
			continue
		}

		if !structFieldValue.CanSet() {
			return reflect.Zero(structType), fmt.Errorf("cannot set field %s", structField.Name)
		}

		if !parsedValue.Type().AssignableTo(structField.Type) {
			return reflect.Zero(structType), fmt.Errorf("%s is not assignable to %s", parsedValue.Type(), structField.Type)
		}

		structFieldValue.Set(parsedValue)
	}

	return retValue, nil
}

// if s can be parsed as a bool, it will return a bool
// if s can be parsed as a real number, it will return a float64
// if s can be parsed as a complex number, it will return a complex128
// if s can be parsed as an RFC3339 timestamp, it will return a time.Time
// if none of the above are true, s will be return unparsed
func fromStringToAny(s string) any {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	if c, err := strconv.ParseComplex(s, 128); err == nil {
		return c
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	return s
}

func fromStringsToValue(values []string, fieldType reflect.Type, format string, join string) (reflect.Value, error) {
	var retVal reflect.Value

	if len(values) == 0 {
		return reflect.Zero(fieldType), errSkip
	}

	isPointerType := false

	retVal = reflect.New(fieldType)
	retType := fieldType
	if fieldType.Kind() == reflect.Pointer {
		isPointerType = true
		retType = fieldType.Elem()
		retVal = reflect.New(fieldType.Elem())
	}

	if (retType.Kind() == reflect.Slice || retType.Kind() == reflect.Array) &&
		len(values) == 1 && join != "" {
		values = strings.Split(values[0], join)
	}

	if retType.Kind() == reflect.Slice {
		sliceVal := reflect.MakeSlice(retType, len(values), len(values)+2)
		if !sliceVal.Type().AssignableTo(retType) {
			return reflect.Zero(retType), fmt.Errorf("cannot assign %s to %s", sliceVal.Type(), retType)
		}
		retVal.Elem().Set(sliceVal)
	}

	if retVal.Elem().Kind() == reflect.Array || retVal.Elem().Kind() == reflect.Slice {
		checkRetLen := retVal.Elem().Kind() == reflect.Array
		for i := 0; i < len(values) && (!checkRetLen || i < retVal.Elem().Len()); i++ {
			// the first Elem returns the value of the pointer, the second returns the underlying type of the iterable
			v, err := fromStringToValue(values[i], retVal.Elem().Type().Elem(), format)
			if err != nil {
				return reflect.Zero(retVal.Elem().Type()), err
			}

			if !retVal.Elem().Index(i).CanSet() {
				return reflect.Zero(fieldType), fmt.Errorf("cannot set item at index %d", i)
			}

			valToSet := retVal.Elem().Index(i)

			if !v.IsValid() || v.IsZero() {
				continue
			}

			if !v.Type().AssignableTo(valToSet.Type()) {
				return reflect.Zero(fieldType), fmt.Errorf("%s is not assignable to %s", v.Type(), valToSet.Type())
			}

			valToSet.Set(v)
		}
	} else if retVal.Elem().Kind() == reflect.String {
		retVal.Elem().Set(reflect.ValueOf(values[0]))
	} else {
		v, err := fromStringToValue(values[0], retVal.Elem().Type(), format)
		if err != nil {
			return reflect.Zero(retVal.Elem().Type()), err
		}

		if !retVal.Elem().CanSet() {
			return reflect.Zero(fieldType), errors.New("cannot set field")
		}

		if !v.Type().AssignableTo(retVal.Elem().Type()) {
			return reflect.Zero(fieldType), fmt.Errorf("%s is not assignable to %s", v.Type(), retVal.Elem().Type())
		}

		retVal.Elem().Set(v)
	}

	if isPointerType {
		return retVal, nil
	}

	return retVal.Elem(), nil
}

func fromStringToValue(s string, t reflect.Type, format string) (reflect.Value, error) {
	// handle durations first, since it's an alias for int64 and would be picked up by the switch
	durationType := reflect.TypeOf((*time.Duration)(nil)).Elem()
	if t == durationType {
		var d time.Duration
		var err error

		parts := strings.Split(format, ",")
		if !strings.EqualFold(parts[0], "int") {
			d, err = time.ParseDuration(s)
			return reflect.ValueOf(d), err
		}

		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.ValueOf(v), err
		}

		unit := time.Nanosecond
		if len(parts) > 1 {
			switch parts[1] {
			case "us":
				unit = time.Microsecond
			case "ms":
				unit = time.Millisecond
			case "s":
				unit = time.Second
			case "m":
				unit = time.Minute
			case "h":
				unit = time.Hour
			}

			if d = time.Duration(v); d == 0 {
				return reflect.ValueOf(d), nil
			}

			return reflect.ValueOf(d * unit), nil
		}
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(s), nil
	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		return reflect.ValueOf(v), err
	case reflect.Complex128:
		v, err := strconv.ParseComplex(s, 128)
		return reflect.ValueOf(v), err
	case reflect.Complex64:
		v, err := strconv.ParseComplex(s, 64)
		return reflect.ValueOf(complex64(v)), err
	case reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		return reflect.ValueOf(v), err
	case reflect.Float32:
		v, err := strconv.ParseFloat(s, 32)
		return reflect.ValueOf(float32(v)), err
	case reflect.Int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(v), err
	case reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return reflect.ValueOf(v), err
	case reflect.Int:
		v, err := strconv.Atoi(s)
		return reflect.ValueOf(v), err
	case reflect.Int8:
		v, err := strconv.ParseInt(s, 10, 8)
		return reflect.ValueOf(int8(v)), err
	case reflect.Int16:
		v, err := strconv.ParseInt(s, 10, 16)
		return reflect.ValueOf(int16(v)), err
	case reflect.Int32:
		v, err := strconv.ParseInt(s, 10, 32)
		return reflect.ValueOf(int32(v)), err
	case reflect.Uint:
		v, err := strconv.ParseUint(s, 10, 0)
		return reflect.ValueOf(uint(v)), err
	case reflect.Uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		return reflect.ValueOf(uint8(v)), err
	case reflect.Uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		return reflect.ValueOf(uint16(v)), err
	case reflect.Uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		return reflect.ValueOf(uint32(v)), err
	case reflect.Pointer:
		v, err := fromStringToValue(s, t.Elem(), format)
		if err != nil {
			return reflect.Zero(t), err
		}

		vPtr := reflect.New(t.Elem())
		vPtr.Elem().Set(v)

		return vPtr, nil
	}

	timeType := reflect.TypeOf((*time.Time)(nil)).Elem()
	if t.AssignableTo(timeType) {
		ts, err := time.Parse(time.RFC3339, s)
		return reflect.ValueOf(ts), err
	}

	errType := reflect.TypeOf((*error)(nil)).Elem()
	if errType.AssignableTo(t) {
		return reflect.ValueOf(errors.New(s)), nil
	}

	return reflect.Zero(t), fmt.Errorf("unsupported type %s", t)
}
