import * as toml from 'npm:@gulujs/toml';
import type { MarkovGenerateOptions } from 'npm:markov-strings';
import Markov from 'npm:markov-strings';

import * as amqp from 'https://deno.land/x/amqp@v0.21.0/mod.ts';
import { Server, Handler } from 'https://deno.land/std@0.170.0/http/server.ts';

import { MarkovRequest, MarkovResponse } from './protobuf/messages/proto/index.ts';

type Config = {
	markov: {
		health_bind: number;
	};
	rmq: {
		uri: string;
	};
};

const args = {
	config: Deno.args[0],
};

if (!args.config) throw new Error('No config file specified');

const queue = 'markov-generator';
const configContent = Deno.readTextFileSync(args.config);

(async () => {
	const config = toml.parse<Config>(configContent);

	const rmqConnection = await amqp.connect(config.rmq.uri);
	const rmqChannel = await rmqConnection.openChannel();

	await rmqChannel.declareQueue({ queue, durable: true });

	rmqChannel.consume({ queue }, async ({ deliveryTag }, { correlationId }, rawData) => {
		const data = await fromProto(rawData);
		console.log({ correlationId });

		rmqChannel.ack({ deliveryTag, multiple: false });

		let markov = '';
		let error: string | undefined = undefined;
		try {
			markov = await generateMarkov(data.messages, data.seed ?? '');
		} catch (e) {
			console.error('Error generating markov', e);

			error = e.message;
		}

		console.log({ markov, error });

		rmqChannel.publish(
			{ routingKey: queue },
			{ correlationId, contentType: 'application/protobuf' },
			await toProto({ result: markov, error }),
		);
	});

	console.log('Listening for markov requests at', { queue });

	startHealth(config.markov.health_bind);
})();

const fromProto = async (data: Uint8Array): Promise<MarkovRequest> => {
	const deserialize = (await import('./protobuf/messages/proto/MarkovRequest.ts')).decodeBinary;

	return deserialize(data);
};

const toProto = async (data: MarkovResponse): Promise<Uint8Array> => {
	const serialize = (await import('./protobuf/messages/proto/MarkovResponse.ts')).encodeBinary;

	return serialize(data);
};

const generateMarkov = (data: string[], seed: string): Promise<string> => {
	// Type 'string' is not assignable to type 'Promise<string>'.deno-ts(2322)
	if (!data.length) return '';

	const m = new Markov.default({ stateSize: 1 });

	m.addData(data);

	const options: MarkovGenerateOptions = {
		maxTries: 10000,
		prng: Math.random,
		filter: (r) =>
			r.score > 5 &&
			r.refs.filter((x) => x.string.includes(seed)).length > 0 &&
			r.string.split(' ').length >= 10,
	};

	// Type 'string' is not assignable to type 'Promise<string>'.deno-ts(2322)
	return m.generate(options).string;
};

async function startHealth(port: number) {
	const handler: Handler = () => {
		const body = JSON.stringify({ status: 'ok' });

		return new Response(body, { status: 200 });
	};

	const server = new Server({ port, handler });

	await server.listenAndServe();
}

export {};
