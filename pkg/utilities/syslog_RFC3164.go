package utilities

import "bytes"
import "strconv"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"

// Date format are sorted from most precise to less precise
var timeFormatListRFC3164 = []timeFormat{
	{timeRFC3339Nanov0,  len(timeRFC3339Nanov0)},
	{timeRFC3339Nanov1,  len(timeRFC3339Nanov1)},
	{timeRFC3339Microv0, len(timeRFC3339Microv0)},
	{timeRFC3339Microv1, len(timeRFC3339Microv1)},
	{timeRFC1123Z,       len(timeRFC1123Z)},
	{timeRFC3339Nanov2,  len(timeRFC3339Nanov2)},
	{timeRFC3339Milliv0, len(timeRFC3339Milliv0)},
	{timeRFC1123,        len(timeRFC1123)},
	{timeRFC3339Milliv1, len(timeRFC3339Milliv1)},
	{timeRFC3339Microv2, len(timeRFC3339Microv2)},
	{timeRFC3339v0,      len(timeRFC3339v0)},
	{timeRFC3339Milliv2, len(timeRFC3339Milliv2)},
	{timeRFC3339v1,      len(timeRFC3339v1)},
	{timeStampYear,      len(timeStampYear)},
	{timeANSIC,          len(timeANSIC)},
	{timeRFC3339v2,      len(timeRFC3339v2)},
	{timeStamp,          len(timeStamp)},
}

func decodeSyslogRFC3164(msg *backend.Message, data []byte, pos int) error {
	var err error
	var nextPos int
	var c byte

	// Try decoding time with all known format
	pos, err = decodeTimestamp(data, pos, timeFormatListRFC3164, msg)
	if err != nil {
		return err
	}

	// Decode hostname, it must be end by a space, search the space.
	nextPos = bytes.IndexByte(data[pos:], ' ')
	if nextPos == -1 {
		return errFmt09
	}
	msg.Hostname = string(data[pos : pos+nextPos])
	pos = pos + nextPos + 1

	// Decode TAG, the tag is ended by a space, a ':' or a '['
	for nextPos, c = range data[pos:] {
		if c == '[' || c == ' ' || c == ':' {
			break
		}
	}
	nextPos += pos
	if nextPos >= len(data) {
		return errFmt10
	}
	msg.Process = string(data[pos:nextPos])
	pos = nextPos

	// If the previous parser break on [, decode pid.
	if data[pos] == '[' {
		pos++
		nextPos = bytes.IndexByte(data[pos:], ']')
		if nextPos == -1 {
			return errFmt14
		}
		msg.Pid, err = strconv.Atoi(string(data[pos : pos+nextPos]))
		if err != nil {
			return errFmt12
		}
		pos = pos + nextPos + 1
	}

	// If we have a single ':' jump it
	if len(data[pos:]) > 0 && data[pos] == ':' {
		pos++
	}

	// We expect simple space, if encountered jump it, otherwise return error
	if len(data[pos:]) <= 0 || data[pos] != ' ' {
		return errFmt15
	}
	pos++

	// The remlaining data is the message
	msg.Data = string(data[pos:])

	return nil
}
