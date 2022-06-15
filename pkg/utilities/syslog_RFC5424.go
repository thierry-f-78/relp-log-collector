package utilities

import "bytes"
import "strconv"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"

// SYSLOG-MSG      = HEADER SP STRUCTURED-DATA [SP MSG]
//
// HEADER          = PRI VERSION SP TIMESTAMP SP HOSTNAME
//                   SP APP-NAME SP PROCID SP MSGID
// PRI             = "<" PRIVAL ">"
// PRIVAL          = 1*3DIGIT ; range 0 .. 191
// VERSION         = NONZERO-DIGIT 0*2DIGIT
// HOSTNAME        = NILVALUE / 1*255PRINTUSASCII
//
// APP-NAME        = NILVALUE / 1*48PRINTUSASCII
// PROCID          = NILVALUE / 1*128PRINTUSASCII
// MSGID           = NILVALUE / 1*32PRINTUSASCII
//
// TIMESTAMP       = NILVALUE / FULL-DATE "T" FULL-TIME
// FULL-DATE       = DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
// DATE-FULLYEAR   = 4DIGIT
// DATE-MONTH      = 2DIGIT  ; 01-12
// DATE-MDAY       = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on
//                           ; month/year
// FULL-TIME       = PARTIAL-TIME TIME-OFFSET
// PARTIAL-TIME    = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND
//                   [TIME-SECFRAC]
// TIME-HOUR       = 2DIGIT  ; 00-23
// TIME-MINUTE     = 2DIGIT  ; 00-59
// TIME-SECOND     = 2DIGIT  ; 00-59
// TIME-SECFRAC    = "." 1*6DIGIT
// TIME-OFFSET     = "Z" / TIME-NUMOFFSET
// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
//
// STRUCTURED-DATA = NILVALUE / 1*SD-ELEMENT
// SD-ELEMENT      = "[" SD-ID *(SP SD-PARAM) "]"
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// SD-ID           = SD-NAME
// PARAM-NAME      = SD-NAME
// PARAM-VALUE     = UTF-8-STRING ; characters '"', '\' and
//                                ; ']' MUST be escaped.
// SD-NAME         = 1*32PRINTUSASCII
//                   ; except '=', SP, ']', %d34 (")
//
// MSG             = MSG-ANY / MSG-UTF8
// MSG-ANY         = *OCTET ; not starting with BOM
// MSG-UTF8        = BOM UTF-8-STRING
// BOM             = %xEF.BB.BF
// UTF-8-STRING    = *OCTET ; UTF-8 string as specified
//                   ; in RFC 3629
//
// OCTET           = %d00-255
// SP              = %d32
// PRINTUSASCII    = %d33-126
// NONZERO-DIGIT   = %d49-57
// DIGIT           = %d48 / NONZERO-DIGIT
// NILVALUE        = "-"

var timeFormatListRFC5424 = []timeFormat{
	{timeRFC3339Nanov0,  len(timeRFC3339Nanov0)},
	{timeRFC3339Nanov1,  len(timeRFC3339Nanov1)},
	{timeRFC3339Microv0, len(timeRFC3339Microv0)},
	{timeRFC3339Microv1, len(timeRFC3339Microv1)},
	{timeRFC3339Nanov2,  len(timeRFC3339Nanov2)},
	{timeRFC3339Milliv0, len(timeRFC3339Milliv0)},
	{timeRFC3339Milliv1, len(timeRFC3339Milliv1)},
	{timeRFC3339Microv2, len(timeRFC3339Microv2)},
	{timeRFC3339v0,      len(timeRFC3339v0)},
	{timeRFC3339Milliv2, len(timeRFC3339Milliv2)},
	{timeRFC3339v1,      len(timeRFC3339v1)},
	{timeRFC3339v2,      len(timeRFC3339v2)},
}

func decodeSyslogRFC5424(msg *backend.Message, data []byte, pos int) error {
	var nextPos int
	var err error

	// Expect timestamp as NILVALUE or RFC3339 date. If we encounter nil value,
	// The log message cannot be processed.
	if len(data[pos:]) > 0 && data[pos] == '-' {
		return errFmt07
	}
	pos, err = decodeTimestamp(data, pos, timeFormatListRFC5424, msg)
	if err != nil {
		return err
	}

	// Expect hostname separed by a space, search the space
	nextPos = bytes.IndexByte(data[pos:], ' ')
	if nextPos == -1 {
		return errFmt09
	}
	msg.Hostname = string(data[pos : pos+nextPos])
	pos = pos + nextPos + 1

	// expect application name separated by a space
	nextPos = bytes.IndexByte(data[pos:], ' ')
	if nextPos == -1 {
		return errFmt10
	}
	msg.Process = string(data[pos : pos+nextPos])
	pos = pos + nextPos + 1

	// Expect pid of application which sent log separated by a space
	nextPos = bytes.IndexByte(data[pos:], ' ')
	if nextPos == -1 {
		return errFmt11
	}
	if data[pos] != '-' {
		_, err = strconv.Atoi(string(data[pos : pos+nextPos]))
		if err != nil {
			return errFmt12
		}
	}
	pos = pos + nextPos + 1

	// Expect MSGID, consider as token separed by whitesapce
	nextPos = bytes.IndexByte(data[pos:], ' ')
	if nextPos == -1 {
		return errFmt13
	}
	pos = pos + nextPos + 1

	// Structure data could be NILVALUE. If it is, jump it.
	if len(data[pos:]) > 1 && data[pos] == '-' && data[pos+1] == ' ' {
		pos += 2
	}

	// Rest of data is message.
	msg.Data = string(data[pos:])

	return nil
}
