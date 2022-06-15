package utilities

import "testing"

var logs []string = []string{
	`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8`,
	`<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`,
	`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource= "Application" eventID="1011"] BOMAn application event log entry...`,
	`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource= "Application" eventID="1011"][examplePriority@32473 class="high"]`,
	`<34>1 2024-06-15T12:00:00.000Z host1 application1 1234 - - This is an informational message.`,
	`<165>Aug 24 05:34:00 CST 1987 mymachine myproc[10]: %% It's time to make the do-nuts.  %%  Ingredients: Mix=OK, Jelly=OK # Devices: Mixer=OK, Jelly_Injector=OK, Frier=OK # Transport: Conveyer1=OK, Conveyer2=OK # %%`,
	`<34>Jun 15 12:00:00 host1 application1: This is a very long message that is meant to test the parsing capabilities of the syslog parser. It includes a lot of text to ensure that the buffer sizes and handling mechanisms are thoroughly tested. This should cover multiple lines and include various characters such as !@#$%^&*().`,
	`<34>Jun 15 12:01:00 host2 application2: Message with special characters: !@#$%^&*()_+-=~` + "`" + `[]{}|;:'",.<>/?\\`,
	`<34>Jun 15 12:03:00 application4: This message is missing the hostname part.`,
	`<34>Jun 15 12:04:00 host5: This message is missing the application name part.`,
	`<0>2022-01-01T00:00:00Z hostname app `,
	`<0>2022-01-01T00:00:00Z hostname app: `,
	`<0>2022-01-01T00:00:00Z hostname app[10]: `,
}

var logErrors []string = []string{
	`<34>Junn 15 12:02:00 host3 application3: This message has an invalid date format.`,
	``,
	`w`,
	`<`,
	`<>`,
	`<0>`,
	`<toto>`,
	`<0>434234`,
	`<0>1 -`,
	`<0>1 - `,
	`<0>1 2022-01-01T00:00:00Z`,
	`<0>1 2022-01-01T00:00:00Z `,
	`<0>1 2022-01-01T00:00:00Z &`,
	`<0>1 2022-01-01T00:00:00Z hostname`,
	`<0>1 2022-01-01T00:00:00Z hostname `,
	`<0>1 2022-01-01T00:00:00Z hostname app`,
	`<0>1 2022-01-01T00:00:00Z hostname app pid`,
	`<0>1 2022-01-01T00:00:00Z hostname app 0`,
	`<0>1 2022-01-01T00:00:00Z hostname app 1234`,
	`<0>1 2022-01-01T00:00:00Z hostname app pid `,
	`<0>1 2022-01-01T00:00:00Z hostname app 0 `,
	`<0>1 2022-01-01T00:00:00Z hostname app 1234 `,
	`<0>2022-01-01T00:00:00Z`,
	`<0>2022-01-01T00:00:00Z `,
	`<0>2022-01-01T00:00:00Z hostname`,
	`<0>2022-01-01T00:00:00Z hostname `,
	`<0>2022-01-01T00:00:00Z hostname app`,
	`<0>2022-01-01T00:00:00Z hostname app[]: `,
	`<0>2022-01-01T00:00:00Z hostname app[a: `,
	`<0>2022-01-01T00:00:00Z hostname app[1: `,
	`<0>2022-01-01T00:00:00Z hostname app[: `,
	`<0>2022-01-01T00:00:00Z hostname app[&]: `,
	// These log are valid, but the format is too old
	`<0>1990 Oct 22 10:52:01 TZ-6 scapegoat.dmz.example.org 10.1.2.3 sched[0]: That's All Folks!`,
}

func TestSyslogParser(t *testing.T) {
	var l string
	var err error

	for _, l = range logs {
		_, err = DecodeSyslog([]byte(l))
		if err != nil {
			t.Error(err)
		}
	}

	for _, l = range logErrors {
		_, err = DecodeSyslog([]byte(l))
		if err == nil {
			t.Errorf("Expect error with %q", l)
		}
	}
}
