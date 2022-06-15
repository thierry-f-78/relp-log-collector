package utilities

import "testing"

var testList []string = []string{
	"",
	"\n",
	"\a\b\t\n\v\f\r\\\x88",
	"<30>2024-06-16T14:35:40.010602+00:00 rc systemd[1358807]: run-docker-runtime\\\\x2drunc-moby-417f03b4fbaba9e04da747c53665b8cfd41351ad470ea6d244aed88ae2b32d6e-runc.3Rudyi.mount: Succeeded.",
}

var errorList []string = []string{
	"string\\",
	"string\\x",
	"string\\x1",
	"string\\xzz",
	"string\\w",
}

func TestEscape(t *testing.T) {
	var l string
	var out []byte
	var err error
	var dec string

	for _, l = range testList {
		out = EscapeNonASCIIPrintable([]byte(l))
		dec, err = UnescapeNonASCIIPrintable(string(out))
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		if dec != l {
			t.Errorf("Expected %q, got %q", l, dec)
		}
	}

	for _, l = range errorList {
		_, err = UnescapeNonASCIIPrintable(l)
		if err == nil {
			t.Errorf("Expect error decoding %q", l)
		}
	}

}
