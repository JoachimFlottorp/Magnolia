package bot

import (
	"fmt"
	"strings"

	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/cmdctx"
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/execlevel"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
)

type clearCommand struct {
	Ctx ctx.Context
}

func newClearCommand(gCtx ctx.Context) Command {
	return clearCommand{
		Ctx: gCtx,
	}
}

func (c clearCommand) Name() string {
	return "clear"
}

func (c clearCommand) ExecutionLevel() execlevel.ExecutionLevel {
	return execlevel.ExecutionLevelAdmin
}

func (c clearCommand) Execute(ctx cmdctx.Context, b Bot, args []string) error {
	channel := args[0]
	if channel == "" {
		b.Say(ctx.Channel(), "Provide a channel FeelsDankMan")
		return nil
	}
	
	channel = strings.Replace(channel, "$this", ctx.Channel(), -1)
	channel = strings.ToLower(channel)
	
	key := fmt.Sprintf("twitch:%s:chat-data", channel)

	err := c.Ctx.Inst().Redis.Del(c.Ctx, key)
	if err != nil {
		b.Say(ctx.Channel(), "Failed to clear chat data FeelsDankMan")
		return err
	}
	b.Say(ctx.Channel(), "ok FeelsDankMan")
	
	return nil
}
