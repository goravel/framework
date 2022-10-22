package time

import "time"

var now *time.Time

func SetTestNow(t time.Time) {
	now = &t
}

func Now() time.Time {
	if now == nil {
		return time.Now()
	}

	return *now
}
