package utilities

import "time"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"

type timeFormat struct {
	format string
	length int
}

const timeRFC3339Nanov0   = "2006-01-02T15:04:05.999999999-07:00"
const timeRFC3339Nanov1   = "2006-01-02T15:04:05.999999999-0700"
const timeRFC3339Microv0  = "2006-01-02T15:04:05.999999-07:00"
const timeRFC3339Microv1  = "2006-01-02T15:04:05.999999-0700"
const timeRFC1123Z        = "Mon, 02 Jan 2006 15:04:05 -0700"
const timeRFC3339Nanov2   = "2006-01-02T15:04:05.999999999Z"
const timeRFC3339Milliv0  = "2006-01-02T15:04:05.999-07:00"
const timeRFC1123         = "Mon, 02 Jan 2006 15:04:05 MST"
const timeRFC3339Milliv1  = "2006-01-02T15:04:05.999-0700"
const timeRFC3339Microv2  = "2006-01-02T15:04:05.999999Z"
const timeRFC3339v0       = "2006-01-02T15:04:05-07:00"
const timeRFC3339Milliv2  = "2006-01-02T15:04:05.999Z"
const timeRFC3339v1       = "2006-01-02T15:04:05-0700"
const timeStampYear       = "Jan _2 15:04:05 MST 2006"
const timeANSIC           = "Mon Jan _2 15:04:05 2006"
const timeRFC3339v2       = "2006-01-02T15:04:05Z"
const timeStamp           = "Jan _2 15:04:05"

func decodeTimestamp(data []byte, pos int, timeFormatList []timeFormat, msg *backend.Message) (int, error) {
	var timeFmt timeFormat
	var err error
	var currentYear int

	// Try decoding time with all known format
	for _, timeFmt = range timeFormatList {

		// If message is too short, for this time format, try next
		if len(data[pos:]) < timeFmt.length+1 {
			continue
		}

		// Before executed complex parsing, ensure the white space separator
		if data[pos+timeFmt.length] != ' ' {
			continue
		}

		// Try decode date, on error, try next
		msg.Date, err = time.Parse(timeFmt.format, string(data[pos:pos+timeFmt.length]))
		if err != nil {
			continue
		}

		// If year is 0, use current year
		if msg.Date.Year() == 0 {
			currentYear = time.Now().Year()
			msg.Date = time.Date(currentYear, msg.Date.Month(), msg.Date.Day(),
				msg.Date.Hour(), msg.Date.Minute(), msg.Date.Second(),
				msg.Date.Nanosecond(), msg.Date.Location())
		}

		// eat the date and continue parsing
		return pos + timeFmt.length + 1, nil
	}
	return -1, errFmt08
}
