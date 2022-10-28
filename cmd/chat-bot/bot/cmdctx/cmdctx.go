package cmdctx

import (
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
)

type Context interface {
	Channel() string
	Prompter() string
}

type context struct {
	ctx     ctx.Context
	user    *pb.IRCUser
	channel string
}

func NewContext(ctx ctx.Context, msg *pb.IRCPrivmsg) Context {
	return &context{
		ctx:     ctx,
		user:    msg.User,
		channel: msg.Channel,
	}
}
func (c *context) Channel() string {
	return c.channel
}

func (c *context) Prompter() string {
	return c.user.Username
}
