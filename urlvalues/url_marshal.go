package urlvalues

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var errSkip = errors.New("skip")

// URLValuesMarshaler lets implementations convert themselves into a url.Values object. This is useful
// for HTTP APIs that take input in application/x-www-form-urlencoded format rather than another serialization
// mechanism like JSON, TOML, or YAML.
type URLValuesMarshaler interface {
	MarshalURLValues() (url.Values, error)
}

// MarshalURLValues will take an interface{} and attempt to serialize it into a url.Values object. The argument
// i must be a struct, map[string]any, URLValuesMarshaler or a pointer thereto. If using a struct, the
// value names can be controlled by the "url" struct tag. For example, given the struct
//
//		type Example struct {
//			MyStringValue  string    `url:"mystring"`
//			MySkippedValue any	     `url:"-"`
//			MyOmittedValue *int64    `url:"intptr,omitempty"`
//			MyTime         time.Time `url:"time"`
//			SomeSlice	   []float64 `url:"slice,omitempty"`
//			ID             int8
//			mySecretValue  bool
//		}
//
// and an instance
//
//		myExample := Example{
// 			MyStringValue: "value1",
// 			MySkippedValue: complex64(3+2i),
// 			MyTime: time.Now().UTC(),
//			SomeSlice: []float64{1.2, 3.4, 5.6},
// 			mySecretValue: false,
// 		}
//
// urlvalues.MarshalURLValues(myExample) will return a url.Values instance whose Encode() method would return
//
//		"mystring=value1&slice=1.2&slice=3.4&slice=5.6&time=2022-07-03T12%3A22%3A09Z&ID=0"
//
// time.Time objects will be formatted in RFC3339 format, and error instances will be serialized by calling their
// Error() method. See the unit tests for deeper examples.
func MarshalURLValues(i any) (url.Values, error) {
fmt.Println(
	"Hello",
)
	if u, ok := i.(URLValuesMarshaler); ok {
		return u.MarshalURLValues()
	}

	if i == nil {
		return url.Values{}, errors.New("value cannot be nil")
	}

	values := url.Values{}

	vo := reflect.ValueOf(i)
	if m, ok := i.(map[string]any); ok {

		for k, v := range m {
			if err := setValueFromMap(&values, k, v); err != nil {
				return url.Values{}, err
			}
		}

		return values, nil
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Struct {
		if err := setValuesFromStruct(&values, i); err != nil {
			return url.Values{}, err
		}

		return values, nil
	}

	if t.Kind() == reflect.Pointer {
		if vo.IsNil() {
			return url.Values{}, errors.New("value cannot be nil")
		}

		if t.Elem().Kind() == reflect.Struct {
			if err := setValuesFromStructPointer(&values, i); err != nil {
				return url.Values{}, err
			}

			return values, nil
		}
	}

	return url.Values{}, errors.New("argument must be a map[string]any, struct, or non-nil pointer to a struct")
}

func setValueFromMap(vals *url.Values, key string, val any) error {
	if vals == nil {
		return errors.New("vals cannot be nil")
	}

	rv := reflect.ValueOf(val)

	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			s, err := stringFromValue(rv.Index(i), rv.Index(i).Type())
			if err != nil {
				return err
			}

			vals.Add(key, s)
		}
	} else {
		s, err := stringFromConcrete(val)
		if err != nil {
			return err
		}

		vals.Set(key, s)
	}

	return nil
}

func setValuesFromStruct(values *url.Values, a any) error {
	t := reflect.TypeOf(a)
	v := reflect.ValueOf(a)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		fv := v.Field(i)

		key := sf.Name
		omit := false
		tag, ok := sf.Tag.Lookup("url")
		if ok {
			omit = strings.HasSuffix(tag, ",omitempty")
			tag = strings.TrimSuffix(tag, ",omitempty")
			if tag == "-" {
				continue
			}

			key = tag
		}

		if !fv.IsValid() || (fv.IsZero() && omit) {
			continue
		}

		if fv.Kind() == reflect.Pointer {
			if !fv.IsValid() || fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}

		if fv.Kind() == reflect.Array || fv.Kind() == reflect.Slice {
			if fv.Kind() == reflect.Slice && (!fv.IsValid() || fv.IsNil()) {
				continue
			}

			for j := 0; j < fv.Len(); j++ {
				str, err := stringFromValue(fv.Index(j), fv.Index(j).Type())
				if err != nil {
					if errors.Is(err, errSkip) {
						continue
					}
					return err
				}

				values.Add(key, str)
			}

			continue
		}

		str, err := stringFromValue(fv, sf.Type)
		if err != nil {
			if errors.Is(err, errSkip) {
				continue
			}
			return err
		}

		values.Set(key, str)
	}

	return nil
}

func setValuesFromStructPointer(values *url.Values, i any) error {
	v := reflect.ValueOf(i).Elem()
	return setValuesFromStruct(values, v.Interface())
}

func stringFromConcrete(a any) (string, error) {
	switch concrete := a.(type) {
	case bool:
		return strconv.FormatBool(concrete), nil
	case complex64:
		return strconv.FormatComplex(complex128(concrete), 'f', -1, 64), nil
	case complex128:
		return strconv.FormatComplex(concrete, 'f', -1, 128), nil
	case float32:
		return big.NewFloat(float64(concrete)).String(), nil
	case float64:
		return big.NewFloat(concrete).String(), nil
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return fmt.Sprintf("%d", concrete), nil
	case string:
		return concrete, nil
	case time.Time:
		return concrete.Format(time.RFC3339), nil
	case error:
		return concrete.Error(), nil
	}

	return stringFromValue(reflect.ValueOf(a), reflect.TypeOf(a))
}

func stringFromValue(v reflect.Value, t reflect.Type) (string, error) {
	if !v.IsValid() {
		if t.Kind() == reflect.Pointer {
			v = reflect.New(t.Elem())
		} else {
			return "", errors.New("invalid value")
		}
	}

	if v.IsZero() {
		return zeroValue(t)
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", errSkip
		}

		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Complex64:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 64), nil
	case reflect.Complex128:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 128), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil
	case reflect.String:
		return v.String(), nil
	}

	i := v.Interface()
	if t, ok := i.(time.Time); ok {
		return t.Format(time.RFC3339), nil
	}

	if e, ok := i.(error); ok {
		return e.Error(), nil
	}

	return "", fmt.Errorf("unsupported type %T", v.Interface())
}

func zeroValue(t reflect.Type) (string, error) {
	var (
		timeType = reflect.TypeOf(time.Time{})
		errType  = reflect.TypeOf((*error)(nil)).Elem()
	)

	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Pointer, reflect.Slice:
		return "", errSkip
	case reflect.Bool:
		return "false", nil
	case reflect.Complex64, reflect.Complex128,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "0", nil
	case reflect.String:
		return "", nil
	}

	if t == timeType {
		return time.Time{}.Format(time.RFC3339), nil
	}

	if t.AssignableTo(errType) {
		return "", nil
	}

	return "", fmt.Errorf("invalid type %s", t)
}
