package parser

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
				Raw: "PING :tmi.twitch.tv",
				Message: "tmi.twitch.tv",
			},
		},
		{
			line: ":tmi.twitch.tv PING :xD",
			want: &PingMessage{
				Raw: ":tmi.twitch.tv PING :xD",
				Message: "xD",
			},
		},
	}


	for _, testCase := range testCases {
		got, err := ParseLine(testCase.line)
		if err != nil { t.Fatalf("ParseLine threw -- %v", err) }
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
				Raw: "PONG :tmi.twitch.tv",
				Message: "",
			},
		},
		{
			line: ":tmi.twitch.tv PONG tmi.twitch.tv :HI-:D",
			want: &PongMessage{
				Raw: ":tmi.twitch.tv PONG tmi.twitch.tv :HI-:D",
				Message: "HI-:D",
			},
		},
		{
			line: ":tmi.twitch.tv PONG tmi.twitch.tv :xD lol",
			want: &PongMessage{
				Raw: ":tmi.twitch.tv PONG tmi.twitch.tv :xD lol",
				Message: "xD",
			},
		},
	}

	for _, testCase := range testCases {
		got, err := ParseLine(testCase.line)
		if err != nil { t.Fatalf("ParseLine threw -- %v", err) }
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
			Line: ":foobar!foobar@foobar.tmi.twitch.tv PRIVMSG #forsen :forsenInsane",
			Want: &PrivmsgMessage{
				Raw: ":foobar!foobar@foobar.tmi.twitch.tv PRIVMSG #forsen :forsenInsane",
				Channel: "forsen",
				User: "foobar",
				Message: "forsenInsane",
			},
		},
		{
			Line: ":kkonaaaaaaaaaaa!kkonaaaaaaaaaaa@kkonaaaaaaaaaaa.tmi.twitch.tv PRIVMSG #forsen :REALLY LONG MESSAGE AAAAAAAA",
			Want: &PrivmsgMessage{
				Raw: ":kkonaaaaaaaaaaa!kkonaaaaaaaaaaa@kkonaaaaaaaaaaa.tmi.twitch.tv PRIVMSG #forsen :REALLY LONG MESSAGE AAAAAAAA",
				Channel: "forsen",
				User: "kkonaaaaaaaaaaa",
				Message: "REALLY LONG MESSAGE AAAAAAAA",
			},
		},
		{
			Line: ":melon095!melon095@melon095.tmi.twitch.tv PRIVMSG #brian6932 :Twitch parser",
			Want: &PrivmsgMessage{
				Raw: ":melon095!melon095@melon095.tmi.twitch.tv PRIVMSG #brian6932 :Twitch parser",
				Channel: "brian6932",
				User: "melon095",
				Message: "Twitch parser",
			},
		},
	}

	for _, test := range tests {
		got, err := ParseLine(test.Line)
		if err != nil { t.Fatalf("ParseLine threw an error -- %v", err) }
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
	if err != nil { t.Fatalf("ParseLine threw -- %v", err) }
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
		if err != nil { t.Fatalf("ParseLine threw an error -- %v", err) }
		rawMsg := msg.(*RawMessage)

		if rawMsg.Raw != testCase {
			t.Errorf("got %v, want %v", rawMsg.Raw, testCase)
		}
		
		if rawMsg.GetType() != UNSURE {
			t.Errorf("got %v, want %v", rawMsg.GetType(), UNSURE)
		}
	}
}

func assertEqual(t *testing.T, lhs, rhs string) {
	if lhs != rhs {
		t.Errorf("got %v, want %v", lhs, rhs)
	}
}