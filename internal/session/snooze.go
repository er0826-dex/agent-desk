package session

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Snooze: park a session until a wall-clock time, then surface it as unread.
//
// The flag/time pair lives on the Instance (Flag == FlagSnoozed, SnoozeUntil
// set) and persists in tool_data. The home tick calls SnoozeDue and
// WakeFromSnooze; nothing else changes the pair except the s key (set), the
// u key and attaching (both cancel).

// DefaultWakeHour is the hour used when a snooze names a day but no time
// ("tomorrow", "mon", "10/12"). 9am, the same "morning" the suggestions use.
const DefaultWakeHour = 9

// SnoozePreset is one suggestion row of the s-key picker: a label and the
// time it resolves to for the moment it was built.
type SnoozePreset struct {
	Label string
	At    time.Time
}

// snoozeMinLead keeps a suggestion from being almost now.
const snoozeMinLead = 30 * time.Minute

// SnoozeSuggestions returns the picker rows for now. They are landmarks in
// the week rather than durations, because that is how Eric snoozes (10/12:
// "push something off to the weekend, till next Monday, till later in the
// week"); durations and odd times are typed.
//
//	Tonight           today 8pm, until 7:30pm
//	Later this week   Thu 9am on Mon/Tue, Fri 9am on Wed/Thu, absent Fri–Sun
//	This weekend      Sat 9am Mon–Fri (and early Sat), Sun 9am on Sat, absent Sun
//	Next Monday       the Monday after today, 9am
//	In 2 weeks        today + 14 days, 9am
//	In a month        today + 1 month, 9am
func SnoozeSuggestions(now time.Time) []SnoozePreset {
	loc := now.Location()
	y, m, d := now.Date()
	at := func(dayOffset, hour int) time.Time {
		return time.Date(y, m, d+dayOffset, hour, 0, 0, 0, loc)
	}
	wd := now.Weekday()
	daysUntil := func(target time.Weekday) int { // strictly after today
		delta := (int(target) - int(wd) + 7) % 7
		if delta == 0 {
			delta = 7
		}
		return delta
	}

	var rows []SnoozePreset
	add := func(label string, t time.Time) {
		if t.Sub(now) < snoozeMinLead {
			return
		}
		for _, r := range rows {
			if r.At.Equal(t) {
				return
			}
		}
		rows = append(rows, SnoozePreset{Label: label, At: t})
	}

	add("Tonight", at(0, 20))

	switch wd {
	case time.Monday, time.Tuesday:
		add("Later this week", at(daysUntil(time.Thursday), DefaultWakeHour))
	case time.Wednesday, time.Thursday:
		add("Later this week", at(daysUntil(time.Friday), DefaultWakeHour))
	}

	switch wd {
	case time.Saturday:
		if sat := at(0, DefaultWakeHour); sat.Sub(now) >= snoozeMinLead {
			add("This weekend", sat)
		} else {
			add("This weekend", at(1, DefaultWakeHour))
		}
	case time.Sunday:
		// no weekend row
	default:
		add("This weekend", at(daysUntil(time.Saturday), DefaultWakeHour))
	}

	add("Next Monday", at(daysUntil(time.Monday), DefaultWakeHour))
	add("In 2 weeks", at(14, DefaultWakeHour))
	month := now.AddDate(0, 1, 0)
	add("In a month", time.Date(month.Year(), month.Month(), month.Day(), DefaultWakeHour, 0, 0, 0, loc))
	return rows
}

// Snooze parks the session until the given time. Setting a marker mutes the
// session's auto-attention, as the u key does.
func (i *Instance) Snooze(until time.Time) {
	i.Flag = FlagSnoozed
	i.SnoozeUntil = until
	i.Acknowledge()
}

// ClearSnooze cancels a pending snooze. The flag drops to none only if the
// session was snoozed; other markers are left alone.
func (i *Instance) ClearSnooze() {
	if i.Flag == FlagSnoozed {
		i.Flag = FlagNone
	}
	i.SnoozeUntil = time.Time{}
}

// SnoozeDue reports whether a snoozed session's wake time has passed.
func (i *Instance) SnoozeDue(now time.Time) bool {
	return i.Flag == FlagSnoozed && !i.SnoozeUntil.IsZero() && !now.Before(i.SnoozeUntil)
}

// WakeFromSnooze flips a due session to the unread bookmark ("come back to
// this") and clears the wake time. Grouping is deliberately untouched.
func (i *Instance) WakeFromSnooze() {
	i.Flag = FlagUnread
	i.SnoozeUntil = time.Time{}
}

var (
	reDuration = regexp.MustCompile(`^(\d+)\s*(m|min|mins|minutes?|h|hr|hrs|hours?|d|days?|w|wk|weeks?)$`)
	reClock    = regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?\s*(am|pm|a|p)?$`)
	reDate     = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})(?:/(\d{2,4}))?$`)
)

var weekdays = map[string]time.Weekday{
	"sun": time.Sunday, "sunday": time.Sunday,
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "weds": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
}

