package expect

import "reflect"

func sameType(a, b interface{}) bool {
	if a == nil || b == nil {
		return isNil(a) && isNil(b)
	}

	at := reflect.TypeOf(a)
	bt := reflect.TypeOf(b)

	return at == bt
}

func typeName(a interface{}) string {
	if a == nil {
		return "<nil>"
	}

	return reflect.TypeOf(a).String()
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}

	switch reflect.TypeOf(a).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(a).IsNil()
	}

	return false
}
