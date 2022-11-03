package irc

import "testing"

func TestCanParsePING(t *testing.T) {
	type testT struct {
		line string
		want *PingMessage
	}

	testCases := []testT{
		{
			line: "PING :tmi.twitch.tv",
			want: &PingMessage{
				Raw:     "PING :tmi.twitch.tv",
				Message: "tmi.twitch.tv",
			},
		},
		{
			line: ":tmi.twitch.tv PING :xD",
			want: &PingMessage{
				Raw:     ":tmi.twitch.tv PING :xD",
				Message: "xD",
			},
		},
	}

	for _, testCase := range testCases {
		got, err := ParseLine(testCase.line)
		if err != nil {
			t.Fatalf("ParseLine threw -- %v", err)
		}
		pingMsg := got.(*PingMessage)

		if pingMsg.GetType() != PING {
			t.Errorf("got %v, want %v", got, testCase.want)
		}

		assertEqual(t, pingMsg.Raw, testCase.want.Raw)
		assertEqual(t, pingMsg.Message, testCase.want.Message)
	}
}

func TestCanParsePONG(t *testing.T) {
	type testT struct {
		line string
		want *PongMessage
	}

	testCases := []testT{
		{
			line: "PONG :tmi.twitch.tv",
			want: &PongMessage{
				Raw:     "PONG :tmi.twitch.tv",
				Message: "",
			},
		},
		{
			line: ":tmi.twitch.tv PONG tmi.twitch.tv :HI-:D",
			want: &PongMessage{
				Raw:     ":tmi.twitch.tv PONG tmi.twitch.tv :HI-:D",
				Message: "HI-:D",
			},
		},
		{
			line: ":tmi.twitch.tv PONG tmi.twitch.tv :xD lol",
			want: &PongMessage{
				Raw:     ":tmi.twitch.tv PONG tmi.twitch.tv :xD lol",
				Message: "xD",
			},
		},
	}

	for _, testCase := range testCases {
		got, err := ParseLine(testCase.line)
		if err != nil {
			t.Fatalf("ParseLine threw -- %v", err)
		}
		pongMsg := got.(*PongMessage)

		if pongMsg.GetType() != PONG {
			t.Errorf("got %v, want %v", got, testCase.want)
		}

		assertEqual(t, pongMsg.Raw, testCase.want.Raw)
		assertEqual(t, pongMsg.Message, testCase.want.Message)
	}
}

func TestCanParsePRIVMSG(t *testing.T) {
	type testType struct {
		Line string
		Want *PrivmsgMessage
	}

	tests := []testType{
		{
			Line: "@badge-info=subscriber/17;badges=subscriber/12,game-developer/1;color=#FFFBFF;display-name=MarkZynk;emotes=;first-msg=0;flags=;historical=1;id=c98dc399-5e46-4727-b1fa-730b7c522c7c;mod=0;returning-chatter=0;rm-received-ts=1666921005122;room-id=22484632;subscriber=1;tmi-sent-ts=1666921002844;turbo=0;user-id=88492428;user-type= :markzynk!markzynk@markzynk.tmi.twitch.tv PRIVMSG #forsen :AlienPls",
			Want: &PrivmsgMessage{
				Raw:     "@badge-info=subscriber/17;badges=subscriber/12,game-developer/1;color=#FFFBFF;display-name=MarkZynk;emotes=;first-msg=0;flags=;historical=1;id=c98dc399-5e46-4727-b1fa-730b7c522c7c;mod=0;returning-chatter=0;rm-received-ts=1666921005122;room-id=22484632;subscriber=1;tmi-sent-ts=1666921002844;turbo=0;user-id=88492428;user-type= :markzynk!markzynk@markzynk.tmi.twitch.tv PRIVMSG #forsen :AlienPls",
				Channel: "forsen",
				User:    "markzynk",
				Message: "AlienPls",
			},
		},
		{
			Line: "@badge-info=subscriber/17;badges=subscriber/12,game-developer/1;color=#FFF2D8;display-name=MarkZynk;emotes=;first-msg=0;flags=;historical=1;id=f55f5e50-490e-4c5e-bb86-bfaef95a9916;mod=0;returning-chatter=0;rm-received-ts=1666921045502;room-id=22484632;subscriber=1;tmi-sent-ts=1666921045293;turbo=0;user-id=88492428;user-type= :markzynk!markzynk@markzynk.tmi.twitch.tv PRIVMSG #forsen :\u0001ACTION AlienPls\u0001",
			Want: &PrivmsgMessage{
				Raw:     "@badge-info=subscriber/17;badges=subscriber/12,game-developer/1;color=#FFF2D8;display-name=MarkZynk;emotes=;first-msg=0;flags=;historical=1;id=f55f5e50-490e-4c5e-bb86-bfaef95a9916;mod=0;returning-chatter=0;rm-received-ts=1666921045502;room-id=22484632;subscriber=1;tmi-sent-ts=1666921045293;turbo=0;user-id=88492428;user-type= :markzynk!markzynk@markzynk.tmi.twitch.tv PRIVMSG #forsen :\u0001ACTION AlienPls\u0001",
				Channel: "forsen",
				User:    "markzynk",
				Message: "AlienPls",
			},
		},
		{
			Line: "@badge-info=;badges=vip/1,overwatch-league-insider_2018B/1;color=#F01DD9;display-name=alyjiahT_T;emotes=;first-msg=0;flags=;historical=1;id=4e95adf6-153c-4152-87ae-7ca03fae1227;mod=0;returning-chatter=0;rm-received-ts=1666921189391;room-id=84180052;subscriber=0;tmi-sent-ts=1666921189173;turbo=0;user-id=145484970;user-type=;vip=1 :alyjiaht_t!alyjiaht_t@alyjiaht_t.tmi.twitch.tv PRIVMSG #brian6932 :hi brian",
			Want: &PrivmsgMessage{
				Raw:     "@badge-info=;badges=vip/1,overwatch-league-insider_2018B/1;color=#F01DD9;display-name=alyjiahT_T;emotes=;first-msg=0;flags=;historical=1;id=4e95adf6-153c-4152-87ae-7ca03fae1227;mod=0;returning-chatter=0;rm-received-ts=1666921189391;room-id=84180052;subscriber=0;tmi-sent-ts=1666921189173;turbo=0;user-id=145484970;user-type=;vip=1 :alyjiaht_t!alyjiaht_t@alyjiaht_t.tmi.twitch.tv PRIVMSG #brian6932 :hi brian",
				Channel: "brian6932",
				User:    "alyjiaht_t",
				Message: "hi brian",
			},
		},
	}

	for _, test := range tests {
		got, err := ParseLine(test.Line)
		if err != nil {
			t.Fatalf("ParseLine threw an error -- %v", err)
		}
		privMsg := got.(*PrivmsgMessage)

		if privMsg.GetType() != PRIVMSG {
			t.Errorf("got %v, want %v", got, test.Want)
		}

		assertEqual(t, privMsg.Raw, test.Want.Raw)
		assertEqual(t, privMsg.Channel, test.Want.Channel)
		assertEqual(t, privMsg.User, test.Want.User)
		assertEqual(t, privMsg.Message, test.Want.Message)
	}
}