// ParseSnooze turns a short spec into a wake time after now. Accepted forms:
//
//	30m, 2h, 1d, 1w            a duration from now
//	7am, 7:30pm, 19:00, 7      a clock time today, or tomorrow if already past
//	tomorrow [clock]           default DefaultWakeHour
//	mon [clock], friday 9am    the next such weekday, strictly after today
//	10/12 [clock], 10/12/27    a date, default DefaultWakeHour; next year if past
//
// Case-insensitive. The result is always strictly after now.
func ParseSnooze(spec string, now time.Time) (time.Time, error) {
	s := strings.ToLower(strings.TrimSpace(spec))
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return time.Time{}, errors.New("empty")
	}

	if m := reDuration.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		if n <= 0 {
			return time.Time{}, errors.New("duration must be positive")
		}
		var d time.Duration
		switch m[2][0] {
		case 'm':
			d = time.Duration(n) * time.Minute
		case 'h':
			d = time.Duration(n) * time.Hour
		case 'd':
			d = time.Duration(n) * 24 * time.Hour
		case 'w':
			d = time.Duration(n) * 7 * 24 * time.Hour
		}
		return now.Add(d), nil
	}

	// Split a leading day word (tomorrow / weekday / date) from an optional clock.
	dayWord, clockPart := s, ""
	if i := strings.IndexByte(s, ' '); i >= 0 {
		dayWord, clockPart = s[:i], s[i+1:]
	}

	// "7am" alone: clock today, else tomorrow.
	if clockPart == "" {
		if h, mi, ok := parseClock(dayWord); ok {
			t := time.Date(now.Year(), now.Month(), now.Day(), h, mi, 0, 0, now.Location())
			if !t.After(now) {
				t = t.AddDate(0, 0, 1)
			}
			return t, nil
		}
	}

	hour, minute := DefaultWakeHour, 0
	if clockPart != "" {
		h, mi, ok := parseClock(clockPart)
		if !ok {
			return time.Time{}, fmt.Errorf("can't read time %q", clockPart)
		}
		hour, minute = h, mi
	}

	var day time.Time
	switch {
	case dayWord == "tomorrow" || dayWord == "tmrw":
		day = now.AddDate(0, 0, 1)
	case dayWord == "today":
		day = now
	default:
		if wd, ok := weekdays[dayWord]; ok {
			delta := (int(wd) - int(now.Weekday()) + 7) % 7
			if delta == 0 {
				delta = 7
			}
			day = now.AddDate(0, 0, delta)
		} else if m := reDate.FindStringSubmatch(dayWord); m != nil {
			mo, _ := strconv.Atoi(m[1])
			d, _ := strconv.Atoi(m[2])
			if mo < 1 || mo > 12 || d < 1 || d > 31 {
				return time.Time{}, fmt.Errorf("bad date %q", dayWord)
			}
			year := now.Year()
			if m[3] != "" {
				y, _ := strconv.Atoi(m[3])
				if y < 100 {
					y += 2000
				}
				year = y
			}
			day = time.Date(year, time.Month(mo), d, 0, 0, 0, 0, now.Location())
			// A yearless date earlier than today means next year ("1/4" in
			// October). Today's date with a past clock is a mistake, not next
			// year; the final check below rejects it.
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			if m[3] == "" && day.Before(today) {
				day = day.AddDate(1, 0, 0)
			}
		} else {
			return time.Time{}, fmt.Errorf("can't read %q", spec)
		}
	}

	t := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, now.Location())
	if !t.After(now) {
		return time.Time{}, fmt.Errorf("%s is already past", t.Format("Mon 3:04pm"))
	}
	return t, nil
}

// parseClock reads "7", "7am", "7:30pm", "19:00", "7p". A bare hour with no
// am/pm is 24-hour.
func parseClock(s string) (hour, minute int, ok bool) {
	m := reClock.FindStringSubmatch(strings.ReplaceAll(s, " ", ""))
	if m == nil {
		return 0, 0, false
	}
	hour, _ = strconv.Atoi(m[1])
	if m[2] != "" {
		minute, _ = strconv.Atoi(m[2])
	}
	if minute > 59 {
		return 0, 0, false
	}
	switch m[3] {
	case "am", "a":
		if hour < 1 || hour > 12 {
			return 0, 0, false
		}
		if hour == 12 {
			hour = 0
		}
	case "pm", "p":
		if hour < 1 || hour > 12 {
			return 0, 0, false
		}
		if hour != 12 {
			hour += 12
		}
	default:
		if hour > 23 {
			return 0, 0, false
		}
	}
	return hour, minute, true
}

// SnoozeWakesToday reports whether the wake time falls on today's calendar
// day: the row shows a clock then, a calendar otherwise.
func SnoozeWakesToday(t, now time.Time) bool {
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// FormatSnoozeUntil renders a wake time for the session row: "3:15p" today,
// "Mon 7a" within the week, "10/12 7a" beyond.
func FormatSnoozeUntil(t, now time.Time) string {
	clock := strings.ToLower(t.Format("3:04pm"))
	clock = strings.TrimSuffix(strings.TrimSuffix(clock, "m"), "")
	// "3:00p" → "3p"; "3:15pm" → "3:15p"
	clock = strings.Replace(clock, ":00", "", 1)
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	days := int(t.Sub(today).Hours() / 24)
	switch {
	case days == 0 && t.After(today):
		return clock
	case days > 0 && days < 7:
		return t.Format("Mon ") + clock
	default:
		return t.Format("1/2 ") + clock
	}
}
