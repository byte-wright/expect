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

type aboutWalk struct {
	deltas  aboutConfig
	visited map[visit]bool
}

type visit struct {
	typ      reflect.Type
	exp, act uintptr
	length   int
	// A pointer reached from an unexported field is compared without deltas.
	unexported bool
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
// The deltas apply to the value itself and to exported fields, list items and map
// values at any depth, also inside unexported fields. Only a leaf that is itself an
// unexported field is compared exactly.
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

	w := aboutWalk{deltas: cfg, visited: map[visit]bool{}}
	if msg, ok := w.compare(e.name, reflect.ValueOf(expected), reflect.ValueOf(e.value), false); !ok {
		e.t.Errorf("%s", msg)
	}

	return e
}

// compare walks expected (exp) and actual (act) in lock-step and returns a message
// describing the first difference together with false, or an empty message and true
// when everything is within tolerance.
func (w aboutWalk) compare(path string, exp, act reflect.Value, unexported bool) (string, bool) {
	if exp.Type() != act.Type() {
		return fmt.Sprintf("expected %s to be of type %v but it is of type %v", path, exp.Type(), act.Type()), false
	}

	// A pair reached again is either still being compared further up a cycle or has
	// already passed, since the walk stops at the first difference.
	if w.revisited(exp, act, unexported) {
		return "", true
	}

	c := w.deltas
	if unexported {
		c = aboutConfig{}
	}

	// time.Time is a struct, so it has to be caught before the struct branch would
	// recurse into its (unexported, location/monotonic carrying) fields.
	if exp.Type() == timeType {
		et := exp.Interface().(time.Time)
		at := act.Interface().(time.Time)
		if !withinDelta(absDiffTime(et, at), int64(c.timeDelta)) {
			return fmt.Sprintf("expected %s to be %s±%v but it is %s", path,
				et.Format(time.RFC3339Nano), c.timeDelta, at.Format(time.RFC3339Nano)), false
		}

		return "", true
	}

	// time.Duration has Kind Int64, so it must be distinguished by type, not by kind,
	// otherwise the duration delta would leak onto every int64 value.
	if exp.Type() == durationType {
		ed, ad := time.Duration(exp.Int()), time.Duration(act.Int())
		if !withinDelta(absDiffInt64(int64(ed), int64(ad)), int64(c.durationDelta)) {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ed, c.durationDelta, ad), false
		}

		return "", true
	}

	switch exp.Kind() {
	case reflect.Float32, reflect.Float64:
		ef, af := exp.Float(), act.Float()
		if !withinFloatDelta(ef, af, c.floatDelta) {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ef, c.floatDelta, af), false
		}

		return "", true

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ei, ai := exp.Int(), act.Int()
		if !withinDelta(absDiffInt64(ei, ai), c.intDelta) {
			return fmt.Sprintf("expected %s to be %v±%v but it is %v", path, ei, c.intDelta, ai), false
		}

		return "", true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		eu, au := exp.Uint(), act.Uint()
		if !withinDelta(absDiffUint64(eu, au), c.intDelta) {
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

		return w.compare(path, exp.Elem(), act.Elem(), unexported)

	case reflect.Struct:
		// Interface() refuses values read through unexported fields, but the time and
		// leaf comparisons need it. Reading such fields through their address avoids
		// that, and needs an addressable struct.
		exp, act = addressable(exp), addressable(act)

		for i := 0; i < exp.NumField(); i++ {
			field := exp.Type().Field(i)
			ef, af := exp.Field(i), act.Field(i)

			if !field.IsExported() {
				ef, af = export(ef), export(af)
			}

			if msg, ok := w.compare(path+"."+field.Name, ef, af, !field.IsExported()); !ok {
				return msg, false
			}
		}

		return "", true

	case reflect.Slice, reflect.Array:
		if exp.Len() != act.Len() {
			return fmt.Sprintf("expected %s to have %d elements but it has %d elements", path, exp.Len(), act.Len()), false
		}

		for i := 0; i < exp.Len(); i++ {
			if msg, ok := w.compare(fmt.Sprintf("%s[%d]", path, i), exp.Index(i), act.Index(i), false); !ok {
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

			if msg, ok := w.compare(fmt.Sprintf("%s[%v]", path, formatV(iter.Key())), iter.Value(), av, false); !ok {
				return msg, false
			}
		}

		return "", true

	default:
		if !reflect.DeepEqual(exp.Interface(), act.Interface()) {
			return fmt.Sprintf("expected %s to be %v but it is %v", path, formatV(exp), formatV(act)), false
		}

		return "", true
	}
}

func (w aboutWalk) revisited(exp, act reflect.Value, unexported bool) bool {
	key := visit{typ: exp.Type(), unexported: unexported}

	switch exp.Kind() {
	case reflect.Ptr, reflect.Map:
		key.exp, key.act = exp.Pointer(), act.Pointer()
	case reflect.Slice:
		// s[:1] and s[:2] share their pointer.
		key.exp, key.act, key.length = exp.Pointer(), act.Pointer(), exp.Len()
	default:
		return false
	}

	if w.visited[key] {
		return true
	}

	w.visited[key] = true

	return false
}

func addressable(v reflect.Value) reflect.Value {
	if v.CanAddr() {
		return v
	}

	copied := reflect.New(v.Type()).Elem()
	copied.Set(v)

	return copied
}

func export(v reflect.Value) reflect.Value {
	return reflect.NewAt(v.Type(), v.Addr().UnsafePointer()).Elem()
}

func formatV(v reflect.Value) string {
	if v.Kind() == reflect.String {
		return "'" + v.String() + "'"
	}

	return fmt.Sprintf("%v", v)
}

func withinFloatDelta(exp, act, delta float64) bool {
	// Inf-Inf is NaN, so equal infinities need the explicit equality check.
	if exp == act || (math.IsNaN(exp) && math.IsNaN(act)) {
		return true
	}

	return math.Abs(act-exp) <= delta
}

func withinDelta(diff uint64, delta int64) bool {
	if delta < 0 {
		delta = 0
	}

	return diff <= uint64(delta)
}

func absDiffTime(a, b time.Time) uint64 {
	if a.After(b) {
		return uint64(a.Sub(b))
	}

	return uint64(b.Sub(a))
}

func absDiffInt64(a, b int64) uint64 {
	if a > b {
		return uint64(a) - uint64(b)
	}

	return uint64(b) - uint64(a)
}

func absDiffUint64(a, b uint64) uint64 {
	if a > b {
		return a - b
	}

	return b - a
}
