package irc

import (
	"strings"
)

type ircMessage struct {
	Raw     string
	Source  ircMessageSource
	Command string
	Params  []string
	Tags    Tags
}

type ircMessageSource struct {
	Nick string
	User string
	Host string
}

func ParseLine(line string) (Message, error) {
	m := ircMessage{
		Raw:    line,
		Params: make([]string, 0),
	}

	split := strings.Split(line, " ")
	idx := 0

	if strings.HasPrefix(split[0], "@") {
		m.Tags = parseTags(split[idx])
		idx++
	}

	if strings.HasPrefix(split[idx], ":") {
		m.Source = parseSource(split[idx])
		idx++
	}

	m.Command = split[idx]
	idx++

	for idx < len(split) {
		if strings.HasPrefix(split[idx], ":") {
			m.Params = append(m.Params, strings.Join(split[idx:], " ")[1:])
			break
		}

		m.Params = append(m.Params, split[idx])
		idx++
	}

	switch m.Command {
	case "376":
		{
			return parseMotd(m), nil
		}
	case "NOTICE":
		{
			return parseNotice(m), nil
		}
	case "PING":
		{
			return parsePing(&m), nil
		}
	case "PONG":
		{
			return parsePong(&m), nil
		}
	case "PRIVMSG":
		{
			return parsePrivmsg(&m), nil
		}
	case "RECONNECT":
		{
			return parseReconnect(&m), nil
		}
	case "JOIN":
		{
			return parseJoin(m), nil
		}
	case "PART":
		{
			return parsePart(m), nil
		}
	default:
		{
			return &RawMessage{
				Raw:  m.Raw,
				Type: UNSURE,
			}, nil
		}
	}
}

func sanitizeChannel(channel string) string {
	return strings.Replace(channel, "#", "", 1)
}

func parseSource(source string) ircMessageSource {
	s := ircMessageSource{}

	split := strings.Split(source, "!")

	if len(split) > 0 {
		s.Nick = split[0][1:]
	}

	if len(split) > 1 {
		split = strings.Split(split[1], "@")
		s.User = split[0]
		s.Host = split[1]
	}

	return s
}

func parsePing(m *ircMessage) *PingMessage {
	p := &PingMessage{
		Raw: m.Raw,
	}

	if len(m.Params) == 1 {
		p.Message = strings.Split(m.Params[0], " ")[0]
	}

	return p
}

func parsePong(m *ircMessage) *PongMessage {
	p := &PongMessage{
		Raw: m.Raw,
	}

	if len(m.Params) == 2 {
		p.Message = strings.Split(m.Params[1], " ")[0]
	}

	return p
}

func parsePrivmsg(m *ircMessage) *PrivmsgMessage {
	msg := m.Params[1]

	/* /me commands */
	if strings.HasPrefix(msg, "\x01ACTION") && strings.HasSuffix(msg, "\x01") {
		msg = strings.TrimPrefix(msg, "\x01ACTION ")
		msg = strings.TrimSuffix(msg, "\x01")
	}

	return &PrivmsgMessage{
		Raw:     m.Raw,
		Channel: sanitizeChannel(m.Params[0]),
		User:    m.Source.Nick,
		Message: msg,
		Tags:    m.Tags,
	}
}

func parseReconnect(m *ircMessage) *ReconnectMessage {
	return &ReconnectMessage{
		Raw: m.Raw,
	}
}

func parseMotd(m ircMessage) *EndOfMotdMessage {
	return &EndOfMotdMessage{
		Raw:     m.Raw,
		User:    sanitizeChannel(m.Params[0]),
		Message: m.Params[1],
	}
}

func parseNotice(m ircMessage) *NoticeMessage {
	return &NoticeMessage{
		Raw:     m.Raw,
		Channel: sanitizeChannel(m.Params[0]),
		Message: m.Params[1],
		Tags:    m.Tags,
	}
}

func parseJoin(m ircMessage) *JoinMessage {
	return &JoinMessage{
		Raw:     m.Raw,
		Channel: sanitizeChannel(m.Params[0]),
		User:    m.Source.Nick,
	}
}

func parsePart(m ircMessage) *PartMessage {
	return &PartMessage{
		Raw:     m.Raw,
		Channel: sanitizeChannel(m.Params[0]),
		User:    m.Source.Nick,
	}
}

func parseTags(tags string) Tags {
	tags = strings.TrimPrefix(tags, "@")

	split := strings.Split(tags, ";")

	t := make(Tags, len(split))

	for _, tag := range split {
		split := strings.Split(tag, "=")
		t[split[0]] = split[1]
	}

	return t
}
