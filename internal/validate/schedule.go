package validate

import (
	"fmt"
	"slices"
	"strings"
	"time"
	_ "time/tzdata" // the image (distroless) and Windows hosts carry no zone database
)

// The grammar of a policy's schedule (amendment A43): days of the week, a
// window between two times of day, and the IANA time zone whose wall clock
// both are read on.

// Weekdays are the days as a schedule spells them, in the week's order.
var Weekdays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// Window is a parsed schedule.
type Window struct {
	// Days are the days the window starts on, by time.Weekday.
	Days [7]bool
	// Start and End are minutes after midnight; an End before Start ends
	// the next day. An AllDay window has neither: each of Days is whole.
	Start, End int
	AllDay     bool
}

// weekday is Weekdays[i] as a time.Weekday (Sunday is 0).
func weekday(i int) time.Weekday { return time.Weekday((i + 1) % 7) }

// Schedule validates a schedule: days from Weekdays, each once (none is
// every day), and a window from start to end, two Clock times, or neither
// for whole days. A schedule with neither days nor times would always
// apply, which is no schedule: refused.
func Schedule(field string, days []string, start, end string) (Window, error) {
	var w Window
	if len(days) == 0 {
		w.Days = [7]bool{true, true, true, true, true, true, true}
	}
	for i, d := range days {
		j := slices.Index(Weekdays, d)
		if j < 0 {
			return Window{}, fail(fmt.Sprintf("%s.days[%d]", field, i), d, "must be one of "+strings.Join(Weekdays, ", "))
		}
		if w.Days[weekday(j)] {
			return Window{}, fail(field+".days", d, "lists a day twice")
		}
		w.Days[weekday(j)] = true
	}
	switch {
	case start == "" && end == "":
		if len(days) == 0 {
			return Window{}, fail(field, "", "a schedule needs days, a start and an end time, or both; without them the policy always applies, so leave the schedule out")
		}
		w.AllDay = true
		return w, nil
	case start == "" || end == "":
		return Window{}, fail(field, start+"-"+end, "give both a start and an end time, or neither for whole days")
	}
	var err error
	if w.Start, err = Clock(field+".start", start); err != nil {
		return Window{}, err
	}
	if w.End, err = Clock(field+".end", end); err != nil {
		return Window{}, err
	}
	if w.Start == w.End {
		return Window{}, fail(field, start+"-"+end, "the window ends where it starts; leave both times out for whole days")
	}
	return w, nil
}

// Clock validates a time of day, HH:MM on a 24-hour clock, and returns it
// as minutes after midnight.
func Clock(field, s string) (int, error) {
	h, m, ok := strings.Cut(s, ":")
	two := func(v string, limit int) (int, bool) {
		if len(v) != 2 || v[0] < '0' || v[0] > '9' || v[1] < '0' || v[1] > '9' {
			return 0, false
		}
		n := int(v[0]-'0')*10 + int(v[1]-'0')
		return n, n <= limit
	}
	hh, okH := two(h, 23)
	mm, okM := two(m, 59)
	if !ok || !okH || !okM {
		return 0, fail(field, s, "must be a time of day HH:MM on a 24-hour clock, such as 09:00 or 22:30")
	}
	return hh*60 + mm, nil
}

// Holds reports whether the window holds at t, read on t's own wall clock:
// pass t.In(zone). A window that ends the next day holds after midnight
// when it started the day before.
func (w Window) Holds(t time.Time) bool {
	day, m := t.Weekday(), t.Hour()*60+t.Minute()
	switch {
	case w.AllDay:
		return w.Days[day]
	case w.Start < w.End:
		return w.Days[day] && w.Start <= m && m < w.End
	}
	return (w.Days[day] && m >= w.Start) || (w.Days[(day+6)%7] && m < w.End)
}

// Edges are the times of day (minutes after midnight) at which whether the
// window holds can change: midnight for whole days, else its start and end.
func (w Window) Edges() []int {
	if w.AllDay {
		return []int{0}
	}
	return []int{w.Start, w.End}
}

// DayNames are the days in the week's order, nil for every day: the
// canonical form of a schedule's days.
func (w Window) DayNames() []string {
	var out []string
	for i, d := range Weekdays {
		if w.Days[weekday(i)] {
			out = append(out, d)
		}
	}
	if len(out) == len(Weekdays) {
		return nil
	}
	return out
}

// String is the window as people read it: "mon-fri 09:00-18:00",
// "sat,sun", "daily 22:00-06:00". Three days or more in a row are a range.
func (w Window) String() string {
	days := "daily"
	if names := w.DayNames(); names != nil {
		var parts []string
		for i := 0; i < len(Weekdays); {
			if !w.Days[weekday(i)] {
				i++
				continue
			}
			j := i
			for j+1 < len(Weekdays) && w.Days[weekday(j+1)] {
				j++
			}
			if j-i >= 2 {
				parts = append(parts, Weekdays[i]+"-"+Weekdays[j])
			} else {
				parts = append(parts, Weekdays[i:j+1]...)
			}
			i = j + 1
		}
		days = strings.Join(parts, ",")
	}
	if w.AllDay {
		return days
	}
	return fmt.Sprintf("%s %02d:%02d-%02d:%02d", days, w.Start/60, w.Start%60, w.End/60, w.End%60)
}

// TimeZone validates an IANA time zone name -- "UTC", "Europe/Berlin",
// "America/Argentina/Buenos_Aires" -- and loads it, from the host's zone
// database or else the one built into the binary. "Local" is refused: it
// would mean a different zone on every host.
func TimeZone(field, s string) (*time.Location, error) {
	ok := len(s) <= 64 && s != "Local"
	for _, part := range strings.Split(s, "/") {
		ok = ok && part != "" && len(part) <= 32 && (part[0] >= 'A' && part[0] <= 'Z' || part[0] >= 'a' && part[0] <= 'z')
		for i := 0; ok && i < len(part); i++ {
			c := part[i]
			ok = c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '+'
		}
	}
	if !ok {
		return nil, fail(field, s, "must be an IANA time zone name such as UTC, Europe/Berlin or Asia/Singapore")
	}
	loc, err := time.LoadLocation(s)
	if err != nil {
		return nil, fail(field, s, "is not a time zone this build knows; use an IANA name such as Asia/Singapore")
	}
	return loc, nil
}

// Timestamp validates a moment in RFC 3339 form, such as
// 2026-10-23T09:30:00+08:00.
func Timestamp(field, s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil || len(s) > 64 {
		return time.Time{}, fail(field, s, "must be a time in RFC 3339 form, such as 2026-10-23T09:30:00+08:00")
	}
	return t, nil
}
