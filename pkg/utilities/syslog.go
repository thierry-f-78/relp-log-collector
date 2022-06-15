package utilities

import "bytes"
import "strconv"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"

func DecodeSyslog(data []byte) (*backend.Message, error) {
	var pos int
	var priority int
	var msg backend.Message
	var nextPos int
	var err error

	// Expect '<'
	pos = 0
	if len(data) <= 0 || data[pos] != '<' {
		return nil, errFmt01
	}
	pos++

	// Search end of priority, so the character '>'
	nextPos = bytes.IndexByte(data[pos:], '>')
	if nextPos == -1 {
		return nil, errFmt04
	}
	if nextPos == 0 {
		return nil, errFmt02
	}
	priority, err = strconv.Atoi(string(data[pos : pos+nextPos]))
	if err != nil {
		return nil, errFmt03
	}
	pos = pos + nextPos + 1

	msg.Facility = priority >> 3
	msg.Severity = priority & 0x07

	// Detecting RFC5424 and select parser
	if len(data[pos:]) > 2 && data[pos] >= '0' && data[pos] <= '9' && data[pos+1] == ' ' {
		err = decodeSyslogRFC5424(&msg, data, pos+2)
	} else {
		err = decodeSyslogRFC3164(&msg, data, pos)
	}
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
