package defaults

type Stringer interface {
	String() string
}

// Dec return default value (second param), if first param is default type value
func Dec[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](val, defaultVal T) T {
	if val == 0 {
		return defaultVal
	}
	return val
}

// LinkDec return default value (second param), if first param is nil
func LinkDec[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](val *T, defaultVal T) T {
	if val == nil {
		return defaultVal
	}
	return *val
}

// Str return default value (second param), if first param is default type value
func Str[T ~string](val, defaultVal T) T {
	if val == "" {
		return defaultVal
	}
	return val
}

// LinkStr return default value (second param), if first param is nil
func LinkStr[T ~string](val *T, defaultVal T) T {
	if val == nil {
		return defaultVal
	}
	return *val
}

// Bool return default value (second param), if first param is default type value
func Bool(val, defaultVal bool) bool {
	if val == defaultVal {
		return defaultVal
	}
	return val
}

// Link2String return default value (second param), if first param is nil
func Link2String(val Stringer, defaultVal string) string {
	if val == nil {
		return defaultVal
	}
	return val.String()
}
