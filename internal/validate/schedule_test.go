package validate

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	for _, c := range []struct {
		days       []string
		start, end string
		text       string
		want       []string // DayNames
	}{
		{[]string{"fri", "mon", "tue", "wed", "thu"}, "09:00", "18:00", "mon-fri 09:00-18:00", []string{"mon", "tue", "wed", "thu", "fri"}},
		{[]string{"sun", "sat"}, "", "", "sat,sun", []string{"sat", "sun"}},
		{nil, "22:00", "06:00", "daily 22:00-06:00", nil},
		{[]string{"mon", "wed", "thu", "fri", "sun"}, "00:00", "23:59", "mon,wed-fri,sun 00:00-23:59", []string{"mon", "wed", "thu", "fri", "sun"}},
		{Weekdays, "", "", "daily", nil},
	} {
		w, err := Schedule("s", c.days, c.start, c.end)
		if err != nil {
			t.Fatalf("Schedule(%v, %q, %q) = %v", c.days, c.start, c.end, err)
		}
		if w.String() != c.text || !reflect.DeepEqual(w.DayNames(), c.want) {
			t.Errorf("Schedule(%v, %q, %q) reads %q, days %v; want %q, %v", c.days, c.start, c.end, w, w.DayNames(), c.text, c.want)
		}
	}
	for _, c := range []struct {
		days       []string
		start, end string
		want       string
	}{
		{nil, "", "", "without them the policy always applies"},
		{[]string{"monday"}, "", "", "s.days[0]: must be one of mon, tue"},
		{[]string{"mon", "mon"}, "", "", "lists a day twice"},
		{[]string{"Mon"}, "", "", "must be one of"},
		{nil, "09:00", "", "both a start and an end"},
		{nil, "", "18:00", "both a start and an end"},
		{nil, "09:00", "09:00", "ends where it starts"},
		{nil, "9:00", "18:00", "s.start: must be a time of day HH:MM"},
		{nil, "09:00", "24:00", "s.end: must be a time of day"},
		{nil, "09:60", "18:00", "s.start"},
		{nil, "+1:00", "18:00", "s.start"},
		{nil, "09:00\n", "18:00", "s.start"},
		{nil, "09.00", "18:00", "s.start"},
	} {
		if _, err := Schedule("s", c.days, c.start, c.end); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Schedule(%v, %q, %q) = %v, want an error about %q", c.days, c.start, c.end, err, c.want)
		}
	}
}

func TestWindowHolds(t *testing.T) {
	at := func(day, hhmm string) time.Time {
		// 2026-10-19 is a Monday.
		d := map[string]int{"mon": 19, "tue": 20, "wed": 21, "thu": 22, "fri": 23, "sat": 24, "sun": 25}[day]
		m, _ := Clock("t", hhmm)
		return time.Date(2026, 10, d, m/60, m%60, 0, 0, time.UTC)
	}
	office, _ := Schedule("s", []string{"mon", "tue", "wed", "thu", "fri"}, "09:00", "18:00")
	night, _ := Schedule("s", []string{"fri"}, "22:00", "06:00")
	weekend, _ := Schedule("s", []string{"sat", "sun"}, "", "")
	for _, c := range []struct {
		name string
		w    Window
		day  string
		at   string
		want bool
	}{
		{"office opens", office, "mon", "09:00", true},
		{"office before", office, "mon", "08:59", false},
		{"office closes", office, "fri", "18:00", false},
		{"office last minute", office, "fri", "17:59", true},
		{"office on saturday", office, "sat", "10:00", false},
		{"night starts", night, "fri", "22:00", true},
		{"night before", night, "fri", "21:59", false},
		{"night after midnight", night, "sat", "05:59", true},
		{"night ends", night, "sat", "06:00", false},
		{"night belongs to the day it starts", night, "fri", "05:00", false},
		{"weekend midnight", weekend, "sat", "00:00", true},
		{"weekend sunday night", weekend, "sun", "23:59", true},
		{"weekend over", weekend, "mon", "00:00", false},
	} {
		if got := c.w.Holds(at(c.day, c.at)); got != c.want {
			t.Errorf("%s: Holds(%s %s) = %v", c.name, c.day, c.at, got)
		}
	}
	if !reflect.DeepEqual(office.Edges(), []int{540, 1080}) || !reflect.DeepEqual(weekend.Edges(), []int{0}) {
		t.Errorf("edges %v, %v", office.Edges(), weekend.Edges())
	}
}

func TestTimeZone(t *testing.T) {
	for _, s := range []string{"UTC", "Asia/Singapore", "America/Argentina/Buenos_Aires", "Etc/GMT+8", "America/Port-au-Prince"} {
		loc, err := TimeZone("tz", s)
		if err != nil || loc.String() != s {
			t.Errorf("TimeZone(%q) = %v, %v", s, loc, err)
		}
	}
	for s, want := range map[string]string{
		"":                 "IANA time zone name",
		"Local":            "IANA time zone name",
		"../etc/passwd":    "IANA time zone name",
		"/etc/localtime":   "IANA time zone name",
		"Asia//Singapore":  "IANA time zone name",
		"Asia/Singapore\n": "IANA time zone name",
		"Asia/Singapore ":  "IANA time zone name",
		"1Asia/Singapore":  "IANA time zone name",
		"Mars/Olympus":     "not a time zone this build knows",
	} {
		if _, err := TimeZone("tz", s); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("TimeZone(%q) = %v, want an error about %q", s, err, want)
		}
	}
}

func TestTimestamp(t *testing.T) {
	got, err := Timestamp("at", "2026-10-23T09:30:00+08:00")
	if err != nil || !got.Equal(time.Date(2026, 10, 23, 1, 30, 0, 0, time.UTC)) {
		t.Fatalf("Timestamp = %v, %v", got, err)
	}
	for _, s := range []string{"", "2026-10-23 09:30", "tomorrow", "2026-10-23T09:30:00+08:00\n"} {
		if _, err := Timestamp("at", s); err == nil || !strings.Contains(err.Error(), "RFC 3339") {
			t.Errorf("Timestamp(%q) = %v", s, err)
		}
	}
}
