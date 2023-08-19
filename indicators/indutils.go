package indicators

func IsBiggerIn(dst Decimal, ts *TimeSeries, opt GetValue, low, high int) bool {
	if ts.LastIndex() < high {
		return false
	}
	for off := low; off <= high; off++ {
		d := opt(ts.Get(off))
		if dst.LTE(d) {
			return false
		}
	}
	return true
}

func IsLessIn(dst Decimal, ts *TimeSeries, opt GetValue, low, high int) bool {
	if ts.LastIndex() < high {
		return false
	}
	for off := low; off <= high; off++ {
		d := opt(ts.Get(off))
		if dst.GTE(d) {
			return false
		}
	}
	return true
}

func IsBiggerInDec(dst Decimal, src []Decimal, low, high int) bool {
	if len(src) <= high {
		return false
	}
	for off := low; off <= high; off++ {
		if dst.LTE(src[off]) {
			return false
		}
	}
	return true
}

func IsLessInDec(dst Decimal, src []Decimal, low, high int) bool {
	if len(src) <= high {
		return false
	}
	for off := low; off <= high; off++ {
		if dst.GTE(src[off]) {
			return false
		}
	}
	return true
}
