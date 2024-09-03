package defaults

func Bool2Str(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func Bool2StrBy(v bool, true, false string) string {
	if v {
		return true
	}
	return false
}
