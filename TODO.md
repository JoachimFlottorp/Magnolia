### TODO

Generate markov chains from twitch chat.
API Endpoint: /api/v1/markov
Supported parameters:

- channel: The channel to generate the markov chain from
- seed: Generate a markov chain from the given input

We expect that a lot of channels will be logged, to do this we use kubernetes to automatically spawn instances that will connect to twitch chats and log data.
Every pod can handle 50 connections, and every connection can log 50 channels each. Which is 2500 channels per pod.

For now we use https://github.com/mb-14/gomarkov to generate the markov chains, but we are looking into making our own.

How does this work

User sends request to /api/v1/markov

We check redis if there is data for the given channel
If there is data, we generate a markov chain from the data and return it

or we use grpc to send a join request to the chat service master, which will find a chat service worker that will join a channel.
If the worker is full, (Has 2500 connections) it will send back "FULL", and the master will spawn a new instance.

This instance is responsible for using redis Pub/Sub and send a single chat message to the master, the master is then responsible for adding that data to redis
Either we use just a single redis key and use a stringable array, or we use redis members

we could also opt into storing it in a better database, such as mongo but not sure.
