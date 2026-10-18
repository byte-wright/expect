package expect

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// AboutOption configures the per-type tolerances used by ToBeAbout.
type AboutOption func(*aboutConfig)

// A nil delta means the option was not given. For time.Time that differs from a delta
// of 0, since only TimeDelta compares times by instant.
type aboutConfig struct {
	floatDelta    *float64
	intDelta      *int64
	timeDelta     *time.Duration
	durationDelta *time.Duration
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

// FloatDelta sets the maximum allowed absolute difference for float32 and float64
// values. NaN never matches, as in ToBe.
func FloatDelta(d float64) AboutOption {
	return func(c *aboutConfig) { c.floatDelta = &d }
}

// IntDelta sets the maximum allowed absolute difference for integers of every size:
// int, int8, int16, int32 (rune), int64, uint, uint8 (byte), uint16, uint32 and uint64.
// Named integer types such as enums get the delta too: with IntDelta(5), time.June
// matches time.January. The exceptions are time.Duration, which uses DurationDelta, and
// uintptr, which is compared exactly since it holds addresses and handles.
func IntDelta(d int64) AboutOption {
	return func(c *aboutConfig) { c.intDelta = &d }
}

// TimeDelta sets the maximum allowed absolute difference for time.Time values. Times
// are then compared by instant (a.Sub(b)), so location and monotonic clock reading are
// ignored, even with a delta of 0. Without it, times are compared like ToBe does.
func TimeDelta(d time.Duration) AboutOption {
	return func(c *aboutConfig) { c.timeDelta = &d }
}

// DurationDelta sets the maximum allowed absolute difference for time.Duration values.
func DurationDelta(d time.Duration) AboutOption {
	return func(c *aboutConfig) { c.durationDelta = &d }
}

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// ToBeAbout asserts that the value deeply equals expected, exactly like ToBe, except
// that values of a type given as an option only need to be within its delta. Deltas
// apply to the value itself, to exported fields and to the elements of slices, arrays
// and maps, at any depth and through pointers and interfaces. Only a value held by an
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

// compare walks expected (exp) and actual (act) in lock-step the way reflect.DeepEqual
// does and returns a message describing the first difference together with false, or
// an empty message and true when everything is within tolerance.
func (w aboutWalk) compare(path string, exp, act reflect.Value, unexported bool) (string, bool) {
	if exp.Type() != act.Type() {
		return fmt.Sprintf("expected %s to be of type %v but it is of type %v", path, exp.Type(), act.Type()), false
	}

	// A pair reached again is either still being compared further up a cycle or has
	// already passed, since the walk stops at the first difference.
	if w.revisited(exp, act, unexported) {
		return "", true
	}

	switch exp.Kind() {
	case reflect.Array:
		return w.compareElements(path, exp, act)

	case reflect.Slice:
		return w.compareSlice(path, exp, act)

	case reflect.Map:
		return w.compareMap(path, exp, act)

	case reflect.Interface:
		if exp.IsNil() != act.IsNil() {
			return mismatch(path, exp, act, "")
		}

		if exp.IsNil() {
			return "", true
		}

		return w.compare(path, exp.Elem(), act.Elem(), unexported)

	case reflect.Ptr:
		if exp.Pointer() == act.Pointer() {
			return "", true
		}

		if exp.IsNil() || act.IsNil() {
			return mismatch(path, exp, act, "")
		}

		return w.compare(path, exp.Elem(), act.Elem(), unexported)

	case reflect.Struct:
		// time.Time is compared as a whole, not by its unexported fields.
		if exp.Type() != timeType {
			return w.compareStruct(path, exp, act)
		}
	}

	if unexported {
		return compareValue(path, exp, act, aboutConfig{})
	}

	return compareValue(path, exp, act, w.deltas)
}

func (w aboutWalk) compareSlice(path string, exp, act reflect.Value) (string, bool) {
	if exp.IsNil() != act.IsNil() {
		return mismatch(path, exp, act, "")
	}

	if exp.Len() != act.Len() {
		return fmt.Sprintf("expected %s to have %d elements but it has %d elements", path, exp.Len(), act.Len()), false
	}

	if exp.Pointer() == act.Pointer() {
		return "", true
	}

	return w.compareElements(path, exp, act)
}

func (w aboutWalk) compareElements(path string, exp, act reflect.Value) (string, bool) {
	for i := 0; i < exp.Len(); i++ {
		if msg, ok := w.compare(fmt.Sprintf("%s[%d]", path, i), exp.Index(i), act.Index(i), false); !ok {
			return msg, false
		}
	}

	return "", true
}

func (w aboutWalk) compareMap(path string, exp, act reflect.Value) (string, bool) {
	if exp.IsNil() != act.IsNil() {
		return mismatch(path, exp, act, "")
	}

	if exp.Len() != act.Len() {
		return fmt.Sprintf("expected %s to have %d entries but it has %d entries", path, exp.Len(), act.Len()), false
	}

	if exp.Pointer() == act.Pointer() {
		return "", true
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
}

func (w aboutWalk) compareStruct(path string, exp, act reflect.Value) (string, bool) {
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
}

// compareValue compares the values compare does not walk into.
func compareValue(path string, exp, act reflect.Value, deltas aboutConfig) (string, bool) {
	// time.Duration has Kind Int64, so it must be distinguished by type, not by kind,
	// otherwise the int delta would apply to durations.
	switch exp.Type() {
	case timeType:
		return compareTime(path, exp, act, deltas.timeDelta)
	case durationType:
		return withinDelta(path, exp, act, absDiffInt64(exp.Int(), act.Int()), deltas.durationDelta)
	}

	switch exp.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return withinDelta(path, exp, act, absDiffInt64(exp.Int(), act.Int()), deltas.intDelta)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return withinDelta(path, exp, act, absDiffUint64(exp.Uint(), act.Uint()), deltas.intDelta)
	case reflect.Float32, reflect.Float64:
		return withinFloatDelta(path, exp, act, deltas.floatDelta)
	}

	if !reflect.DeepEqual(exp.Interface(), act.Interface()) {
		return mismatch(path, exp, act, "")
	}

	return "", true
}

func compareTime(path string, exp, act reflect.Value, delta *time.Duration) (string, bool) {
	et, at := exp.Interface().(time.Time), act.Interface().(time.Time)

	if delta != nil {
		return withinDelta(path, exp, act, absDiffTime(et, at), delta)
	}

	if reflect.DeepEqual(et, at) {
		return "", true
	}

	msg, _ := mismatch(path, exp, act, "")
	if et.Equal(at) {
		msg += " (same instant, TimeDelta would ignore location and monotonic clock)"
	}

	return msg, false
}

func withinDelta[D ~int64](path string, exp, act reflect.Value, diff uint64, delta *D) (string, bool) {
	if diff == 0 || (delta != nil && *delta >= 0 && diff <= uint64(*delta)) {
		return "", true
	}

	return mismatch(path, exp, act, formatDelta(delta))
}

func withinFloatDelta(path string, exp, act reflect.Value, delta *float64) (string, bool) {
	// Inf-Inf is NaN, so equal infinities need the explicit equality check.
	if ef, af := exp.Float(), act.Float(); ef == af || (delta != nil && math.Abs(af-ef) <= *delta) {
		return "", true
	}

	return mismatch(path, exp, act, formatDelta(delta))
}

func mismatch(path string, exp, act reflect.Value, delta string) (string, bool) {
	return fmt.Sprintf("expected %s to be %s%s but it is %s", path, formatV(exp), delta, formatV(act)), false
}

func formatDelta[D ~float64 | ~int64](delta *D) string {
	if delta == nil {
		return ""
	}

	return fmt.Sprintf("±%v", *delta)
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
	if v.Type() == timeType {
		return v.Interface().(time.Time).Format(time.RFC3339Nano)
	}

	switch v.Kind() {
	case reflect.String:
		return "'" + v.String() + "'"
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if v.IsNil() {
			return "nil"
		}
	}

	return fmt.Sprintf("%v", v)
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
