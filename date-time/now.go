package datetime

import "time"

var nowFn = time.Now

func Now() time.Time {
	return nowFn()
}

func SetNowFunc(fn func() time.Time) {
	nowFn = fn
}

func ResetNowFunc() {
	nowFn = time.Now
}
