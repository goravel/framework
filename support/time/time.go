package time

import "time"

type virtualTime struct {
	virtual time.Time
	real    time.Time
}

var now *virtualTime

func SetTestNow(t ...time.Time) {
	if len(t) == 0 {
		now = nil
		return
	}

	now = &virtualTime{
		virtual: t[0],
		real:    time.Now(),
	}
}

func Now() time.Time {
	if now == nil {
		return time.Now()
	}

	diff := time.Since(now.real)

	return now.virtual.Add(diff)
}
