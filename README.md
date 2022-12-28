# Magnolia

Magnolia is an application with a variety of features. Such as the ability to generate markov chains from a extensive list of channels, and in the future will be able to measure emote usages from third party services such as [FFZ](https://www.frankerfacez.com/), [BTTV](https://betterttv.com/) and [7TV](https://7tv.app/)

[This](https://magnolia.melon095.live/) is a running example of Magnolia.

And has integrated itself with [Botbear](https://github.com/hotbear1110/botbear) and [Melonbot](https://github.com/JoachimFlottorp/Melonbot)

## Setting up

Setting up magnolia is simple, first install and setup [Redis](https://redis.io/), [MongoDB](https://www.mongodb.com) and [RabbitMQ](https://www.rabbitmq.com/) afterwards copy _config.example.toml_ to _config.toml_ and fill in the required data.

A valid twitch oauth token is required for the Chat-Bot as it needs the ability to type in chat, however the program which reads chat for markov data utilizes an anonymous connections, you can tell it to use a authed account by changing the variables [here](https://github.com/JoachimFlottorp/Magnolia/blob/main/pkg/irc/irc.go#L14)

Using an account instead of an anonymous connection can give it the ability to join channels faster, however it requires to be [Verified](https://dev.twitch.tv/docs/irc)

You can get a _Twitch_ oauth password using [this website](https://twitchtokengenerator.com/) by clicking on `Bot Chat Token`, authorize and copying _Access Token_.

Once it's setup you can run the programs manually by building with _go build_, However [docker compose](https://docs.docker.com/compose/) is recommended to automatically keep control of each program.

To start magnolia via docker compose run

```bash
make docker # Create a shared docker-file before building each program.
docker compose build # Build every docker container.
docker compose up # Start up every docker container, use -d to detach.
```

Magnolia will not join any channels before a request to the server has been made. So first open up a browser and make a GET request to _/api/markov?channel=yourchannel_ and swap `yourchannel`.

Afterwards it will join the specified channel and log it forever, or until you invoke the _part_ command on the chat-bot.

#### Chat-bot

The chat-bot is a pretty simple interface for manually joining or parting a channel,

| Command | Description                                           |
| ------- | ----------------------------------------------------- |
| join    | Joins the specified channel and starts logging it     |
| leave   | Parts the specified channel, clearing all of its data |

It will only listen to messages which start with the _Prefix_ and users that matches the uid given in the configuration file. You can find your accounts user-id [here](https://www.twitchdatabase.com/channels/melon095)

### Privacy And Policy

Magnolia collects public chat messages for generating markov chains.

However magnolia does NOT log usernames and does NOT have the ability to link messages to a specific users.

Chat messages are deleted once a certain threshold is met.
