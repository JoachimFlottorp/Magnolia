import type { MarkovGenerateOptions } from 'markov-strings';

import { connect } from 'amqplib';

import { MarkovResponse, MarkovRequest } from './protobuf/markov';
import Markov from 'markov-strings';

interface Config {
	markov: {
		health_bind: number;
	};
	rmq: {
		uri: string;
	};
}

const argv = process.argv.slice(2);
const args = {
	config: argv[0],
};

if (!args.config) throw new Error('No config file specified');

const config = require(args.config) as Config;
const queue = 'markov-generator';

(async () => {
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
			markov = generateMarkov(data.uuid, data.messages, data.seed ?? '');
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
})();

const fromProto = (data: Buffer): MarkovRequest => MarkovRequest.decode(data);
const toProto = (data: MarkovResponse): Buffer => Buffer.from(MarkovResponse.encode(data).finish());

const generateMarkov = (uuid: string, data: string[], seed: string): string => {
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

(async (bind: number) => {
	const fastify = (await import('fastify')).default();
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
})(config.markov.health_bind);
