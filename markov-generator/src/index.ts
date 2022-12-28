import type { MarkovGenerateOptions } from 'markov-strings';

import { connect } from 'amqplib';

import { MarkovResponse, MarkovRequest } from './protobuf/markov.js';
import Markov from 'markov-strings';
import * as toml from '@gulujs/toml';
import fs from 'node:fs';

type Config = {
	markov: {
		health_bind: number;
	};
	rmq: {
		uri: string;
	};
};

const argv = process.argv.slice(2);
const args = {
	config: argv[0],
};

if (!args.config) throw new Error('No config file specified');

const queue = 'markov-generator';
const configContent = fs.readFileSync(args.config, 'utf-8');

(async () => {
	const config = toml.parse<Config>(configContent);

	const rmqconnection = await connect(config.rmq.uri);
	const rmqchannel = await rmqconnection.createChannel();

	await rmqchannel.assertQueue(queue, { durable: true });

	rmqchannel.consume(queue, async (msg) => {
		if (!msg) return;

		const data = fromProto(msg.content);
		const id = msg.properties.correlationId;
		console.log({ id });

		rmqchannel.ack(msg);

		let markov = '';
		let error: string | undefined = undefined;
		try {
			markov = generateMarkov(data.messages, data.seed ?? '');
		} catch (e) {
			console.error('Error generating markov', e);

			error = e.message;
		}

		console.log({ markov, error });

		rmqchannel.sendToQueue(queue, toProto({ result: markov, error }), {
			correlationId: id,
		});
	});

	console.log('Listening for markov requests at', { queue });

	startHealth(config.markov.health_bind);
})();

const fromProto = (data: Buffer): MarkovRequest => MarkovRequest.decode(data);
const toProto = (data: MarkovResponse): Buffer => Buffer.from(MarkovResponse.encode(data).finish());

const generateMarkov = (data: string[], seed: string): string => {
	if (!data.length) return '';

	const m = new Markov({ stateSize: 1 });

	m.addData(data);

	const options: MarkovGenerateOptions = {
		maxTries: 10000,
		prng: Math.random,
		filter: (r) =>
			r.score > 5 &&
			r.refs.filter((x) => x.string.includes(seed)).length > 0 &&
			r.string.split(' ').length >= 10,
	};

	return m.generate(options).string;
};

import fastifyConstructor from 'fastify';
const fastify = fastifyConstructor();
const startHealth = async (bind: number) => {
	fastify.get('/health', (req, reply) => {
		reply.send({ status: 'ok' });
	});

	fastify.listen({ port: bind, host: '0.0.0.0' }, (e, a) => {
		if (e) {
			console.error(e);
			process.exit(1);
		}

		console.log(`Fastify listening on ${a}`);
	});
};

export {};
