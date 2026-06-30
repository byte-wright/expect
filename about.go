package expect

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// AboutOption configures the per-type tolerances used by ToBeAbout.
type AboutOption func(*aboutConfig)

type aboutConfig struct {
	floatDelta    float64
	intDelta      int64
	timeDelta     time.Duration
	durationDelta time.Duration
}

// FloatDelta sets the maximum allowed absolute difference for float32/float64 leaves.
func FloatDelta(d float64) AboutOption {
	return func(c *aboutConfig) { c.floatDelta = d }
}

// IntDelta sets the maximum allowed absolute difference for signed and unsigned
// integer leaves.
func IntDelta(d int64) AboutOption {
	return func(c *aboutConfig) { c.intDelta = d }
}

// TimeDelta sets the maximum allowed absolute difference for time.Time leaves.
// Times are compared by instant (a.Sub(b)), not field by field.
func TimeDelta(d time.Duration) AboutOption {
	return func(c *aboutConfig) { c.timeDelta = d }
}

// DurationDelta sets the maximum allowed absolute difference for time.Duration leaves.
func DurationDelta(d time.Duration) AboutOption {
	return func(c *aboutConfig) { c.durationDelta = d }
}

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// ToBeAbout asserts that the value deeply equals expected, like ToBe, except that
// numeric, time.Time and time.Duration leaves only need to be within the deltas given
// by the options. Every other leaf and the structural shape must still match exactly.
// Without options it behaves like an exact match.
//
// Tolerances apply to top-level values and to values reachable through exported
// fields, slices, arrays and maps. Unexported leaves are compared exactly.
func (e Val) ToBeAbout(expected interface{}, opts ...AboutOption) Val {
	e.t.Helper()

	if !sameType(e.value, expected) {
		e.t.Errorf("expected %v to be of type %v but it is of type %v", e.name, typeName(expected), typeName(e.value))
		return e
	}

	if isNil(e.value) && isNil(expected) {
		return e
	}

	cfg := aboutConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	if msg, ok := cfg.compare(e.name, reflect.ValueOf(expected), reflect.ValueOf(e.value)); !ok {
		e.t.Errorf("%s", msg)
	}

	return e
}

// compare walks expected (exp) and actual (act) in lock-step and returns a message
// describing the first difference together with false, or an empty message and true
// when everything is within tolerance.
func (c aboutConfig) compare(path string, exp, act reflect.Value) (string, bool) {
	if exp.Type() != act.Type() {
		return fmt.Sprintf("expected %s to be of type %v but it is of type %v", path, exp.Type(), act.Type()), false
	}

	// time.Time is a struct, so it has to be caught before the struct branch would
	// recurse into its (unexported, location/monotonic carrying) fields.
	if exp.Type() == timeType && exp.CanInterface() {
		et := exp.Interface().(time.Time)
		at := act.Interface().(time.Time)
		if d := absDuration(at.Sub(et)); d > c.timeDelta {
			return fmt.Sprintf("expected %s to be %s±%v but it is %s", path,
				et.Format(time.RFC3339Nano), c.timeDelta, at.Format(time.RFC3339Nano)), false
		}

		return "", true
	}

	// time.Duration has Kind Int64, so it must be distinguished by type, not by kind,
	// otherwise the duration delta would leak onto every int64 value.
	if exp.Type() == durationType {
		ed, ad := time.Duration(exp.Int()), time.Duration(act.Int())
		if d := absDuration(ad - ed); d > c.durationDelta {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ed, c.durationDelta, ad), false
		}

		return "", true
	}

	switch exp.Kind() {
	case reflect.Float32, reflect.Float64:
		ef, af := exp.Float(), act.Float()
		if math.Abs(af-ef) > c.floatDelta {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ef, c.floatDelta, af), false
		}

		return "", true

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ei, ai := exp.Int(), act.Int()
		if absInt64(ai-ei) > c.intDelta {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ei, c.intDelta, ai), false
		}

		return "", true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		eu, au := exp.Uint(), act.Uint()
		if absDiffUint64(eu, au) > uint64(c.intDelta) {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, eu, c.intDelta, au), false
		}

		return "", true

	case reflect.Ptr, reflect.Interface:
		if exp.IsNil() || act.IsNil() {
			if exp.IsNil() != act.IsNil() {
				return fmt.Sprintf("expected %s to be %v but it is %v", path, formatV(exp), formatV(act)), false
			}

			return "", true
		}

		return c.compare(path, exp.Elem(), act.Elem())

	case reflect.Struct:
		for i := 0; i < exp.NumField(); i++ {
			name := exp.Type().Field(i).Name
			if msg, ok := c.compare(path+"."+name, exp.Field(i), act.Field(i)); !ok {
				return msg, false
			}
		}

		return "", true

	case reflect.Slice, reflect.Array:
		if exp.Len() != act.Len() {
			return fmt.Sprintf("expected %s to have %d elements but it has %d elements", path, exp.Len(), act.Len()), false
		}

		for i := 0; i < exp.Len(); i++ {
			if msg, ok := c.compare(fmt.Sprintf("%s[%d]", path, i), exp.Index(i), act.Index(i)); !ok {
				return msg, false
			}
		}

		return "", true

	case reflect.Map:
		if exp.Len() != act.Len() {
			return fmt.Sprintf("expected %s to have %d entries but it has %d entries", path, exp.Len(), act.Len()), false
		}

		iter := exp.MapRange()
		for iter.Next() {
			av := act.MapIndex(iter.Key())
			if !av.IsValid() {
				return fmt.Sprintf("expected %s to have key %v but it does not", path, formatV(iter.Key())), false
			}

			if msg, ok := c.compare(fmt.Sprintf("%s[%v]", path, formatV(iter.Key())), iter.Value(), av); !ok {
				return msg, false
			}
		}

		return "", true

	default:
		if !directEqual(exp, act) {
			return fmt.Sprintf("expected %s to be %v but it is %v", path, formatV(exp), formatV(act)), false
		}

		return "", true
	}
}

// directEqual compares leaf values that carry no tolerance (strings, bools, complex,
// channels, funcs, ...). It prefers reflect.DeepEqual but stays usable on unexported
// fields, where Interface() would panic.
func directEqual(exp, act reflect.Value) bool {
	if exp.CanInterface() && act.CanInterface() {
		return reflect.DeepEqual(exp.Interface(), act.Interface())
	}

	switch exp.Kind() {
	case reflect.String:
		return exp.String() == act.String()
	case reflect.Bool:
		return exp.Bool() == act.Bool()
	case reflect.Complex64, reflect.Complex128:
		return exp.Complex() == act.Complex()
	default:
		return fmt.Sprintf("%v", exp) == fmt.Sprintf("%v", act)
	}
}

func formatV(v reflect.Value) string {
	if v.Kind() == reflect.String {
		return "'" + v.String() + "'"
	}

	return fmt.Sprintf("%v", v)
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}

	return d
}

func absInt64(v int64) int64 {
	if v < 0 {
		return -v
	}

	return v
}

func absDiffUint64(a, b uint64) uint64 {
	if a > b {
		return a - b
	}

	return b - a
}
