package pghelp

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type TimeSlice []time.Time

func errorf(s string, args ...interface{}) {
	panic(fmt.Errorf("pq: %s", fmt.Sprintf(s, args...)))
}
func expect(str, char string, pos int) {
	if c := str[pos : pos+1]; c != char {
		errorf("expected '%v' at position %v; got '%v'", char, pos, c)
	}
}

func mustAtoi(str string) int {
	result, err := strconv.Atoi(str)
	if err != nil {
		errorf("expected number; got '%v'", str)
	}
	return result
}

func parseTs(currentLocation *time.Location, str string) (result time.Time) {
	monSep := strings.IndexRune(str, '-')
	year := mustAtoi(str[:monSep])
	daySep := monSep + 3
	month := mustAtoi(str[monSep+1 : daySep])
	expect(str, "-", daySep)
	timeSep := daySep + 3
	day := mustAtoi(str[daySep+1 : timeSep])

	var hour, minute, second int
	if len(str) > monSep+len("01-01")+1 {
		expect(str, " ", timeSep)
		minSep := timeSep + 3
		expect(str, ":", minSep)
		hour = mustAtoi(str[timeSep+1 : minSep])
		secSep := minSep + 3
		expect(str, ":", secSep)
		minute = mustAtoi(str[minSep+1 : secSep])
		secEnd := secSep + 3
		second = mustAtoi(str[secSep+1 : secEnd])
	}
	remainderIdx := monSep + len("01-01 00:00:00") + 1
	// Three optional (but ordered) sections follow: the
	// fractional seconds, the time zone offset, and the BC
	// designation. We set them up here and adjust the other
	// offsets if the preceding sections exist.

	nanoSec := 0
	tzOff := 0
	bcSign := 1

	if remainderIdx < len(str) && str[remainderIdx:remainderIdx+1] == "." {
		fracStart := remainderIdx + 1
		fracOff := strings.IndexAny(str[fracStart:], "-+ ")
		if fracOff < 0 {
			fracOff = len(str) - fracStart
		}
		fracSec := mustAtoi(str[fracStart : fracStart+fracOff])
		nanoSec = fracSec * (1000000000 / int(math.Pow(10, float64(fracOff))))

		remainderIdx += fracOff + 1
	}
	if tzStart := remainderIdx; tzStart < len(str) && (str[tzStart:tzStart+1] == "-" || str[tzStart:tzStart+1] == "+") {
		// time zone separator is always '-' or '+' (UTC is +00)
		var tzSign int
		if c := str[tzStart : tzStart+1]; c == "-" {
			tzSign = -1
		} else if c == "+" {
			tzSign = +1
		} else {
			errorf("expected '-' or '+' at position %v; got %v", tzStart, c)
		}
		tzHours := mustAtoi(str[tzStart+1 : tzStart+3])
		remainderIdx += 3
		var tzMin, tzSec int
		if tzStart+3 < len(str) && str[tzStart+3:tzStart+4] == ":" {
			tzMin = mustAtoi(str[tzStart+4 : tzStart+6])
			remainderIdx += 3
		}
		if tzStart+6 < len(str) && str[tzStart+6:tzStart+7] == ":" {
			tzSec = mustAtoi(str[tzStart+7 : tzStart+9])
			remainderIdx += 3
		}
		tzOff = (tzSign * tzHours * (60 * 60)) + (tzMin * 60) + tzSec
	}
	if remainderIdx < len(str) && str[remainderIdx:remainderIdx+3] == " BC" {
		bcSign = -1
		remainderIdx += 3
	}
	if remainderIdx < len(str) {
		errorf("expected end of input, got %v", str[remainderIdx:])
	}
	t := time.Date(bcSign*year, time.Month(month), day,
		hour, minute, second, nanoSec,
		time.FixedZone("", tzOff))

	if currentLocation != nil {
		// Set the location of the returned Time based on the session's
		// TimeZone value, but only if the local time zone database agrees with
		// the remote database on the offset.
		lt := t.In(currentLocation)
		_, newOff := lt.Zone()
		if newOff == tzOff {
			t = lt
		}
	}

	return t
}
func (f *TimeSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		tmp := parsePGArray(string(t))
		rev := make([]time.Time, len(tmp))
		for i, tv := range tmp {
			rev[i] = parseTs(nil, tv)
		}
		*f = rev

		return nil
	case string:
		tmp := parsePGArray(t)
		rev := make([]time.Time, len(tmp))
		for i, tv := range tmp {
			rev[i] = parseTs(nil, tv)
		}
		*f = rev

		return nil
	case TimeSlice:
		*f = t
		return nil
	default:
		return ERROR_Convert(value, f)
	}
}
func (f TimeSlice) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	rev := make([]string, len(f))
	for i, v := range f {
		rev[i] = v.Format(time.RFC3339Nano)
	}
	return "{" + strings.Join(rev, ",") + "}", nil
}

type NullTimeSlice struct {
	Slice TimeSlice
	Valid bool
}

func (f *NullTimeSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullTimeSlice:
		*f = t
		return nil
	case nil:
		f.Valid = false
		f.Slice = nil
		return nil
	default:
		f.Valid = true
		return (&f.Slice).Scan(value)
	}
}
func (f NullTimeSlice) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Slice.Value()
	}
}
func (f NullTimeSlice) IsNull() bool {
	return !f.Valid
}
