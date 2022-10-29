/*
	This does not directly connect to a twitch channel

	However, it utilizes the twitch-reader which exposes data on a pub|sub channel

	We require one IRC connection to type in chat.
	This IRC connection does not need to be connected to a channel.
*/

package bot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/cmdctx"
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/execlevel"
	"github.com/JoachimFlottorp/magnolia/cmd/twitch-reader/irc"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNoUsername = errors.New("no username provided")
	ErrNoPassword = errors.New("no password provided")
)

type Credentials struct {
	Username string
	Password string
}

type bot struct {
	ctx         ctx.Context
	credentials Credentials
	irc         *irc.IrcConnection
	prefix      string
	admins      []string
}

func NewBot(ctx ctx.Context, creds Credentials) (Bot, error) {
	if creds.Username == "" {
		return nil, ErrNoUsername
	}

	if creds.Password == "" {
		return nil, ErrNoPassword
	}

	i := irc.NewClient(creds.Username, creds.Password)

	b := &bot{
		ctx:         ctx,
		credentials: creds,
		irc:         i,
		prefix:      ctx.Config().Twitch.Bot.Prefix,
		admins:      ctx.Config().Twitch.Bot.Admins,
	}

	return b, nil
}

func (b *bot) Run() error {
	if err := b.irc.Connect(); err != nil {
		return err
	}

	b.setupCommands()

	go func() {
		newMsg, err := b.ctx.Inst().Redis.Subscribe(b.ctx, "twitch:messages")
		if err != nil {
			zap.S().Errorw("Failed to subscribe to twitch:messages", "error", err)
			return
		}

		for {
			select {
			case <-b.ctx.Done():
				{
					return
				}
			case rawMsg := <-newMsg:
				{
					msg := &pb.IRCPrivmsg{}
					if err := proto.Unmarshal([]byte(rawMsg), msg); err != nil {
						zap.S().Errorw("Failed to unmarshal message", "error", err)
						continue
					}

					commandName, args := cleanInput(b.prefix, msg.Message)
					if commandName == "" {
						continue
					}

					execLevel := getExec(b.admins, msg.User.UserId)

					ok, command := CanExecute(commandName, execLevel)
					if !ok {
						continue
					}

					ctx := cmdctx.NewContext(b.ctx, msg)

					if err := command.Execute(ctx, b, args); err != nil {
						zap.S().Errorw("Failed to execute command", "error", err)
						b.Say(ctx.Channel(), "Something bad happened FeelsDankMan")
						continue
					}

					zap.S().Infow("Executed command", "command", commandName, "args", args)
					continue
				}
			}
		}
	}()

	return nil
}

func (b *bot) Say(channel, message string, args ...interface{}) {
	b.irc.Send(fmt.Sprintf("PRIVMSG #%s :%s", channel, fmt.Sprintf(message, args...)))
}

func (b *bot) setupCommands() {
	healthCommand := newJoinCommand(b.ctx)
	commands[healthCommand.Name()] = healthCommand
	
	clearCommand := newClearCommand(b.ctx)
	commands[clearCommand.Name()] = clearCommand
}


func cleanInput(prefix, input string) (command string, args []string) {
	command = strings.Split(input, " ")[0]
	args = strings.Split(input, " ")[1:]

	command = strings.TrimPrefix(command, prefix)

	return
}

func getExec(admins []string, userId string) execlevel.ExecutionLevel {
	for _, admin := range admins {
		if admin == userId {
			return execlevel.ExecutionLevelAdmin
		}
	}

	return execlevel.ExecutionLevelEveryone
}
