package bot

import (
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/cmdctx"
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/execlevel"
)

var commands = make(map[string]Command)

type Command interface {
	Name() string
	ExecutionLevel() execlevel.ExecutionLevel
	Execute(ctx cmdctx.Context, bot Bot, args []string) error
}

func CanExecute(cmd string, level execlevel.ExecutionLevel) (bool, Command) {
	if c, ok := commands[cmd]; ok {
		return c.ExecutionLevel() <= level, c
	}

	return false, nil
}

func Get(cmd string) *Command {
	if c, ok := commands[cmd]; ok {
		return &c
	}

	return nil
}
