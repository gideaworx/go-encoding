package urlvalues // import "go.gideaworx.io/go-encoding/urlvalues"

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
//	     	SomeJoinedSlice []string `url:"joined,join=', '`
//			ID             int8
//			mySecretValue  bool
//		}
//
// and an instance
//
//		myExample := Example{
//			MyStringValue: "value1",
//			MySkippedValue: complex64(3+2i),
//			MyTime: time.Now().UTC(),
//			SomeSlice: []float64{1.2, 3.4, 5.6},
//	     	SomeJoinedSlice: []string{"hello", "world"}
//			mySecretValue: false,
//		}
//
// urlvalues.MarshalURLValues(myExample) will return a url.Values instance whose Encode() method would return
//
//	"mystring=value1&slice=1.2&slice=3.4&slice=5.6&joined=hello%2C%20world&time=2022-07-03T12%3A22%3A09Z&ID=0"
//
// time.Time objects will be formatted in RFC3339 format, and error instances will be serialized by calling their
// Error() method. See the unit tests for deeper examples.
func MarshalURLValues(i any) (url.Values, error) {
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
			s, err := stringFromValue(rv.Index(i), rv.Index(i).Type(), "")
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
		join := ""
		tagString, ok := sf.Tag.Lookup("url")
		if ok {
			tag, err := parseTag(tagString)
			if err != nil {
				if errors.Is(err, errSkip) {
					continue
				}

				return err
			}

			key = tag.name
			join = tag.joinString
			omit = tag.omitEmpty
		}

		format, ok := sf.Tag.Lookup("urlformat")
		if !ok {
			format = ""
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

			valueStrings := make([]string, 0, fv.Len())
			for j := 0; j < fv.Len(); j++ {
				str, err := stringFromValue(fv.Index(j), fv.Index(j).Type(), format)
				if err != nil {
					if errors.Is(err, errSkip) {
						continue
					}
					return err
				}

				if join == "" {
					values.Add(key, str)
					continue
				}

				valueStrings = append(valueStrings, str)
			}

			if len(valueStrings) > 0 {
				values.Set(key, strings.Join(valueStrings, join))
			}

			continue
		}

		str, err := stringFromValue(fv, sf.Type, format)
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
	case time.Duration:
		return concrete.String(), nil
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return fmt.Sprintf("%d", concrete), nil
	case string:
		return concrete, nil
	case time.Time:
		return concrete.Format(time.RFC3339), nil
	case error:
		return concrete.Error(), nil
	}

	return stringFromValue(reflect.ValueOf(a), reflect.TypeOf(a), "")
}

func stringFromValue(v reflect.Value, t reflect.Type, format string) (string, error) {
	if !v.IsValid() {
		if t.Kind() == reflect.Pointer {
			v = reflect.New(t.Elem())
		} else {
			return "", errors.New("invalid value")
		}
	}

	if t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if v.IsZero() {
			return zeroValue(t)
		}
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", errSkip
		}

		v = v.Elem()
	}

	i := v.Interface()
	if t, ok := i.(time.Time); ok {
		if format != "" {
			return t.Format(format), nil
		}
		return t.Format(time.RFC3339), nil
	}

	if d, ok := i.(time.Duration); ok {
		if format != "" {
			parts := strings.Split(format, ",")
			if !strings.EqualFold(parts[0], "int") {
				return d.String(), nil
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

				return fmt.Sprintf("%d", int64(d/unit)), nil
			}
		}

		return d.String(), nil
	}

	if e, ok := i.(error); ok {
		return e.Error(), nil
	}

	switch v.Kind() {
	case reflect.Bool:
		b := v.Bool()
		switch strings.ToLower(format) {
		case "int":
			if b {
				return "1", nil
			}
			return "0", nil
		case "shortlower":
			if b {
				return "t", nil
			}
			return "f", nil
		case "short":
			if b {
				return "T", nil
			}
			return "F", nil
		case "upper":
			if b {
				return "TRUE", nil
			}
			return "FALSE", nil
		case "camel":
			if b {
				return "True", nil
			}
			return "False", nil
		case "lower":
			fallthrough
		default:
			return strconv.FormatBool(b), nil
		}
	case reflect.Complex64:
		var f byte = 'f'
		if format != "" {
			f = format[0]
			if f != 'e' && f != 'E' && f != 'f' && f != 'g' && f != 'G' {
				return "", fmt.Errorf("bad verb %s. only e, E, f, g, and G are currently supported", string(format[0]))
			}
		}
		str := strconv.FormatComplex(v.Complex(), f, -1, 64)
		return str, nil
	case reflect.Complex128:
		var f byte = 'f'
		if format != "" {
			f = format[0]
			if f != 'e' && f != 'E' && f != 'f' && f != 'g' && f != 'G' {
				return "", fmt.Errorf("bad verb %s. only e, E, f, g, and G are currently supported", string(format[0]))
			}
		}
		str := strconv.FormatComplex(v.Complex(), f, -1, 128)
		return str, nil
	case reflect.Float32:
		var f byte = 'f'
		if format != "" {
			f = format[0]
			if f != 'e' && f != 'E' && f != 'f' && f != 'g' && f != 'G' {
				return "", fmt.Errorf("bad verb %s. only e, E, f, g, and G are currently supported", string(format[0]))
			}
		}
		str := strconv.FormatFloat(v.Float(), f, -1, 32)
		return str, nil
	case reflect.Float64:
		var f byte = 'f'
		if format != "" {
			f = format[0]
			if f != 'e' && f != 'E' && f != 'f' && f != 'g' && f != 'G' {
				return "", fmt.Errorf("bad verb %s. only e, E, f, g, and G are currently supported", string(format[0]))
			}
		}
		str := strconv.FormatFloat(v.Float(), f, -1, 64)
		return str, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil
	case reflect.String:
		return v.String(), nil
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
