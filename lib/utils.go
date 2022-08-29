package lib

import "reflect"

func If(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}
