package utilities

import "fmt"

var errFmt01 = fmt.Errorf("'<' expected at start of message")
var errFmt02 = fmt.Errorf("priority cannot be empty")
var errFmt03 = fmt.Errorf("can't decode priority")
var errFmt04 = fmt.Errorf("priority not terminated")
var errFmt05 = fmt.Errorf("expect protocol version")
var errFmt06 = fmt.Errorf("wrong protocol version")
var errFmt07 = fmt.Errorf("cannot process log without date")
var errFmt08 = fmt.Errorf("cannot decode timestamp")
var errFmt09 = fmt.Errorf("hostname not found")
var errFmt10 = fmt.Errorf("process name not found")
var errFmt11 = fmt.Errorf("pid not found")
var errFmt12 = fmt.Errorf("can't decode pid")
var errFmt13 = fmt.Errorf("msgid not found")
var errFmt14 = fmt.Errorf("close ] not found")
var errFmt15 = fmt.Errorf("tag not followed by seperation marker")