func TestCanParseReconnect(t *testing.T) {
	line := ":tmi.twitch.tv RECONNECT"
	want := &ReconnectMessage{
		Raw: line,
	}

	got, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine threw -- %v", err)
	}
	reconnectMsg := got.(*ReconnectMessage)

	if reconnectMsg.GetType() != want.GetType() {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnknownCommand(t *testing.T) {
	testCases := []string{
		"FOOBAR",
		"FOOBAR :xD",
		"FOOBAR #forsen :xD",
		":forsen!forsen@forsen.tmi.twtich.tv COMMAND #forsen :xD",
	}

	for _, testCase := range testCases {
		msg, err := ParseLine(testCase)
		if err != nil {
			t.Fatalf("ParseLine threw an error -- %v", err)
		}
		rawMsg := msg.(*RawMessage)

		if rawMsg.Raw != testCase {
			t.Errorf("got %v, want %v", rawMsg.Raw, testCase)
		}

		if rawMsg.GetType() != UNSURE {
			t.Errorf("got %v, want %v", rawMsg.GetType(), UNSURE)
		}
	}
}

func TestParseMotd(t *testing.T) {
	type testType struct {
		Line string
		Want *EndOfMotdMessage
	}

	tests := []testType{
		{
			Line: ":tmi.twitch.tv 376 foobar :>",
			Want: &EndOfMotdMessage{
				Raw:     ":tmi.twitch.tv 376 foobar :>",
				User:    "foobar",
				Message: ">",
			},
		},
	}

	for _, test := range tests {
		got, err := ParseLine(test.Line)
		if err != nil {
			t.Fatalf("ParseLine threw an error -- %v", err)
		}
		motdMsg := got.(*EndOfMotdMessage)

		if motdMsg.GetType() != ENDOFMOTD {
			t.Errorf("got %v, want %v", got, test.Want)
		}

		assertEqual(t, motdMsg.Raw, test.Want.Raw)
		assertEqual(t, motdMsg.User, test.Want.User)
		assertEqual(t, motdMsg.Message, test.Want.Message)
	}
}

func TestParseNotice(t *testing.T) {
	type testType struct {
		Line string
		Want *NoticeMessage
	}

	tests := []testType{
		{
			Line: ":tmi.twitch.tv NOTICE #forsen :Login unsuccessful",
			Want: &NoticeMessage{
				Raw:     ":tmi.twitch.tv NOTICE #forsen :Login unsuccessful",
				Channel: "forsen",
				Message: "Login unsuccessful",
			},
		},
		{
			Line: ":tmi.twitch.tv NOTICE * :Login authentication failed",
			Want: &NoticeMessage{
				Raw:     ":tmi.twitch.tv NOTICE * :Login authentication failed",
				Channel: "*",
				Message: "Login authentication failed",
			},
		},
	}

	for _, test := range tests {
		got, err := ParseLine(test.Line)
		if err != nil {
			t.Fatalf("ParseLine threw an error -- %v", err)
		}
		noticeMsg := got.(*NoticeMessage)

		if noticeMsg.GetType() != NOTICE {
			t.Errorf("got %v, want %v", got, test.Want)
		}

		assertEqual(t, noticeMsg.Raw, test.Want.Raw)
		assertEqual(t, noticeMsg.Channel, test.Want.Channel)
		assertEqual(t, noticeMsg.Message, test.Want.Message)
	}
}

func TestParseJoin(t *testing.T) {
	type testType struct {
		Line string
		Want *JoinMessage
	}

	tests := []testType{
		{
			Line: ":justinfan123!justinfan123@justinfan123.tmi.twitch.tv JOIN #pajlada",
			Want: &JoinMessage{
				Raw:     ":justinfan123!justinfan123@justinfan123.tmi.twitch.tv JOIN #pajlada",
				User:    "justinfan123",
				Channel: "pajlada",
			},
		},
		{
			Line: ":forsen!forsen@forsen.tmi.twitch.tv JOIN #forsen",
			Want: &JoinMessage{
				Raw:     ":forsen!forsen@forsen.tmi.twitch.tv JOIN #forsen",
				User:    "forsen",
				Channel: "forsen",
			},
		},
	}

	for _, test := range tests {
		got, err := ParseLine(test.Line)
		if err != nil {
			t.Fatalf("ParseLine threw an error -- %v", err)
		}
		joinMsg := got.(*JoinMessage)

		if joinMsg.GetType() != JOIN {
			t.Errorf("got %v, want %v", got, test.Want)
		}

		assertEqual(t, joinMsg.Raw, test.Want.Raw)
		assertEqual(t, joinMsg.User, test.Want.User)
		assertEqual(t, joinMsg.Channel, test.Want.Channel)
	}
}

func TestParseTags(t *testing.T) {
	type testType struct {
		Line string
		Want Tags
	}

	tests := []testType{
		{
			Line: "@badge-info=;badges=vip/1,game-developer/1;color=#FF0000;display-name=melon095;emotes=;first-msg=0;flags=;historical=1;id=4a48a887-3a4b-4751-941c-4ddcad9cc22c;mod=0;returning-chatter=0;rm-received-ts=1666920226793;room-id=84180052;subscriber=0;tmi-sent-ts=1666920226595;turbo=0;user-id=146910710;user-type=;vip=1",
			Want: Tags{
				"badge-info":        "",
				"badges":            "vip/1,game-developer/1",
				"color":             "#FF0000",
				"display-name":      "melon095",
				"emotes":            "",
				"first-msg":         "0",
				"flags":             "",
				"historical":        "1",
				"id":                "4a48a887-3a4b-4751-941c-4ddcad9cc22c",
				"mod":               "0",
				"returning-chatter": "0",
				"rm-received-ts":    "1666920226793",
				"room-id":           "84180052",
				"subscriber":        "0",
				"tmi-sent-ts":       "1666920226595",
			},
		},
		{
			Line: "@badge-info=subscriber/5;badges=broadcaster/1,subscriber/0,game-developer/1;color=#0096CC;display-name=brian6932;emotes=;first-msg=0;flags=;historical=1;id=640f8e15-1825-4b96-bbb7-2783172b5ec0;mod=0;returning-chatter=0;rm-received-ts=1666900249841;room-id=84180052;subscriber=1;tmi-sent-ts=1666900249620;turbo=0;user-id=84180052;user-type=",
			Want: Tags{
				"badge-info":        "subscriber/5",
				"badges":            "broadcaster/1,subscriber/0,game-developer/1",
				"color":             "#0096CC",
				"display-name":      "brian6932",
				"emotes":            "",
				"first-msg":         "0",
				"flags":             "",
				"historical":        "1",
				"id":                "640f8e15-1825-4b96-bbb7-2783172b5ec0",
				"mod":               "0",
				"returning-chatter": "0",
				"rm-received-ts":    "1666900249841",
				"room-id":           "84180052",
				"subscriber":        "1",
				"tmi-sent-ts":       "1666900249620",
				"Hmm":               "",
				"b":                 "",
			},
		},
	}

	for _, test := range tests {
		got := parseTags(test.Line)

		for k, v := range test.Want {
			assertEqual(t, got[k], v)
		}
	}
}

func assertEqual(t *testing.T, lhs, rhs string) {
	if lhs != rhs {
		t.Errorf("got %v, want %v", lhs, rhs)
	}
}
